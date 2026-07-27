package chart

import (
	"bytes"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var write = flag.Bool("write", false, "ゴールデンファイルとプレビューHTMLを書き直す")

// goldenDir はゴールデンSVGの置き場。テストの期待値なのでコミットする。
const goldenDir = "testdata/golden"

// previewHTML は目視確認用の生成物。テストの入力でも期待値でもないので gitignore 済み。
const previewHTML = "testdata/preview.html"

// renderCase は 1 パターン分の描画条件。ゴールデン比較とプレビューで共有する。
type renderCase struct {
	name string
	opt  Options
}

func renderCases() []renderCase {
	return []renderCase{
		{"default", fixtureOptions()},
		{"bars-only", func() Options {
			o := fixtureOptions()
			o.Line = Series{}
			return o
		}()},
		{"narrow", func() Options {
			o := fixtureOptions()
			o.Width, o.Height = 720, 200
			return o
		}()},
		{"partial-line", func() Options {
			// PV が末尾 7 日分しか無い場合。Offset とハイライト帯の確認用。
			o := fixtureOptions()
			o.Line.Values = fixturePVTail()
			o.Line.Offset = fixturePVOffset()
			return o
		}()},
		{"seven-days", func() Options {
			o := fixtureOptions()
			o.Bars.Values = fixtureCost()[fixturePVOffset():]
			o.Line.Values = fixturePVTail()
			o.XLabels = []XLabel{{At: 0, Text: "Jul 19"}, {At: 3, Text: "Jul 22"}, {At: 6, Text: "Jul 25"}}
			return o
		}()},
	}
}

// --- ゴールデン比較 ---
//
// 出力が 1 バイトでも変わったら落ちる。リファクタで座標が動いていないことを
// 自動で確認するための砦。意図して見た目を変えたときは
//
//	go test ./internal/chart -write
//
// で testdata/golden/ を更新し、差分を目で見てからコミットする。
func TestGoldenSVG(t *testing.T) {
	if err := os.MkdirAll(goldenDir, 0o755); err != nil {
		t.Fatal(err)
	}

	for _, c := range renderCases() {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := RenderSVG(&buf, c.opt); err != nil {
				t.Fatalf("RenderSVG: %v", err)
			}

			path := filepath.Join(goldenDir, c.name+".svg")

			if *write {
				if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
					t.Fatal(err)
				}
				t.Logf("更新: %s (%d バイト)", path, buf.Len())
				return
			}

			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ゴールデンが読めない（-write で作成する）: %v", err)
			}
			if !bytes.Equal(want, buf.Bytes()) {
				t.Errorf("出力が %s と一致しない。意図した変更なら -write で更新する\n%s", path, firstDiff(string(want), buf.String()))
			}
		})
	}
}

// firstDiff は最初に食い違った行を返す。差分全体を出すと長すぎるため。
func firstDiff(want, got string) string {
	wl, gl := strings.Split(want, "\n"), strings.Split(got, "\n")
	for i := 0; i < len(wl) && i < len(gl); i++ {
		if wl[i] != gl[i] {
			return fmt.Sprintf("  行 %d:\n    want: %.120s\n    got : %.120s", i+1, wl[i], gl[i])
		}
	}
	return fmt.Sprintf("  行数が違う: want %d 行, got %d 行", len(wl), len(gl))
}

