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

func WriteTable(w io.Writer, rows [][]string) {
	colWidths := make([]int, len(rows[0]))

	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell) + 3
			}
		}
	}

	for _, row := range rows {
		_, _ = fmt.Fprint(w, " ")
		for i, cell := range row {
			_, _ = fmt.Fprintf(w, "%-*s  ", colWidths[i], cell)
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
	line := strings.Repeat(ui.MossGray("─"), barWidth-len(label))
	fmt.Printf("\n%s %s\n", label, line)
}

func PrintSomeDayReports(start, end time.Time, reports []DailyReport, aiResponse string, llm string, topPages []ga4.PagePathRank, topPagesLimit int64) {
	var (
		allPv, allCost, metricsRows = Metrics(reports)
		summaryRows                 = summary(start, end, reports, llm, allPv, allCost)
		costByService               = LatestServiceCosts(end, reports)
		topPath                     = formatTopPage(topPages)
	)

	PrintSection("Summary")
	fmt.Println()
	WriteTable(os.Stdout, RowsToCells(summaryRows))

	PrintSection("Metrics")
	fmt.Println()
	WriteTable(os.Stdout, metricsRows)

	PrintSection("Graph of CostJPY/PV")
	fmt.Println()
	printGraph(reports)

	PrintSection("Service Costs on Latest Date")
	fmt.Println()
	WriteTable(os.Stdout, RowsToCells(costByService))

	PrintSection(fmt.Sprintf("Top %d Page Paths", topPagesLimit))
	fmt.Println()
	WriteTable(os.Stdout, RowsToCells(topPath))

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
