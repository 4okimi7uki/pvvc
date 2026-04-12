package ga4

func (r *Report) TotalPageViewByDay() map[string]int64 {
	var total = make(map[string]int64)
	for _, row := range r.Rows {
		key := row.Date
		total[key] += row.Views
	}
	return total
}
