package ga4

import "github.com/shopspring/decimal"

func (r *Report) TotalPageViewByDay() map[string]decimal.Decimal {
	var total = make(map[string]decimal.Decimal)
	for _, row := range r.Rows {
		key := row.Date
		views := decimal.NewFromInt(row.Views)
		total[key] = total[key].Add(views)
	}
	return total
}
