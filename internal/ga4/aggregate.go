package ga4

func (r *Report) TotalPageView() int64 {
	var total int64
	for _, row := range r.Rows {
		total += row.Views
	}
	return total
}
