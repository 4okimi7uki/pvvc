package vercel

import (
	"sort"
)

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

func (r *Report) DailyCostByService(projectId string) map[string][]ServiceCost {
	type ServiceCostMap = map[string]float64
	type DailyMap = map[string]ServiceCostMap
	intermediate := make(DailyMap)

	for _, charge := range r.Charges {
		if charge.Tags.ProjectID != projectId {
			continue
		}

		_date := charge.ChargePeriodStart
		date := _date.Format("20060102")
		if intermediate[date] == nil {
			intermediate[date] = make(ServiceCostMap)
		}
		intermediate[date][charge.ServiceName] += charge.BilledCost
	}

	result := make(map[string][]ServiceCost)
	for date, services := range intermediate {
		for name, cost := range services {
			result[date] = append(result[date], ServiceCost{
				ServiceName: name,
				BilledCost:  cost,
			})
		}
	}

	// sort by billedCost
	for date := range result {
		sort.Slice(result[date], func(i, j int) bool {
			return result[date][i].BilledCost > result[date][j].BilledCost
		})
	}
	return result
}
