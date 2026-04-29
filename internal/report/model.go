package report

import (
	"time"

	"github.com/4okimi7uki/pvvc/internal/datasource/vercel"
)

type DailyReport struct {
	Date           time.Time
	PV             int64
	TotalCost      float64
	TotalCostJPY   float64
	Rate           float64
	CostByServices map[string][]vercel.ServiceCost
}