// --- 目視確認用 ---
//
//	go test ./internal/chart -run Preview -write
//
// で testdata/preview.html が出るのでブラウザで開く。ライト／ダーク両方の
// 背景に並べて表示されるので、配色の確認もここでできる。
// fixture は testdata_test.go にあるので、調整したい形はそこをいじる。
func TestPreview(t *testing.T) {
	if !*write {
		t.Skip("-write が無いのでスキップ（書き出すには: go test ./internal/chart -run Preview -write）")
	}

	var html strings.Builder
	html.WriteString("<!doctype html><meta charset=\"utf-8\"><title>pvvc chart preview</title>\n")
	html.WriteString("<style>body{margin:0;font:13px system-ui}section{padding:24px}h2{font-size:12px;font-weight:600;letter-spacing:.04em;text-transform:uppercase;opacity:.5;margin:0 0 12px}svg{max-width:100%;height:auto}.light{background:#fff;color:#000}.dark{background:#0a0a0a;color:#fff}</style>\n")

	for _, c := range renderCases() {
		var buf bytes.Buffer
		if err := RenderSVG(&buf, c.opt); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		for _, theme := range []string{"light", "dark"} {
			fmt.Fprintf(&html, "<section class=%q><h2>%s / %s</h2>%s</section>\n", theme, c.name, theme, buf.String())
		}
	}

	if err := os.WriteFile(previewHTML, []byte(html.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote %s -- open it in a browser", previewHTML)
}

// --- 構造の検証 ---

func TestRenderSVGWellFormed(t *testing.T) {
	var buf bytes.Buffer
	if err := RenderSVG(&buf, fixtureOptions()); err != nil {
		t.Fatal(err)
	}

	dec := xml.NewDecoder(&buf)
	for {
		_, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("XML として壊れている: %v", err)
		}
	}
}

// バーがプロット領域を上下にはみ出していないこと。
// 軸上端が dataMax を下回るとバーの y が負になり、viewBox の外で頭が欠ける。
func TestRenderSVGBarsWithinPlot(t *testing.T) {
	opt := fixtureOptions()
	opt.applyDefaults() // plotH などを引くためテスト側でも解決しておく
	doc := render(t, opt)

	plotH := opt.plotH()
	// theme.go の値を直に見る。リテラルを書くと色や濃度を変えた瞬間に見失う。
	bars := doc.find(func(g group) bool { return g.Opacity == barOpacity })
	if bars == nil {
		t.Fatal("バーのグループが見つからない")
	}
	if len(bars.Rects) != len(opt.Bars.Values) {
		t.Fatalf("バーの本数が %d、期待は %d", len(bars.Rects), len(opt.Bars.Values))
	}

	for i, r := range bars.Rects {
		if r.Y < 0 {
			t.Errorf("bar[%d]: y=%v が負（プロット領域の上にはみ出している）", i, r.Y)
		}
		if got := r.Y + r.Height; got > plotH+0.01 {
			t.Errorf("bar[%d]: y+height=%v が plotH=%v を超えている", i, got, plotH)
		}
		if r.Width <= 0 || r.Height < 0 {
			t.Errorf("bar[%d]: 幅/高さが不正 w=%v h=%v", i, r.Width, r.Height)
		}
	}
}

// 左右の軸でグリッド線の Y 座標が一致すること（Ticks を共有しているので重なる）。
func TestRenderSVGGridLinesAlign(t *testing.T) {
	opt := fixtureOptions()
	opt.applyDefaults()
	doc := render(t, opt)

	var (
		plotW = opt.plotW()
		left  []float64
		right []float64
	)
	doc.walk(func(g group) {
		for _, l := range g.Lines {
			switch {
			case l.X1 == 0 && l.X2 == plotW && l.StrokeWidth == "1":
				left = append(left, l.Y1)
			case l.X1 == plotW && l.X2 == plotW+5:
				right = append(right, l.Y1)
			}
		}
	})

	if len(left) != opt.Ticks+1 {
		t.Fatalf("左軸のグリッド線が %d 本、期待は %d 本", len(left), opt.Ticks+1)
	}
	if len(right) != len(left) {
		t.Fatalf("右軸の目盛りが %d 個、左軸は %d 個", len(right), len(left))
	}
	for i := range left {
		if math.Abs(left[i]-right[i]) > 0.01 {
			t.Errorf("tick[%d]: 左 y=%v と右 y=%v がずれている", i, left[i], right[i])
		}
	}
}

// 折れ線が背景ハイライトの範囲内に収まっていること。
func TestRenderSVGLineWithinHighlight(t *testing.T) {
	// PV が一部区間だけのケース（本番の ChartOptions は全日分を渡すので使わない経路）
	opt := fixtureOptions()
	opt.Line.Values = fixturePVTail()
	opt.Line.Offset = fixturePVOffset()
	opt.applyDefaults()
	doc := render(t, opt)

	if len(doc.Polylines) != 1 {
		t.Fatalf("polyline が %d 本、期待は 1 本", len(doc.Polylines))
	}
	pts := parsePoints(t, doc.Polylines[0].Points)
	if len(pts) != len(opt.Line.Values) {
		t.Fatalf("折れ線の点が %d 個、期待は %d 個", len(pts), len(opt.Line.Values))
	}

	var (
		plotW = opt.plotW()
		plotH = opt.plotH()
		from  = float64(opt.Line.Offset) * opt.slot()
	)
	for i, p := range pts {
		if p[0] < from || p[0] > plotW {
			t.Errorf("point[%d]: x=%v がハイライト範囲 [%v, %v] の外", i, p[0], from, plotW)
		}
		if p[1] < 0 || p[1] > plotH {
			t.Errorf("point[%d]: y=%v がプロット領域 [0, %v] の外", i, p[1], plotH)
		}
	}
}

// 最大値の注記が正しい値・正しい位置に付くこと。
func TestRenderSVGAnnotatesLineMax(t *testing.T) {
	var buf bytes.Buffer
	if err := RenderSVG(&buf, fixtureOptions()); err != nil {
		t.Fatal(err)
	}
	if want := Comma(fixturePVPeak); !strings.Contains(buf.String(), ">"+want+"<") {
		t.Errorf("最大値ラベル %q が出力に無い", want)
	}
}

func TestRenderSVGLineOmittedWhenEmpty(t *testing.T) {
	opt := fixtureOptions()
	opt.Line = Series{}
	opt.AriaLabel = "Vercel daily cost" // aria-label にも系列名が載るので外しておく

	var buf bytes.Buffer
	if err := RenderSVG(&buf, opt); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	// "PV" はホバー文言にも含まれるので、右軸の見出し要素そのものを見る
	for _, s := range []string{"polyline", ">PV</text>", "<circle"} {
		if strings.Contains(out, s) {
			t.Errorf("Line が空なのに %q が出力されている", s)
		}
	}
}

// ラベルに < & " が混ざっても壊れないこと。
func TestRenderSVGEscapesLabels(t *testing.T) {
	opt := fixtureOptions()
	opt.Bars.Name = `cost <USD> & "JPY"`
	opt.AriaLabel = "a & b"
	opt.XLabels = []XLabel{{At: 0, Text: "<script>"}}

	var buf bytes.Buffer
	if err := RenderSVG(&buf, opt); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "<script>") {
		t.Error("ラベルがエスケープされていない")
	}

	dec := xml.NewDecoder(bytes.NewReader(buf.Bytes()))
	for {
		_, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("エスケープ後に XML が壊れた: %v", err)
		}
	}
}

func TestRenderSVGEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		opt     Options
		wantErr bool
	}{
		{"データなし", Options{}, true},
		{"1点だけ", Options{Bars: Series{Values: []float64{42}}}, false},
		{"全部ゼロ", Options{Bars: Series{Values: []float64{0, 0, 0}}}, false},
		{"NaN 混じり", Options{Bars: Series{Values: []float64{1, math.NaN(), 3}}}, false},
		{"+Inf 混じり", Options{Bars: Series{Values: []float64{1, math.Inf(1), 3}}}, false},
		{"-Inf 混じり", Options{Bars: Series{Values: []float64{1, math.Inf(-1), 3}}}, false},
		{"負の値", Options{Bars: Series{Values: []float64{-50, 10, 20}}}, false},
		{"折れ線に Inf", Options{
			Bars: Series{Values: []float64{1, 2, 3}},
			Line: Series{Name: "pv", Values: []float64{1, math.Inf(1), 3}},
		}, false},
		{"折れ線だけ", Options{Line: Series{Name: "pv", Values: []float64{1, 2, 3}}}, false},
		{"極小の値", Options{Bars: Series{Values: []float64{0.0001, 0.0003}}}, false},
		{"極大の値", Options{Bars: Series{Values: []float64{1e12, 3e12}}}, false},
		{"ラベルが範囲外", Options{
			Bars:    Series{Values: []float64{1, 2}},
			XLabels: []XLabel{{At: 99, Text: "out"}},
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := RenderSVG(&buf, tt.opt)

			if tt.wantErr {
				if err == nil {
					t.Fatal("エラーを期待したが nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if strings.Contains(buf.String(), "NaN") || strings.Contains(buf.String(), "Inf") {
				t.Error("出力に NaN / Inf が混入している")
			}
			// 幅・高さが負の図形は不正な SVG になる
			if i := strings.Index(buf.String(), `height="-`); i >= 0 {
				t.Errorf("height が負になっている: %.60s", buf.String()[i:])
			}
			if i := strings.Index(buf.String(), `width="-`); i >= 0 {
				t.Errorf("width が負になっている: %.60s", buf.String()[i:])
			}

			dec := xml.NewDecoder(bytes.NewReader(buf.Bytes()))
			for {
				if _, err := dec.Token(); errors.Is(err, io.EOF) {
					break
				} else if err != nil {
					t.Fatalf("XML が壊れている: %v", err)
				}
			}
		})
	}
}

// 同じ入力なら常に同じ出力（fixture が決定的であることの確認）。
func TestRenderSVGDeterministic(t *testing.T) {
	var a, b bytes.Buffer
	if err := RenderSVG(&a, fixtureOptions()); err != nil {
		t.Fatal(err)
	}
	if err := RenderSVG(&b, fixtureOptions()); err != nil {
		t.Fatal(err)
	}
	if a.String() != b.String() {
		t.Error("出力が実行ごとに変わっている")
	}
}

// --- スケール ---

func TestNewScale(t *testing.T) {
	tests := []struct {
		dataMax  float64
		ticks    int
		wantStep float64
		wantMax  float64
	}{
		{331, 5, 75, 375},
		{859_241, 5, 200_000, 1_000_000},
		{5, 5, 1, 5},
		{1, 5, 0.2, 1},
		{0, 5, 0.2, 1}, // データが空/ゼロでもゼロ除算しない
		{12, 6, 2, 12},
	}

	for _, tt := range tests {
		t.Run(Plain(tt.dataMax), func(t *testing.T) {
			s := NewScale(tt.dataMax, tt.ticks, 200)
			if math.Abs(s.Step-tt.wantStep) > 1e-9 {
				t.Errorf("Step = %v, want %v", s.Step, tt.wantStep)
			}
			if math.Abs(s.Max-tt.wantMax) > 1e-9 {
				t.Errorf("Max = %v, want %v", s.Max, tt.wantMax)
			}
			if len(s.TickValues()) != tt.ticks+1 {
				t.Errorf("目盛りが %d 個、期待は %d 個", len(s.TickValues()), tt.ticks+1)
			}
		})
	}
}

// 軸上端は必ずデータ最大値以上（バーがはみ出さないことの根拠）。
func TestNewScaleNeverClipsData(t *testing.T) {
	for _, v := range []float64{0.7, 1, 9.99, 10, 99, 100, 101, 249, 250, 251, 329, 331, 1001, 859_241, 1e9} {
		for _, ticks := range []int{4, 5, 6, 8} {
			if s := NewScale(v, ticks, 200); s.Max < v {
				t.Errorf("NewScale(%v, %d): Max=%v < dataMax", v, ticks, s.Max)
			}
		}
	}
}

func TestScaleY(t *testing.T) {
	s := NewScale(100, 5, 200) // Max = 100, height = 200

	tests := []struct{ v, want float64 }{
		{0, 200},  // ゼロは底
		{100, 0},  // 最大値は天井
		{50, 100}, // 中間
	}
	for _, tt := range tests {
		if got := s.Y(tt.v); math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("Y(%v) = %v, want %v", tt.v, got, tt.want)
		}
	}
}

