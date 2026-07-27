package report

import (
	"fmt"
	"math"
	"time"

	"github.com/4okimi7uki/pvvc/internal/chart"
	"github.com/4okimi7uki/pvvc/internal/decimalfmt"
)

// maxXLabels は X 軸に並べる日付ラベルの最大数。
// これを超える日数は等間隔で間引く。
const maxXLabels = 8

// ChartOptions は日次レポートを chart パッケージが描ける形に変換する。
// 棒グラフが日次コスト(USD)、折れ線が PV。
// generatedAt は図の下に入れる生成時刻。呼び出し側から渡すのは、
// chart 側で time.Now() を呼ぶと出力が毎回変わってゴールデンテストが成立しないため。
func ChartOptions(reports []DailyReport, generatedAt time.Time) chart.Options {
	cost := make([]float64, len(reports))
	pv := make([]float64, len(reports))
	for i, r := range reports {
		cost[i] = r.TotalCost.InexactFloat64()
		pv[i] = r.PV.InexactFloat64()
	}

	return chart.Options{
		AriaLabel:       chartAriaLabel(reports),
		Ticks:           5,
		AnnotateLineMax: true,
		Bars: chart.Series{
			Name:       "Vercel cost / day (USD)",
			Values:     cost,
			FormatTick: chart.USD,
		},
		Line: chart.Series{
			Name:       "GA4 pageviews",
			Values:     pv,
			FormatTick: chart.Compact,
		},
		XLabels:    xLabels(reports, maxXLabels),
		Caption:    chartCaption(reports, generatedAt),
		SlotTitles: slotTitles(reports),
	}
}

// slotTitles は 1 日ぶんのホバー文言。日付・コスト・PV を 1 行にまとめる。
// SVG をブラウザで直接開いたときだけ見えるので、表示できない環境の前提で
// 同じ情報が読めなくなるものは入れない。
func slotTitles(reports []DailyReport) []string {
	out := make([]string, len(reports))
	for i, r := range reports {
		out[i] = fmt.Sprintf("%s\n$%s  ·  %s PV",
			r.Date.Format("2006-01-02 (Mon)"),
			decimalfmt.DecimalCommaf(r.TotalCost, 2),
			decimalfmt.DecimalCommaf(r.PV, 0),
		)
	}
	return out
}

// chartCaption は図の下に置く補足。対象期間と生成時刻。
// 日付ラベルは年を省いているので、ここで年まで含めて補う。
func chartCaption(reports []DailyReport, generatedAt time.Time) string {
	if len(reports) == 0 {
		return ""
	}
	var (
		from = reports[0].Date.Format("2006-01-02")
		to   = reports[len(reports)-1].Date.Format("2006-01-02")
	)
	if from == to {
		return fmt.Sprintf("%s  ·  generated %s", from, generatedAt.Format("2006-01-02 15:04 MST"))
	}
	return fmt.Sprintf("%s → %s (%d days)  ·  generated %s",
		from, to, len(reports), generatedAt.Format("2006-01-02 15:04 MST"))
}

func chartAriaLabel(reports []DailyReport) string {
	const base = "Vercel daily cost with GA4 pageviews overlaid"
	if len(reports) == 0 {
		return base
	}
	return fmt.Sprintf("%s, %s to %s", base,
		reports[0].Date.Format("2006-01-02"),
		reports[len(reports)-1].Date.Format("2006-01-02"),
	)
}

// xLabels は等間隔に最大 limit 個の日付ラベルを返す。
// 両端は必ず含む。月初だけを拾う方式は月の境界が偏るとラベル間隔がガタつくので採らない。
func xLabels(reports []DailyReport, limit int) []chart.XLabel {
	n := len(reports)
	if n == 0 {
		return nil
	}

	// 日数が少ないときは全部出す。
	if limit < 2 || n <= limit {
		out := make([]chart.XLabel, 0, n)
		for i, r := range reports {
			out = append(out, chart.XLabel{At: i, Text: r.Date.Format("Jan 2")})
		}
		return out
	}

	step := float64(n-1) / float64(limit-1)
	out := make([]chart.XLabel, 0, limit)
	prev := -1
	for k := range limit {
		i := int(math.Round(float64(k) * step))
		if i == prev { // 丸めで同じ日に着地したら捨てる
			continue
		}
		prev = i
		out = append(out, chart.XLabel{At: i, Text: reports[i].Date.Format("Jan 2")})
	}
	return out
}
