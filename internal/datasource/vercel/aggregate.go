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
			billedCost, _ := decimal.NewFromString(charge.BilledCost.String())
			totals[key] = totals[key].Add(billedCost)
		}
	}
	return totals
}

func (r *Report) TotalCostByService(projectIds []string) map[string]decimal.Decimal {
	var totals = make(map[string]decimal.Decimal)
	for _, charge := range r.Charges {
		if slices.Contains(projectIds, charge.Tags.ProjectID) {
			serviceName := charge.ServiceName
			billedCost, _ := decimal.NewFromString(charge.BilledCost.String())
			totals[serviceName] = totals[serviceName].Add(billedCost)
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
		billedCost, _ := decimal.NewFromString(charge.BilledCost.String())
		intermediate[date][serviceName] = intermediate[date][serviceName].Add(billedCost)
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
			return result[date][i].BilledCost.GreaterThan(result[date][j].BilledCost)
		})
	}
	return result
}
