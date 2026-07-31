package report

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/4okimi7uki/pvvc/internal/datasource/ga4"
	"github.com/4okimi7uki/pvvc/internal/datasource/vercel"
)

type DailyReport struct {
	Date           time.Time
	PV             decimal.Decimal
	TotalCost      decimal.Decimal
	TotalCostJPY   decimal.Decimal
	CostJPYPerPV   decimal.Decimal
	Rate           decimal.Decimal
	CostByServices map[string][]vercel.ServiceCost
	TopPages       []ga4.PagePathRank
}

type Row struct {
	Label string
	Value string
}
