package vercel

func (r *Report) TotalCostByDay() map[string]float64 {
	var totals = make(map[string]float64)
	for _, charge := range r.Charges {
		key := charge.ChargePeriodStart.Format("20060102")
		totals[key] += charge.BilledCost
	}
	return totals
}

func (r *Report) TotalCostByService() map[string]float64 {
	var totals = make(map[string]float64)
	for _, charge := range r.Charges {
		totals[charge.ServiceName] += charge.BilledCost
	}
	return totals
}
