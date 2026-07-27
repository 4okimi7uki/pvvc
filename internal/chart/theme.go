package chart

// チャートの見た目に関する既定値。
// 色や余白を変えたいときはここだけ見れば済むようにしてある。

// 寸法
const (
	defaultWidth  = 1369.0
	defaultHeight = 300.0

	// 軸ラベルと凡例のぶんだけ内側に寄せる。Right は右軸の目盛りラベル用。
	defaultPadTop    = 26.0
	defaultPadRight  = 78.0
	defaultPadBottom = 50.0 // 日付ラベル + Caption の 2 行分
	defaultPadLeft   = 58.0

	defaultTicks = 5 // 横グリッド線の区間数
)

// 色
const (
	defaultBarColor  = "#0072f5" // Vercel bar
	defaultLineColor = "#e06c00" // PV line
	defaultGridColor = "#ebebeb"
	defaultTextColor = "#8f8f8f"
	dotFillColor     = "#fff" // 折れ線の点の中身

	defaultFontFamily = "Geist, -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Helvetica Neue', Arial, sans-serif"
)

// 不透明度。SVG 属性にそのまま流すので文字列で持つ（浮動小数の桁揺れを避ける）。
const (
	barOpacity       = "0.45" // 棒は折れ線より後ろに見せたいので薄く
	highlightOpacity = "0.07" // 折れ線が一部区間だけのときの背景
	swatchOpacity    = "0.85" // 凡例の色見本
	rightTickOpacity = "0.5"  // 右軸の目盛り線
	rightAxisOpacity = "0.8"  // 右軸の見出し（"PV"）
	captionOpacity   = "0.7"  // 図の下の補足テキスト
)

// 形状
const (
	defaultBarRatio = 0.56 // スロット幅に対する棒の太さ

	// maxLineDots は折れ線に白丸を打つ上限の点数。
	// これを超えると数珠つなぎになって線が読めないので、線だけにする。
	maxLineDots = 14
)

// applyDefaults はゼロ値のフィールドに既定値を入れる。
// RenderSVG の冒頭で 1 度だけ呼ばれる。
func (o *Options) applyDefaults() {
	if o.Width == 0 {
		o.Width = defaultWidth
	}
	if o.Height == 0 {
		o.Height = defaultHeight
	}
	if o.Pad == (Padding{}) {
		o.Pad = Padding{
			Top:    defaultPadTop,
			Right:  defaultPadRight,
			Bottom: defaultPadBottom,
			Left:   defaultPadLeft,
		}
	}
	if o.Ticks <= 0 {
		o.Ticks = defaultTicks
	}
	if o.GridColor == "" {
		o.GridColor = defaultGridColor
	}
	if o.TextColor == "" {
		o.TextColor = defaultTextColor
	}
	if o.FontFamily == "" {
		o.FontFamily = defaultFontFamily
	}
	if o.BarRatio == 0 {
		o.BarRatio = defaultBarRatio
	}
	if o.Bars.Color == "" {
		o.Bars.Color = defaultBarColor
	}
	if o.Line.Color == "" {
		o.Line.Color = defaultLineColor
	}
}