// --- フォーマッタ ---

func TestFormatters(t *testing.T) {
	tests := []struct {
		fn   func(float64) string
		in   float64
		want string
	}{
		{USD, 0, "$0"},
		{USD, 130, "$130"},
		{Compact, 0, "0"},
		{Compact, 999, "999"},
		{Compact, 1_000, "1k"},
		{Compact, 200_000, "200k"},
		{Compact, 1_000_000, "1M"},
		{Compact, 1_500_000, "1.5M"},
		{Comma, 0, "0"},
		{Comma, 999, "999"},
		{Comma, 1_000, "1,000"},
		{Comma, 859_241, "859,241"},
		{Comma, -1_234, "-1,234"},
	}

	for _, tt := range tests {
		if got := tt.fn(tt.in); got != tt.want {
			t.Errorf("f(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// --- SVG をパースするヘルパー ---

type rect struct {
	X      float64 `xml:"x,attr"`
	Y      float64 `xml:"y,attr"`
	Width  float64 `xml:"width,attr"`
	Height float64 `xml:"height,attr"`
	Fill   string  `xml:"fill,attr"`
}

type line struct {
	X1          float64 `xml:"x1,attr"`
	Y1          float64 `xml:"y1,attr"`
	X2          float64 `xml:"x2,attr"`
	Y2          float64 `xml:"y2,attr"`
	StrokeWidth string  `xml:"stroke-width,attr"`
}

type polyline struct {
	Points string `xml:"points,attr"`
}

type text struct {
	X       float64 `xml:"x,attr"`
	Y       float64 `xml:"y,attr"`
	Anchor  string  `xml:"text-anchor,attr"`
	Content string  `xml:",chardata"`
}

type group struct {
	Fill      string     `xml:"fill,attr"`
	Opacity   string     `xml:"opacity,attr"`
	Transform string     `xml:"transform,attr"`
	Rects     []rect     `xml:"rect"`
	Lines     []line     `xml:"line"`
	Texts     []text     `xml:"text"`
	Polylines []polyline `xml:"polyline"`
	Groups    []group    `xml:"g"`
}

type svgDoc struct {
	AriaLabel string     `xml:"aria-label,attr"`
	Width     float64    `xml:"width,attr"`
	Height    float64    `xml:"height,attr"`
	Groups    []group    `xml:"g"`
	Polylines []polyline `xml:"polyline"`
}

func render(t *testing.T, opt Options) *svgDoc {
	t.Helper()

	var buf bytes.Buffer
	if err := RenderSVG(&buf, opt); err != nil {
		t.Fatalf("RenderSVG: %v", err)
	}
	return parse(t, buf.Bytes())
}

// parse は出力済みの SVG を読み込む。
func parse(t *testing.T, b []byte) *svgDoc {
	t.Helper()

	var doc svgDoc
	if err := xml.Unmarshal(b, &doc); err != nil {
		t.Fatalf("SVG のパースに失敗: %v", err)
	}

	// polyline はプロット用グループの中にあるので、探しやすいよう root に集める。
	doc.walk(func(g group) { doc.Polylines = append(doc.Polylines, g.Polylines...) })
	return &doc
}

// walk は入れ子のグループを深さ優先で全て訪問する。
func (d *svgDoc) walk(fn func(group)) {
	var rec func([]group)
	rec = func(gs []group) {
		for _, g := range gs {
			fn(g)
			rec(g.Groups)
		}
	}
	rec(d.Groups)
}

// find は条件に合う最初のグループを返す。無ければ nil。
func (d *svgDoc) find(pred func(group) bool) *group {
	var found *group
	d.walk(func(g group) {
		if found == nil && pred(g) {
			hit := g
			found = &hit
		}
	})
	return found
}

func parsePoints(t *testing.T, s string) [][2]float64 {
	t.Helper()

	var out [][2]float64
	for p := range strings.FieldsSeq(s) {
		var x, y float64
		if _, err := fmt.Sscanf(p, "%g,%g", &x, &y); err != nil {
			t.Fatalf("points のパースに失敗 (%q): %v", p, err)
		}
		out = append(out, [2]float64{x, y})
	}
	return out
}

// FontFamily は公開フィールドなので、二重引用符で属性から抜け出せてはいけない。
func TestRenderSVGEscapesFontFamily(t *testing.T) {
	opt := fixtureOptions()
	opt.FontFamily = `Foo" onload="alert(1)`

	var buf bytes.Buffer
	if err := RenderSVG(&buf, opt); err != nil {
		t.Fatal(err)
	}
	// 生の " が残っていると属性が途切れて別の属性として解釈される。
	// エスケープ後は onload=&quot; になるので、これは単なる文字列として無害。
	if strings.Contains(buf.String(), `onload="`) {
		t.Errorf("属性から抜け出せている: %.160s", buf.String())
	}
	if !strings.Contains(buf.String(), "&quot;") {
		t.Error("二重引用符がエスケープされていない")
	}

	// 既定のフォントスタックのシングルクォートはそのまま残ってほしい（可読性）
	var def bytes.Buffer
	if err := RenderSVG(&def, fixtureOptions()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(def.String(), `'Segoe UI'`) {
		t.Error("フォントスタックのシングルクォートが壊れている")
	}
}

// ハイライト帯は折れ線が占める区間だけを覆うこと。
// 右端まで伸ばす実装だと Offset=0 のときに全幅が塗られてしまう。
func TestRenderSVGHighlightMatchesLineExtent(t *testing.T) {
	for _, offset := range []int{0, 40, 85} {
		opt := fixtureOptions()
		opt.Line.Values = fixturePVTail()
		opt.Line.Offset = offset

		var buf bytes.Buffer
		if err := RenderSVG(&buf, opt); err != nil {
			t.Fatal(err)
		}
		// plotH / slot は既定値を入れてからでないと 0 になる
		opt.applyDefaults()
		var (
			plotH = opt.plotH()
			slot  = opt.slot()
			wantX = float64(offset) * slot
			wantW = float64(len(opt.Line.Values)) * slot
		)

		// 当たり判定の矩形も y=0・全高なので、帯は自身の fill で特定する
		var hl *rect
		parse(t, buf.Bytes()).walk(func(g group) {
			for i, r := range g.Rects {
				if r.Fill == opt.Line.Color && r.Y == 0 && math.Abs(r.Height-plotH) < 0.01 {
					hl = &g.Rects[i]
				}
			}
		})
		if hl == nil {
			t.Fatalf("offset=%d: ハイライト帯が見つからない", offset)
		}
		if math.Abs(hl.X-wantX) > 0.01 {
			t.Errorf("offset=%d: 帯の x=%v, want %v", offset, hl.X, wantX)
		}
		if math.Abs(hl.Width-wantW) > 0.01 {
			t.Errorf("offset=%d: 帯の幅=%v, want %v（折れ線 %d 点分）", offset, hl.Width, wantW, len(opt.Line.Values))
		}
	}
}

// ホバー用の当たり判定が列ごとに出て、<title> が入っていること。
func TestRenderSVGHitAreas(t *testing.T) {
	opt := fixtureOptions()

	var buf bytes.Buffer
	if err := RenderSVG(&buf, opt); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	if got, want := strings.Count(out, "<title>"), len(opt.SlotTitles); got != want {
		t.Errorf("<title> が %d 個、期待は %d 個（1列1件）", got, want)
	}
	if !strings.Contains(out, `pointer-events="all"`) {
		t.Error("当たり判定に pointer-events が付いていない")
	}

	// 当たり判定は最前面でないとバーや折れ線に取られる
	if strings.Index(out, `pointer-events="all"`) < strings.LastIndex(out, "<polyline") {
		t.Error("当たり判定が折れ線より前に出力されている（最前面にならない）")
	}

	// 列は全高・スロット幅ぴったりを覆う
	opt.applyDefaults()
	var hits []rect
	parse(t, buf.Bytes()).walk(func(g group) {
		if g.Fill == "transparent" {
			hits = append(hits, g.Rects...)
		}
	})
	if len(hits) != len(opt.SlotTitles) {
		t.Fatalf("当たり判定が %d 個、期待は %d 個", len(hits), len(opt.SlotTitles))
	}
	for i, r := range hits {
		if math.Abs(r.Width-opt.slot()) > 0.01 {
			t.Errorf("hit[%d]: 幅=%v, want %v", i, r.Width, opt.slot())
		}
		if math.Abs(r.Height-opt.plotH()) > 0.01 {
			t.Errorf("hit[%d]: 高さ=%v, want %v", i, r.Height, opt.plotH())
		}
	}
}

// SlotTitles が空なら当たり判定を出さないこと（余計な要素を増やさない）。
func TestRenderSVGNoHitAreasWithoutTitles(t *testing.T) {
	opt := fixtureOptions()
	opt.SlotTitles = nil

	var buf bytes.Buffer
	if err := RenderSVG(&buf, opt); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"<title>", "pointer-events"} {
		if strings.Contains(buf.String(), s) {
			t.Errorf("SlotTitles が空なのに %q が出ている", s)
		}
	}
}

// ホバー文言に <  & が混ざっても壊れないこと。
func TestRenderSVGEscapesSlotTitles(t *testing.T) {
	opt := fixtureOptions()
	opt.SlotTitles = []string{`<script>alert(1)</script>`, "a & b"}

	var buf bytes.Buffer
	if err := RenderSVG(&buf, opt); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "<script>") {
		t.Error("ホバー文言がエスケープされていない")
	}

	dec := xml.NewDecoder(bytes.NewReader(buf.Bytes()))
	for {
		if _, err := dec.Token(); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			t.Fatalf("XML が壊れている: %v", err)
		}
	}
}
