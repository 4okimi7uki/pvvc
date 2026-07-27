package chart

import "math"

// Series は 1 本の系列。Values は X 軸のインデックス順に並ぶ。
// Offset は Values[0] が X 軸の何番目に対応するかを表す（末尾 7 日だけの
// 折れ線を全期間の棒グラフに重ねる、といった用途）。
type Series struct {
	Name   string
	Color  string
	Values []float64
	Offset int

	// FormatTick は軸の目盛りラベルの整形。nil なら Plain。
	FormatTick func(float64) string
}

// max は系列の最大値。NaN と ±Inf は無視する。
// Inf を通すと Scale.Max が Inf になり、全座標が NaN になって出力が壊れる。
func (s Series) max() float64 {
	m := 0.0
	for _, v := range s.Values {
		if finite(v) && v > m {
			m = v
		}
	}
	return m
}

// finite は描画可能な値か。NaN / ±Inf は欠損として扱う。
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

// plotValue は描画に使う値。この二軸チャートは非負の系列（コスト・PV）専用なので、
// 負の値は 0 として扱う。そのまま通すと height が負の不正な SVG になる。
func plotValue(v float64) float64 {
	if v < 0 {
		return 0
	}
	return v
}

func (s Series) tick(v float64) string {
	if s.FormatTick == nil {
		return Plain(v)
	}
	return s.FormatTick(v)
}

// XLabel は X 軸の目盛り。At は X 軸のインデックス。
type XLabel struct {
	At   int
	Text string
}

type Padding struct {
	Top, Right, Bottom, Left float64
}

type Options struct {
	Width, Height float64
	Pad           Padding

	// Bars は左軸に紐づく棒グラフ。X 軸の全区間を占める系列を想定。
	Bars Series
	// Line は右軸に紐づく折れ線。Values が空なら描画しない。
	Line Series

	XLabels []XLabel

	// Caption は図の下に右寄せで置く補足。対象期間や生成時刻を入れる想定。
	// 生成時刻のような可変値を chart 側で作らないのは、出力を決定的に保つため。
	Caption string

	// SlotTitles は X 軸 1 コマごとのホバー文言（<title> の中身）。
	// インライン表示か SVG を直接開いたときだけ効く。<img> 経由では無反応。
	// 空なら当たり判定そのものを出力しない。
	SlotTitles []string

	// Ticks は横グリッド線の区間数。左右の軸で共有するのでグリッドが重なる。
	Ticks int

	// AnnotateLineMax を true にすると折れ線の最大値に数値ラベルを付ける。
	AnnotateLineMax bool

	AriaLabel  string
	FontFamily string

	// 以下は見た目の微調整。ゼロ値でデフォルトが入る（theme.go 参照）。
	GridColor string
	TextColor string
	BarRatio  float64 // スロット幅に対する棒の太さ
}

// --- 座標系 ---
//
// いずれも applyDefaults 済みの Options で呼ぶこと。
// 未解決だと Width/Pad がゼロ値のまま計算されて 0 が返る。

// plotW / plotH は軸の内側の描画領域。
func (o Options) plotW() float64 { return o.Width - o.Pad.Left - o.Pad.Right }
func (o Options) plotH() float64 { return o.Height - o.Pad.Top - o.Pad.Bottom }

// slot は X 軸 1 コマ分の幅。
func (o Options) slot() float64 { return o.plotW() / float64(o.domain()) }

// centerX は X 軸インデックス i のスロット中心。
func (o Options) centerX(i int) float64 { return float64(i)*o.slot() + o.slot()/2 }

// domain は X 軸のスロット数。棒グラフと折れ線の届く範囲の広い方。
func (o Options) domain() int {
	n := len(o.Bars.Values)
	if end := o.Line.Offset + len(o.Line.Values); end > n {
		n = end
	}
	return n
}
