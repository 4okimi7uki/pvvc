package vercel

func (r *Report) TotalCostByDay(projectId string) map[string]float64 {
	var totals = make(map[string]float64)
	for _, charge := range r.Charges {
		if charge.Tags.ProjectID == projectId {
			key := charge.ChargePeriodStart.Format("20060102")
			totals[key] += charge.BilledCost
		}
	}
	return totals
}

func (r *Report) TotalCostByService(projectId string) map[string]float64 {
	var totals = make(map[string]float64)
	for _, charge := range r.Charges {
		if charge.Tags.ProjectID == projectId {
			totals[charge.ServiceName] += charge.BilledCost
		}
	}
	return totals
}
