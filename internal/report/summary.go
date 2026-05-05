package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/4okimi7uki/pvvc/internal/decimalfmt"
	"github.com/4okimi7uki/pvvc/internal/ui"
	"github.com/shopspring/decimal"
)

const barWidth = 100

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
	line := strings.Repeat(ui.MossGray("─"), barWidth-len(label))
	fmt.Printf("\n%s %s\n", label, line)
}

type Row struct {
	Label string
	Value string
}

func PrintSomeDayReports(start, end time.Time, reports []DailyReport, aiResponse string, llm string) {
	var allPv decimal.Decimal
	var allCost decimal.Decimal

	// metrics
	var metricsRows [][]string
	metricsRows = append(metricsRows, []string{"Date", "PV", "Cost(USD)", "Cost(JPY)", "Cost/PV(JPY)", "USD/JPY"})
	for _, r := range reports {

		var _costPerPVJPY decimal.Decimal
		if !r.PV.IsZero() {
			_costPerPVJPY = r.TotalCostJPY.Div(r.PV)
		}
		allPv = allPv.Add(r.PV)
		allCost = allCost.Add(r.TotalCost)

		metricsRows = append(metricsRows, []string{
			r.Date.Format("01/02 (Mon)"),
			decimalfmt.DecimalCommaf(r.PV, 0),
			decimalfmt.DecimalCommaf(r.TotalCost, 4),
			decimalfmt.DecimalCommaf(r.TotalCostJPY, 2),
			decimalfmt.DecimalCommaf(_costPerPVJPY, 4),
			decimalfmt.DecimalCommaf(r.Rate, 2),
		})

	}

	// summary
	var period strings.Builder
	if start.Equal(end.AddDate(0, 0, -1)) {
		fmt.Fprintf(&period, "%s", start.Format("2006/01/02"))
	} else {
		fmt.Fprintf(&period, "%s → %s", start.Format("2006/01/02"), end.AddDate(0, 0, -1).Format("2006/01/02"))
	}
	reportsLen := decimal.NewFromInt(int64(len(reports)))
	var summaryRows []Row
	if llm != "" {
		summaryRows = append(summaryRows, Row{"LLM", llm}, Row{"", ""})
	}
	summaryRows = append(summaryRows, []Row{
		{"Period", period.String()},
		{"PV", ""},
		{" ⋅ total", decimalfmt.DecimalCommaf(allPv, 0)},
		{" ⋅ avg", decimalfmt.DecimalCommaf(allPv.Div(reportsLen), 0)},
		{"Cost Avg", "$" + decimalfmt.DecimalCommaf(allCost.Div(reportsLen), 4)},
	}...)

	// service
	var (
		totalCostByService decimal.Decimal
		costByService      [][]string
	)

	latestDate := end.AddDate(0, 0, -1).Format("20060102")
	costByService = append(costByService, []string{"SERVICE NAME", "BILLED COST"})
	for _, cs := range reports[0].CostByServices[latestDate] {
		totalCostByService = totalCostByService.Add(cs.BilledCost)
		costByService = append(costByService,
			[]string{cs.ServiceName, "$" + decimalfmt.DecimalCommaf(cs.BilledCost, 4)})
	}
	costByService = append(costByService,
		[]string{"---", "---"},
		[]string{"TOTAL", "$" + decimalfmt.DecimalCommaf(totalCostByService, 4)},
	)

	// Print
	PrintSection("Summary")
	fmt.Println()
	for _, s := range summaryRows {
		fmt.Printf(" %-10s %s\n", s.Label, s.Value)
	}

	PrintSection("Metrics")
	fmt.Println()
	printTable(metricsRows)

	PrintSection("Service Costs on Latest Date")
	fmt.Println()
	printTable(costByService)

	if aiResponse != "" {
		PrintSection("AI Analytics")
		fmt.Println()
		fmt.Println(aiResponse)
	}
}

// for slack
var weekdaysJa = [...]string{"日", "月", "火", "水", "木", "金", "土"}

func LatestDaySummary(end time.Time, reports []DailyReport) []Row {
	otherReports := reports[:len(reports)-1]
	latest := reports[len(reports)-1]

	var sumPV decimal.Decimal
	var sumCost decimal.Decimal
	for _, r := range otherReports {
		sumPV = sumPV.Add(r.PV)
		sumCost = sumCost.Add(r.TotalCost)
	}
	otherReportsLen := decimal.NewFromInt(int64(len(otherReports)))
	avgPV := sumPV.Div(otherReportsLen)
	avgCost := sumCost.Div(otherReportsLen)

	pvChangePct := (latest.PV.Sub(avgPV)).Div(avgPV).Mul(decimal.NewFromInt(100))
	costChangePct := (latest.TotalCost.Sub(avgCost)).Div(avgCost).Mul(decimal.NewFromInt(100))

	formatPct := func(pct decimal.Decimal) string {
		if pct.Sign() >= 0 {
			return fmt.Sprintf("+%s%%", decimalfmt.DecimalCommaf(pct, 1))
		}
		return fmt.Sprintf("%s%%", pct)
	}

	return []Row{
		{"Date", latest.Date.Format("2006/01/02") + fmt.Sprintf(" (%s)", weekdaysJa[latest.Date.Weekday()])},
		{"PV", fmt.Sprintf("%s   ----------   %s 　vs 7d avg", decimalfmt.DecimalCommaf(latest.PV, 0), formatPct(pvChangePct))},
		{"Cost", fmt.Sprintf("$%s   ----------   %s 　vs 7d avg", decimalfmt.DecimalCommaf(latest.TotalCost, 2), formatPct(costChangePct))},
	}
}
