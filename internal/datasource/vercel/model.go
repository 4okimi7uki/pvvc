package vercel

import "time"

type ServiceCost struct {
	ServiceName string
	BilledCost  float64
}

type Report struct {
	Charges []BillingCharge
}

type BillingChargeTags struct {
	ProjectID   string `json:"ProjectId"`
	ProjectName string `json:"ProjectName"`
}

type BillingCharge struct {
	ChargePeriodStart   time.Time         `json:"ChargePeriodStart"`
	ChargePeriodEnd     time.Time         `json:"ChargePeriodEnd"`
	ChargeCategory      string            `json:"ChargeCategory"`
	BilledCost          float64           `json:"BilledCost"`
	BillingCurrency     string            `json:"BillingCurrency"`
	EffectiveCost       float64           `json:"EffectiveCost"`
	ServiceName         string            `json:"ServiceName"`
	ServiceCategory     string            `json:"ServiceCategory"`
	ServiceProviderName string            `json:"ServiceProviderName"`
	ConsumedQuantity    float64           `json:"ConsumedQuantity"`
	ConsumedUnit        string            `json:"ConsumedUnit"`
	RegionID            string            `json:"RegionId"`
	RegionName          string            `json:"RegionName"`
	Tags                BillingChargeTags `json:"Tags"`
	PricingCategory     string            `json:"PricingCategory"`
	PricingCurrency     string            `json:"PricingCurrency"`
	PricingQuantity     float64           `json:"PricingQuantity"`
	PricingUnit         string            `json:"PricingUnit"`
}
