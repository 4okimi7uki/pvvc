package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/4okimi7uki/pvvc/internal/decimalfmt"
)

func metrics(reports []DailyReport) (decimal.Decimal, decimal.Decimal, [][]string) {
	var (
		allPv       decimal.Decimal
		allCost     decimal.Decimal
		metricsRows [][]string
	)
	metricsRows = append(metricsRows, []string{"Date", "PV", "Cost(USD)", "Cost(JPY)", "Cost/PV(JPY)", "USD/JPY"})
	dash := func(zero bool, s string) string {
		if zero {
			return "-"
		}
		return s
	}
	for _, r := range reports {
		noCost := r.TotalCost.IsZero()
		var _costPerPVJPY decimal.Decimal
		if !noCost && !r.PV.IsZero() {
			_costPerPVJPY = r.TotalCostJPY.Div(r.PV)
		}
		allPv = allPv.Add(r.PV)
		allCost = allCost.Add(r.TotalCost)

		metricsRows = append(metricsRows, []string{
			r.Date.Format("01/02 (Mon)"),
			decimalfmt.DecimalCommaf(r.PV, 0),
			dash(noCost, decimalfmt.DecimalCommaf(r.TotalCost, 4)),
			dash(noCost, decimalfmt.DecimalCommaf(r.TotalCostJPY, 2)),
			dash(noCost || r.PV.IsZero(), decimalfmt.DecimalCommaf(_costPerPVJPY, 4)),
			decimalfmt.DecimalCommaf(r.Rate, 2),
		})

	}
	return allPv, allCost, metricsRows
}

func summary(start, end time.Time, reports []DailyReport, llm string, allPv, allCost decimal.Decimal) []Row {
	var period strings.Builder
	if start.Equal(end) {
		fmt.Fprintf(&period, "%s", start.Format("2006/01/02"))
	} else {
		fmt.Fprintf(&period, "%s → %s", start.Format("2006/01/02"), end.Format("2006/01/02"))
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

	return summaryRows
}

// for slack
var weekdaysJa = [...]string{"日", "月", "火", "水", "木", "金", "土"}

func LatestDaySummary(reports []DailyReport) []Row {
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
		return fmt.Sprintf("%s%%", decimalfmt.DecimalCommaf(pct, 1))
	}

	return []Row{
		{"Date", latest.Date.Format("2006/01/02") + fmt.Sprintf(" (%s)", weekdaysJa[latest.Date.Weekday()])},
		{"PV", fmt.Sprintf("%s   ----------   %s 　vs 7d avg", decimalfmt.DecimalCommaf(latest.PV, 0), formatPct(pvChangePct))},
		{"Cost", fmt.Sprintf("$%s   ----------   %s 　vs 7d avg", decimalfmt.DecimalCommaf(latest.TotalCost, 2), formatPct(costChangePct))},
	}
}

func LatestServiceCosts(end time.Time, reports []DailyReport) []Row {
	var (
		totalCostByService decimal.Decimal
		costByService      []Row
		latestDate         = end.Format("20060102")
	)
	costByService = append(costByService, Row{"SERVICE NAME", "BILLED COST"})
	for _, cs := range reports[0].CostByServices[latestDate] {
		totalCostByService = totalCostByService.Add(cs.EffectiveCost)
		costByService = append(costByService,
			Row{cs.ServiceName, "$" + decimalfmt.DecimalCommaf(cs.EffectiveCost, 4)})
	}
	costByService = append(costByService,
		Row{"---", "---"},
		Row{"TOTAL", "$" + decimalfmt.DecimalCommaf(totalCostByService, 4)},
	)

	return costByService
}
