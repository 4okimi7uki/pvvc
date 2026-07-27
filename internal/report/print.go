package report

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/4okimi7uki/pvvc/internal/datasource/ga4"
	"github.com/4okimi7uki/pvvc/internal/decimalfmt"
	"github.com/4okimi7uki/pvvc/internal/ui"
)

const barWidth = 100

func StrLen(s string) int { return len(s) }

func WriteTableFn(w io.Writer, rows [][]string, widthFn func(string) int) {
	colWidths := make([]int, len(rows[0]))

	for _, row := range rows {
		for i, cell := range row {
			if widthFn(cell) > colWidths[i] {
				colWidths[i] = widthFn(cell) + 3
			}
		}
	}

	for _, row := range rows {
		_, _ = fmt.Fprint(w, " ")
		for i, cell := range row {
			padding := max(colWidths[i]-widthFn(cell), 0)
			_, _ = fmt.Fprintf(w, "%s%s  ", cell, strings.Repeat(" ", padding))
		}
		_, _ = fmt.Fprintln(w)
	}
}

func RowsToCells(rows []Row) [][]string {
	cells := make([][]string, len(rows))
	for i, r := range rows {
		cells[i] = []string{r.Label, r.Value}
	}
	return cells
}

func PrintSection(label string) {
	FprintSection(os.Stdout, label)
}

// FprintSection は見出しの出力先を選べる PrintSection。
// SVG を標準出力に流すときに stderr へ逃がすために使う。
func FprintSection(w io.Writer, label string) {
	line := strings.Repeat(ui.MossGray("─"), barWidth-len(label))
	_, _ = fmt.Fprintf(w, "\n%s %s\n", label, line)
}

// FprintSlackSent は Slack 送信完了のセクション。
// slack.Send は送信だけを担うので、表示はここが持つ。
func FprintSlackSent(w io.Writer) {
	FprintSection(w, "Notification")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, " Sent the analysis result to Slack 🔔")
	_, _ = fmt.Fprintln(w)
}

// FprintSVGBuilt はチャート書き出し完了のセクション。
func FprintSVGBuilt(w io.Writer, path string) {
	FprintSection(w, "SVG")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, " Built the traffic-and-cost chart 📊")
	_, _ = fmt.Fprintf(w, " %s %s\n", "::out::", path)
	_, _ = fmt.Fprintln(w)
}

func PrintSomeDayReports(start, end time.Time, reports []DailyReport, aiResponse string, llm string, topPages []ga4.PagePathRank, topPagesLimit int64) {
	var (
		allPv, allCost, metricsRows = Metrics(reports)
		summaryRows                 = summary(start, end, reports, llm, allPv, allCost)
		costByService               = LatestServiceCosts(end, reports)
		topPath                     = FormatTopPage(topPages)
		endStr                      = end.Format("2006-01-02")
	)

	PrintSection("Summary")
	fmt.Println()
	WriteTableFn(os.Stdout, RowsToCells(summaryRows), StrLen)

	PrintSection("Metrics")
	fmt.Println()
	WriteTableFn(os.Stdout, metricsRows, StrLen)

	PrintSection("Graph of CostJPY/PV")
	fmt.Println()
	printGraph(reports)

	PrintSection(fmt.Sprintf("Service Costs on %s", endStr))
	fmt.Println()
	WriteTableFn(os.Stdout, RowsToCells(costByService), StrLen)

	PrintSection(fmt.Sprintf("Top %d Page Paths on %s", topPagesLimit, endStr))
	fmt.Println()
	WriteTableFn(os.Stdout, RowsToCells(topPath), StrLen)

	if aiResponse != "" {
		PrintSection("AI Analytics")
		fmt.Println()
		fmt.Println(aiResponse)
	}
}

func printGraph(reports []DailyReport) {
	const _maxWidth = 55
	maxWidth := decimal.NewFromInt(int64(_maxWidth))
	var maxCostPerPVJPY, minCostPerPVJPY decimal.Decimal
	initialized := false

	for _, r := range reports {
		if r.PV.IsZero() {
			continue
		}
		c := r.CostJPYPerPV
		if !initialized {
			maxCostPerPVJPY = c
			minCostPerPVJPY = c
			initialized = true
			continue
		}

		if c.GreaterThan(maxCostPerPVJPY) {
			maxCostPerPVJPY = c
		}
		if c.LessThan(minCostPerPVJPY) {
			minCostPerPVJPY = c
		}
	}
	if !initialized || maxCostPerPVJPY.IsZero() {
		fmt.Println("no data")
		return
	}

	allSame := maxCostPerPVJPY.Equal(minCostPerPVJPY)

	for _, r := range reports {
		if r.PV.IsZero() {
			fmt.Printf(" %s ▏  %s\n", r.Date.Format("01/02 (Mon)"), ui.MossGray("no data"))
			continue
		}
		var legend string

		c := r.CostJPYPerPV
		barLen := c.Div(maxCostPerPVJPY).Mul(maxWidth).IntPart()
		costPerPVJPYStr := "¥" + decimalfmt.DecimalCommaf(c, 4)

		switch {
		case allSame:
			legend = ui.LightGray(costPerPVJPYStr)
		case c.Equal(maxCostPerPVJPY):
			legend = ui.Red(costPerPVJPYStr)
		case c.Equal(minCostPerPVJPY):
			legend = ui.Blue(costPerPVJPYStr)
		default:
			legend = ui.LightGray(costPerPVJPYStr)
		}

		fmt.Printf(" %s ▏ %s  %s\n", r.Date.Format("01/02 (Mon)"), strings.Repeat("━", int(barLen)), legend)
	}
}
