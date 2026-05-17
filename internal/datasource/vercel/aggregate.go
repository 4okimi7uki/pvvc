package vercel

import (
	"slices"
	"sort"

	"github.com/shopspring/decimal"
)

func (r *Report) TotalCostByDay(projectIds []string) map[string]decimal.Decimal {
	var totals = make(map[string]decimal.Decimal)
	for _, charge := range r.Charges {
		if slices.Contains(projectIds, charge.Tags.ProjectID) {
			key := charge.ChargePeriodStart.Format("20060102")
			EffectiveCost, _ := decimal.NewFromString(string(charge.EffectiveCost))
			totals[key] = totals[key].Add(EffectiveCost)
		}
	}
	return totals
}

func (r *Report) TotalCostByService(projectIds []string) map[string]decimal.Decimal {
	var totals = make(map[string]decimal.Decimal)
	for _, charge := range r.Charges {
		if slices.Contains(projectIds, charge.Tags.ProjectID) {
			serviceName := charge.ServiceName
			EffectiveCost, _ := decimal.NewFromString(string(charge.EffectiveCost))
			totals[serviceName] = totals[serviceName].Add(EffectiveCost)
		}
	}
	return totals
}

func (r *Report) DailyCostByService(projectIds []string) map[string][]ServiceCost {
	type ServiceCostMap = map[string]decimal.Decimal
	type DailyMap = map[string]ServiceCostMap
	intermediate := make(DailyMap)

	for _, charge := range r.Charges {
		if !slices.Contains(projectIds, charge.Tags.ProjectID) {
			continue
		}

		_date := charge.ChargePeriodStart
		date := _date.Format("20060102")
		if intermediate[date] == nil {
			intermediate[date] = make(ServiceCostMap)
		}
		serviceName := charge.ServiceName
		EffectiveCost, _ := decimal.NewFromString(charge.EffectiveCost.String())
		intermediate[date][serviceName] = intermediate[date][serviceName].Add(EffectiveCost)
	}

	result := make(map[string][]ServiceCost)
	for date, services := range intermediate {
		for name, cost := range services {
			result[date] = append(result[date], ServiceCost{
				ServiceName:   name,
				EffectiveCost: cost,
			})
		}
	}

	// sort by EffectiveCost
	for date := range result {
		sort.Slice(result[date], func(i, j int) bool {
			return result[date][i].EffectiveCost.GreaterThan(result[date][j].EffectiveCost)
		})
	}
	return result
}
