package vercel

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type ServiceCost struct {
	ServiceName string
	BilledCost  decimal.Decimal
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
	BilledCost          json.Number       `json:"BilledCost"`
	BillingCurrency     string            `json:"BillingCurrency"`
	EffectiveCost       json.Number       `json:"EffectiveCost"`
	ServiceName         string            `json:"ServiceName"`
	ServiceCategory     string            `json:"ServiceCategory"`
	ServiceProviderName string            `json:"ServiceProviderName"`
	ConsumedQuantity    json.Number       `json:"ConsumedQuantity"`
	ConsumedUnit        string            `json:"ConsumedUnit"`
	RegionID            string            `json:"RegionId"`
	RegionName          string            `json:"RegionName"`
	Tags                BillingChargeTags `json:"Tags"`
	PricingCategory     string            `json:"PricingCategory"`
	PricingCurrency     string            `json:"PricingCurrency"`
	PricingQuantity     json.Number       `json:"PricingQuantity"`
	PricingUnit         string            `json:"PricingUnit"`
}
