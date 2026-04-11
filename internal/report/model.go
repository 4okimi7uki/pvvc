package report

import "time"

type DailyReport struct {
	Date         time.Time
	PV           int64
	TotalCost    float64
	TotalCostJPY float64
	Rate         float64
}
