package chart

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// RenderSVG は棒グラフ＋折れ線の二軸チャートを SVG として書き出す。
// 型は model.go、色や寸法の既定値は theme.go、軸の目盛り計算は scale.go にある。
func RenderSVG(w io.Writer, opt Options) error {
	opt.applyDefaults()

	n := opt.domain()
	if n == 0 {
		return fmt.Errorf("chart: no data points")
	}

	var (
		plotW = opt.plotW()
		plotH = opt.plotH()
		slot  = opt.slot()
		barW  = slot * opt.BarRatio

		left  = NewScale(opt.Bars.max(), opt.Ticks, plotH)
		right = NewScale(opt.Line.max(), opt.Ticks, plotH)
	)

	var b strings.Builder
	b.Grow(4096)

	fmt.Fprintf(&b,
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %s %s" width="%s" height="%s" font-family="%s" role="img" aria-label="%s">`,
		num(opt.Width), num(opt.Height), num(opt.Width), num(opt.Height),
		escAttr(opt.FontFamily), esc(opt.AriaLabel),
	)

	writeLegend(&b, opt)

	fmt.Fprintf(&b, "\n<g transform=\"translate(%s,%s)\">", num(opt.Pad.Left), num(opt.Pad.Top))

	// 折れ線が一部区間だけの場合、その範囲に薄い背景を敷いて対象期間を示す。
	// 幅は折れ線が占めるスロット数から出す（右端まで伸ばすと Offset=0 のときに全幅になる）。
	if len(opt.Line.Values) > 0 && len(opt.Line.Values) < n {
		var (
			x = float64(opt.Line.Offset) * slot
			w = float64(len(opt.Line.Values)) * slot
		)
		fmt.Fprintf(&b, "\n<rect x=\"%s\" y=\"0\" width=\"%s\" height=\"%s\" fill=\"%s\" opacity=\"%s\"/>",
			num(x), num(min(w, plotW-x)), num(plotH), opt.Line.Color, highlightOpacity)
	}

	writeLeftAxis(&b, opt, left, plotW)
	if len(opt.Line.Values) > 0 {
		writeRightAxis(&b, opt, right, plotW)
	}

	writeBars(&b, opt, left, slot, barW, plotH)
	if len(opt.Line.Values) > 0 {
		writeLine(&b, opt, right)
	}

	writeXAxis(&b, opt, plotW, plotH)
	writeHitAreas(&b, opt, plotH, slot)

	b.WriteString("</g></svg>\n")

	_, err := io.WriteString(w, b.String())
	return err
}

func writeLegend(b *strings.Builder, opt Options) {
	// TODO: ラベル幅に応じて折れ線凡例の x 座標を決めたい。今は固定オフセット。
	const swatch = 162.0

	fmt.Fprintf(b, "\n<g transform=\"translate(%s,16)\" font-size=\"11\" fill=\"%s\">", num(opt.Pad.Left), opt.TextColor)
	fmt.Fprintf(b, `<rect x="0" y="-7" width="9" height="9" rx="1.5" fill="%s" opacity="%s"/><text x="15" y="1">%s</text>`,
		opt.Bars.Color, swatchOpacity, esc(opt.Bars.Name))

	if len(opt.Line.Values) > 0 {
		fmt.Fprintf(b,
			`<line x1="%s" y1="-3" x2="%s" y2="-3" stroke="%s" stroke-width="2"/><circle cx="%s" cy="-3" r="2.8" fill="%s"/><text x="%s" y="1">%s</text>`,
			num(swatch), num(swatch+22), opt.Line.Color,
			num(swatch+11), opt.Line.Color,
			num(swatch+28), esc(opt.Line.Name),
		)
	}
	b.WriteString("</g>")
}

func writeLeftAxis(b *strings.Builder, opt Options, s Scale, plotW float64) {
	// 目盛りラベルは棒の色に合わせる（右軸が折れ線の色なのと対になる）。
	// グリッド線は line 側で GridColor を明示しているのでこの fill の影響を受けない。
	fmt.Fprintf(b, "\n<g font-size=\"11\" fill=\"%s\">", opt.Bars.Color)
	for _, v := range s.TickValues() {
		y := s.Y(v)
		fmt.Fprintf(b,
			"\n<line x1=\"0\" y1=\"%s\" x2=\"%s\" y2=\"%s\" stroke=\"%s\" stroke-width=\"1\" shape-rendering=\"crispEdges\"/>"+
				`<text x="-10" y="%s" text-anchor="end">%s</text>`,
			num(y), num(plotW), num(y), opt.GridColor, num(y+4), esc(opt.Bars.tick(v)),
		)
	}
	b.WriteString("\n</g>")
}

func writeRightAxis(b *strings.Builder, opt Options, s Scale, plotW float64) {
	fmt.Fprintf(b, "\n<g font-size=\"11\" fill=\"%s\">", opt.Line.Color)
	fmt.Fprintf(b, "\n<text x=\"%s\" y=\"-8\" font-size=\"10\" opacity=\"%s\">%s</text>",
		num(plotW+11), rightAxisOpacity, esc("PV"))

	for _, v := range s.TickValues() {
		y := s.Y(v)
		fmt.Fprintf(b,
			"\n<line x1=\"%s\" y1=\"%s\" x2=\"%s\" y2=\"%s\" stroke=\"%s\" stroke-width=\"1\" opacity=\"%s\"/>"+
				`<text x="%s" y="%s">%s</text>`,
			num(plotW), num(y), num(plotW+5), num(y), opt.Line.Color, rightTickOpacity,
			num(plotW+11), num(y+4), esc(opt.Line.tick(v)),
		)
	}
	b.WriteString("\n</g>")
}

