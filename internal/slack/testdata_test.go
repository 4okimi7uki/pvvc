package slack

import (
	"github.com/4okimi7uki/pvvc/internal/datasource/ga4"
	rep "github.com/4okimi7uki/pvvc/internal/report"
)

// sampleTopPages は ga4.PagePathRank のサンプル。
var sampleTopPages = []ga4.PagePathRank{
	{PagePath: "/", Views: 5000},
	{PagePath: "/blog", Views: 2500},
	{PagePath: "/about", Views: 1200},
	{PagePath: "/contact", Views: 800},
	{PagePath: "/pricing", Views: 600},
}
var formattedSmampleTopPages = rep.FormatTopPage(sampleTopPages)

// // テスト用の基準日時
// var testEndDate = time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

// // makeDailyReport は DailyReport を組み立てるヘルパー。
// func makeDailyReport(dateStr string, pv, totalCostUSD, rate float64, services map[string][]vercel.ServiceCost) rep.DailyReport {
// 	date, _ := time.Parse("20060102", dateStr)
// 	pvDec := decimal.NewFromFloat(pv)
// 	costDec := decimal.NewFromFloat(totalCostUSD)
// 	rateDec := decimal.NewFromFloat(rate)
// 	costJPY := costDec.Mul(rateDec)
// 	var costPerPV decimal.Decimal
// 	if pvDec.IsPositive() {
// 		costPerPV = costJPY.Div(pvDec)
// 	}
// 	return rep.DailyReport{
// 		Date:           date,
// 		PV:             pvDec,
// 		TotalCost:      costDec,
// 		TotalCostJPY:   costJPY,
// 		CostJPYPerPV:   costPerPV,
// 		Rate:           rateDec,
// 		CostByServices: services,
// 	}
// }

// // makeServiceCost は vercel.ServiceCost を組み立てるヘルパー。
// func makeServiceCost(name string, cost float64) vercel.ServiceCost {
// 	c := decimal.NewFromFloat(cost)
// 	return vercel.ServiceCost{
// 		ServiceName:   name,
// 		BilledCost:    c,
// 		EffectiveCost: c,
// 	}
// }

// // ---- サンプルデータ ----

// // sampleReports は 8 日分のサンプルレポート（末尾が「最新日」）。
// // LatestDaySummary / LatestServiceCosts など report パッケージの関数が
// // len >= 2 を前提としているため 8 件用意する。
// var sampleReports = func() []rep.DailyReport {
// 	latestDate := "20240615"
// 	services := map[string][]vercel.ServiceCost{
// 		latestDate: {
// 			makeServiceCost("Edge Functions", 0.0050),
// 			makeServiceCost("Image Optimization", 0.0020),
// 		},
// 	}

// 	reports := []rep.DailyReport{
// 		makeDailyReport("20240608", 9800, 0.0123, 157.0, nil),
// 		makeDailyReport("20240609", 10200, 0.0145, 157.5, nil),
// 		makeDailyReport("20240610", 8500, 0.0098, 156.8, nil),
// 		makeDailyReport("20240611", 11000, 0.0201, 158.0, nil),
// 		makeDailyReport("20240612", 9300, 0.0110, 157.2, nil),
// 		makeDailyReport("20240613", 12000, 0.0230, 158.5, nil),
// 		makeDailyReport("20240614", 10500, 0.0180, 157.8, nil),
// 		makeDailyReport(latestDate, 13500, 0.0250, 159.0, services),
// 	}
// 	return reports
// }()

// // sampleSummaryRows は buildHeader に渡す []rep.Row のサンプル。
// var sampleSummaryRows = []rep.Row{
// 	{Label: "Date", Value: "2024/06/15 (土)"},
// 	{Label: "PV", Value: "13,500   ----------   +30.2% 　vs 7d avg"},
// 	{Label: "Cost", Value: "$0.03   ----------   +40.1% 　vs 7d avg"},
// }

// // sampleCostRows は buildVercelCostSection に渡す []rep.Row のサンプル。
// var sampleCostRows = []rep.Row{
// 	{Label: "SERVICE NAME", Value: "BILLED COST"},
// 	{Label: "Edge Functions", Value: "$0.0050"},
// 	{Label: "Image Optimization", Value: "$0.0020"},
// 	{Label: "---", Value: "---"},
// 	{Label: "TOTAL", Value: "$0.0070"},
// }

// // sampleMetricsRows は buildMetricsSection に渡す [][]string のサンプル。
// var sampleMetricsRows = [][]string{
// 	{"Date", "PV", "Cost(USD)", "Cost(JPY)", "Cost/PV(JPY)", "USD/JPY"},
// 	{"06/08 (Sat)", "9,800", "$0.0123", "¥1.93", "¥0.0002", "157.00"},
// 	{"06/09 (Sun)", "10,200", "$0.0145", "¥2.29", "¥0.0002", "157.50"},
// 	{"06/15 (Sat)", "13,500", "$0.0250", "¥3.98", "¥0.0003", "159.00"},
// }
