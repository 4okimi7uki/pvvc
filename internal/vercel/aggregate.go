package vercel

func (r *Report) TotalCost() float64 {
	var total float64
	for _, charge := range r.Charges {
		total += charge.BilledCost
	}
	return total
}

func (r *Report) TotalCostByService() map[string]float64 {
	var totals = make(map[string]float64)
	for _, charge := range r.Charges {
		totals[charge.ServiceName] += charge.BilledCost
	}
	return totals
}
