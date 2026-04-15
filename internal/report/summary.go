package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/4okimi7uki/pvvc/internal/ui"
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

func PrintSection(label string) {
	line := strings.Repeat(ui.MossGray("─"), barWith-len(label))
	fmt.Printf("\n%s %s\n", label, line)
}

type Row struct {
	Label string
	Value string
}

func PrintSomeDayReports(start, end time.Time, reports []DailyReport, aiResponse string) {
	var allPv int64
	var allCost float64

	var metricsRows [][]string
	metricsRows = append(metricsRows, []string{"Date", "PV", "Cost(USD)", "Cost(JPY)", "Cost/PV(JPY)", "USD/JPY"})
	for _, r := range reports {
		_costPerPVJPY := r.TotalCostJPY / float64(r.PV)
		allPv += r.PV
		allCost += r.TotalCost

		metricsRows = append(metricsRows, []string{
			r.Date.Format("01/02 (Mon)"),
			humanize.Comma(r.PV),
			humanize.CommafWithDigits(r.TotalCost, 4),
			humanize.CommafWithDigits(r.TotalCostJPY, 2),
			humanize.CommafWithDigits(_costPerPVJPY, 4),
			humanize.CommafWithDigits(r.Rate, 2),
		})

	}

	summaryRows := []Row{
		{"Period", fmt.Sprintf("%s → %s", start.Format("2006/01/02"), end.AddDate(0, 0, -1).Format("2006/01/02"))},
		{"PV Avg", humanize.Comma(allPv / int64(len(reports)))},
		{"Cost Avg", "$" + humanize.CommafWithDigits(allCost/float64(len(reports)), 4)},
	}
	PrintSection("Summary")
	fmt.Println()
	for _, s := range summaryRows {
		fmt.Printf(" %-10s %s\n", s.Label, s.Value)
	}

	PrintSection("Metrics")
	fmt.Println()
	printTable(metricsRows)

	if aiResponse != "" {
		PrintSection("AI Analytics")
		fmt.Println()
		fmt.Println(aiResponse)
	}
}

// func sameDay(a, b time.Time) bool {
// 	ay, am, ad := a.Date()
// 	by, bm, bd := b.Date()
// 	return ay == by && am == bm && ad == bd
// }

var weekdaysJa = [...]string{"日", "月", "火", "水", "木", "金", "土"}

func LatestDaySummary(end time.Time, reports []DailyReport) []Row {
	fmt.Println(reports)
	otherReports := reports[:len(reports)-1]
	latest := reports[len(reports)-1]

	fmt.Println("-----------------")
	fmt.Println(otherReports)

	fmt.Println(latest)

	var sumPV int64
	var sumCost float64
	for _, r := range otherReports {
		sumPV += r.PV
		sumCost += r.TotalCost
	}
	avgPV := float64(sumPV) / float64(len(otherReports))
	avgCost := sumCost / float64(len(otherReports))

	pvChangePct := (float64(latest.PV) - avgPV) / avgPV * 100
	costChangePct := (latest.TotalCost - avgCost) / avgCost * 100

	formatPct := func(pct float64) string {
		if pct >= 0 {
			return fmt.Sprintf("+%.1f%%", pct)
		}
		return fmt.Sprintf("%.1f%%", pct)
	}

	costPerPV := latest.TotalCost / float64(latest.PV)

	return []Row{
		{"Date", latest.Date.Format("2006/01/02") + fmt.Sprintf(" (%s)", weekdaysJa[latest.Date.Weekday()])},
		{},
		{"PV", fmt.Sprintf("%s   　 %s vs 7d avg 　", humanize.Comma(latest.PV), formatPct(pvChangePct))},
		{"Cost", fmt.Sprintf("$%s   　 %s vs 7d avg 　", humanize.CommafWithDigits(latest.TotalCost, 2), formatPct(costChangePct))},
		{"Cost / PV", fmt.Sprintf("$%s", humanize.CommafWithDigits(costPerPV, 6))},
	}
}
