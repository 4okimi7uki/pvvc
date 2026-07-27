package chart

import (
	"fmt"
	"math"
)

// このパッケージのテストは外部APIを一切叩かない。
// SVG の見た目を調整したいときは fixture をいじって -write で回す。
// testdata/golden/ が更新され、目視用の testdata/preview.html も出る。
//
//	go test ./internal/chart -write

const (
	fixtureDays    = 92 // 棒グラフの日数（約3ヶ月）
	fixturePVDays  = 7  // PV が一部区間だけの場合に使う末尾の日数
	fixtureCostMin = 95.0
	fixtureCostMax = 331.0
)

// fixtureCost は 92 日分の日次コスト(USD)。
// 週次の波 + 緩やかな増加トレンド + 小さなノイズを重ねた決定的な値。
// 乱数を使わないので、出力SVGは実行ごとに完全に一致する。
func fixtureCost() []float64 {
	out := make([]float64, fixtureDays)
	for i := range out {
		var (
			t      = float64(i) / float64(fixtureDays-1)
			trend  = 0.55 + 0.45*t                             // 右肩上がり
			weekly = 1 + 0.18*math.Sin(2*math.Pi*float64(i)/7) // 週次の波
			noise  = 1 + 0.06*math.Sin(float64(i)*2.399963)    // 黄金角で疑似ノイズ
			v      = fixtureCostMax * trend * weekly * noise
		)
		out[i] = math.Max(fixtureCostMin, math.Min(fixtureCostMax, v))
	}
	return out
}

// fixturePV は棒グラフと同じ日数分の PV。本番の ChartOptions は
// 全日分の PV を渡すので、これが基準の形。fixturePVPeak を最大値として含む。
func fixturePV() []float64 {
	out := make([]float64, fixtureDays)
	for i := range out {
		// ゆるい波 + 後半に向かってわずかに増加
		out[i] = 620_000 + 120_000*math.Sin(float64(i)/11) + 900*float64(i)
	}
	out[fixtureDays-2] = fixturePVPeak // 最大値ラベルの位置を固定する
	return out
}

const fixturePVPeak = 859_241.0

// fixturePVTail は末尾 7 日分だけの PV。
// Series.Offset とハイライト帯（PV が一部区間だけの場合）を確認するために使う。
func fixturePVTail() []float64 {
	return []float64{770_400, 749_100, 742_800, 763_500, 845_000, fixturePVPeak, 717_200}
}

// fixturePVOffset は fixturePVTail()[0] が X 軸の何番目に対応するか。
func fixturePVOffset() int { return fixtureDays - fixturePVDays }

// fixtureXLabels は 12 日おきの日付ラベル。実データでは report 側で組む想定。
func fixtureXLabels() []XLabel {
	months := []string{"May 2", "May 14", "May 26", "Jun 7", "Jun 19", "Jul 1", "Jul 13", "Jul 25"}
	out := make([]XLabel, 0, len(months))
	for i, m := range months {
		out = append(out, XLabel{At: min(i*12, fixtureDays-1), Text: m})
	}
	return out
}

// fixtureSlotTitles は列ごとのホバー文言。本番と同じく 1 日 1 件。
func fixtureSlotTitles() []string {
	cost, pv := fixtureCost(), fixturePV()
	out := make([]string, fixtureDays)
	for i := range out {
		out[i] = fmt.Sprintf("day %d\n$%.2f  ·  %.0f PV", i+1, cost[i], pv[i])
	}
	return out
}

// fixtureOptions は「本番と同じ形」の Options。テストの基準ケース。
func fixtureOptions() Options {
	return Options{
		AriaLabel: "Vercel daily cost with GA4 pageviews overlaid",
		// 生成時刻は固定値。ここに time.Now() を混ぜるとゴールデンが毎回変わる。
		Caption:         "2026-04-26 → 2026-07-26 (92 days)  ·  generated 2026-07-27 10:00 JST",
		Ticks:           5,
		AnnotateLineMax: true,
		Bars: Series{
			Name:       "Vercel cost / day (USD)",
			Values:     fixtureCost(),
			FormatTick: USD,
		},
		Line: Series{
			Name:       "GA4 pageviews",
			Values:     fixturePV(),
			FormatTick: Compact,
		},
		XLabels:    fixtureXLabels(),
		SlotTitles: fixtureSlotTitles(),
	}
}
