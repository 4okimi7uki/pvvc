package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

const barWith = 100

func printTable(rows [][]string) {
	colWidths := make([]int, len(rows[0]))

	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell) + 3
			}
		}
	}

	for _, row := range rows {
		fmt.Print(" ")
		for i, cell := range row {
			fmt.Printf("%-*s  ", colWidths[i], cell)
		}
		fmt.Println()
	}
	fmt.Println()
}

func printSection(label string) {
	line := strings.Repeat("─", barWith-len(label))
	fmt.Printf("\n%s %s\n", label, line)
}

func PrintOneDayReport(r *DailyReport) {
	_costPerPVUSD := r.TotalCost / float64(r.PV)
	_costPerPVJPY := r.TotalCostJPY / float64(r.PV)

	pv := humanize.Comma(r.PV)
	totalCost := humanize.CommafWithDigits(r.TotalCost, 4)
	totalCostJPY := humanize.CommafWithDigits(r.TotalCostJPY, 2)
	costPerPVUSD := humanize.CommafWithDigits(_costPerPVUSD, 6)
	costPerPVJPY := humanize.CommafWithDigits(_costPerPVJPY, 4)
	rate := humanize.CommafWithDigits(r.Rate, 2)

	targetDay := r.Date

	type row struct {
		label string
		value string
	}
	summaryRows := []row{
		{"Period", targetDay.Format("2006/01/02")},
		{"PV", pv},
		{"Rate", fmt.Sprintf("$1 = ¥%s", rate)},
	}

	printSection("Summary")
	fmt.Println()
	for _, s := range summaryRows {
		fmt.Printf(" %-8s %s\n", s.label, s.value)
	}

	printSection("Cost")
	fmt.Println()
	printTable([][]string{
		{"", "USD", "JPY"},
		{"Total", totalCost, totalCostJPY},
		{"/ PV", costPerPVUSD, costPerPVJPY},
	})

	fmt.Println(strings.Repeat("─", barWith))
}

type row struct {
	label string
	value string
}

func PrintSomeDayReports(start, end time.Time, reports []DailyReport) {
	summaryRows := []row{
		{"Period", fmt.Sprintf("%s ~ %s", start.Format("2006/01/02"), end.Format("2006/01/02"))},
	}
	printSection("Summary")
	fmt.Println()
	for _, s := range summaryRows {
		fmt.Printf(" %-8s %s\n", s.label, s.value)
	}

	var tableRows [][]string
	tableRows = append(tableRows, []string{"Date", "PV", "Cost (USD)", "Cost (JPY)", "Cost/PV (USD)", "Cost/PV (JPY)", "Rate"})
	for _, r := range reports {
		_costPerPVUSD := r.TotalCost / float64(r.PV)
		_costPerPVJPY := r.TotalCostJPY / float64(r.PV)

		tableRows = append(tableRows, []string{
			r.Date.Format("01/02 (Mon)"),
			humanize.Comma(r.PV),
			humanize.CommafWithDigits(r.TotalCost, 4),
			humanize.CommafWithDigits(r.TotalCostJPY, 2),
			humanize.CommafWithDigits(_costPerPVUSD, 6),
			humanize.CommafWithDigits(_costPerPVJPY, 4),
			humanize.CommafWithDigits(r.Rate, 2),
		})

	}

	printSection("Weely Cost")
	fmt.Println()
	printTable(tableRows)
}

// ## Metrics

// | Date  | PV     | Cost (USD) | Cost/PV (USD) | Rate    |
// |-------|--------|------------|----------------|---------|
// | 04/04 | 12,450 | 0.0231     | 0.000001856    | ¥149.50 |
// | 04/05 | 18,200 | 0.0312     | 0.000001714    | ¥149.80 |
// ...

// Date         time.Time
// 	PV           int64
// 	TotalCost    float64
// 	TotalCostJPY float64
// 	Rate         float64
