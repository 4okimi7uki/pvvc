package report

import (
	"time"

	"github.com/4okimi7uki/pvvc/internal/datasource/vercel"
	"github.com/shopspring/decimal"
)

type DailyReport struct {
	Date           time.Time
	PV             decimal.Decimal
	TotalCost      decimal.Decimal
	TotalCostJPY   decimal.Decimal
	Rate           decimal.Decimal
	CostByServices map[string][]vercel.ServiceCost
}