func writeBars(b *strings.Builder, opt Options, s Scale, slot, barW, plotH float64) {
	fmt.Fprintf(b, "\n<g fill=\"%s\" opacity=\"%s\">", opt.Bars.Color, barOpacity)
	for i, v := range opt.Bars.Values {
		if !finite(v) {
			continue
		}
		var (
			x = float64(i)*slot + (slot-barW)/2
			y = s.Y(plotValue(v))
		)
		fmt.Fprintf(b, "\n<rect x=\"%s\" y=\"%s\" width=\"%s\" height=\"%s\"/>",
			num(x), num(y), num(barW), num(plotH-y))
	}
	b.WriteString("\n</g>")
}

func writeLine(b *strings.Builder, opt Options, s Scale) {
	type pt struct{ x, y, v float64 }

	pts := make([]pt, 0, len(opt.Line.Values))
	for i, v := range opt.Line.Values {
		if !finite(v) {
			continue // TODO: 欠損でポリラインを分割するならここで区切る
		}
		pts = append(pts, pt{opt.centerX(opt.Line.Offset + i), s.Y(plotValue(v)), v})
	}
	if len(pts) == 0 {
		return
	}

	b.WriteString("\n<polyline points=\"")
	for i, p := range pts {
		if i > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(b, "%s,%s", num(p.x), num(p.y))
	}
	fmt.Fprintf(b, "\" fill=\"none\" stroke=\"%s\" stroke-width=\"2\" stroke-linejoin=\"round\" stroke-linecap=\"round\"/>", opt.Line.Color)

	// 点が多いと白丸が数珠つなぎになって線が読めなくなるので、少ないときだけ打つ。
	if len(pts) <= maxLineDots {
		for _, p := range pts {
			fmt.Fprintf(b, "\n<circle cx=\"%s\" cy=\"%s\" r=\"3\" fill=\"%s\" stroke=\"%s\" stroke-width=\"2\"/>",
				num(p.x), num(p.y), dotFillColor, opt.Line.Color)
		}
	}

	if opt.AnnotateLineMax {
		top := pts[0]
		for _, p := range pts[1:] {
			if p.v > top.v {
				top = p
			}
		}
		fmt.Fprintf(b,
			"\n<text x=\"%s\" y=\"%s\" text-anchor=\"middle\" font-size=\"11\" font-weight=\"600\" fill=\"%s\" stroke=\"white\" stroke-width=\"2\" paint-order=\"stroke fill\">%s</text>",
			num(top.x), num(top.y-10), opt.Line.Color, esc(Comma(top.v)),
		)
	}
}

func writeXAxis(b *strings.Builder, opt Options, plotW, plotH float64) {
	fmt.Fprintf(b, "\n<g font-size=\"11\" fill=\"%s\">", opt.TextColor)
	fmt.Fprintf(b, "\n<line x1=\"0\" y1=\"%s\" x2=\"%s\" y2=\"%s\" stroke=\"%s\" shape-rendering=\"crispEdges\"/>",
		num(plotH), num(plotW), num(plotH), opt.GridColor)

	for _, l := range opt.XLabels {
		fmt.Fprintf(b, "\n<text x=\"%s\" y=\"%s\" text-anchor=\"middle\">%s</text>",
			num(opt.centerX(l.At)), num(plotH+20), esc(l.Text))
	}

	// 期間や生成時刻。日付ラベルの 1 行下に右寄せで置く。
	if opt.Caption != "" {
		fmt.Fprintf(b, "\n<text x=\"%s\" y=\"%s\" text-anchor=\"end\" font-size=\"10\" opacity=\"%s\">%s</text>",
			num(plotW), num(plotH+38), captionOpacity, esc(opt.Caption))
	}
	b.WriteString("\n</g>")
}

// writeHitAreas は列ごとの透明な当たり判定を最前面に置き、<title> でホバー文言を出す。
// バー自体に <title> を付けると幅 7.5px を狙う必要があって使いづらいため、
// 列の全高を覆う矩形にしている。塗りが transparent なので見た目には影響しない。
func writeHitAreas(b *strings.Builder, opt Options, plotH, slot float64) {
	if len(opt.SlotTitles) == 0 {
		return
	}

	b.WriteString("\n<g fill=\"transparent\" pointer-events=\"all\">")
	for i, title := range opt.SlotTitles {
		if title == "" {
			continue
		}
		fmt.Fprintf(b, "\n<rect x=\"%s\" y=\"0\" width=\"%s\" height=\"%s\"><title>%s</title></rect>",
			num(float64(i)*slot), num(slot), num(plotH), esc(title))
	}
	b.WriteString("\n</g>")
}

// --- SVG 出力の内部ヘルパー ---

// num は座標値を小数 2 桁までで出力する（整数なら小数部を落とす）。
func num(v float64) string {
	if v == math.Trunc(v) {
		return strconv.FormatFloat(v, 'f', 0, 64)
	}
	return strconv.FormatFloat(v, 'f', 2, 64)
}

func esc(s string) string {
	var buf bytes.Buffer
	_ = xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

// escAttr は二重引用符で囲む属性値用のエスケープ。
// xml.EscapeText はシングルクォートまで &#39; にしてしまい、
// フォントスタック（'Segoe UI' など）が読みづらくなるので必要な3文字だけ潰す。
var attrReplacer = strings.NewReplacer("&", "&amp;", "<", "&lt;", `"`, "&quot;")

func escAttr(s string) string { return attrReplacer.Replace(s) }
