package app

import (
	"context"
	"fmt"
	"time"

	"github.com/4okimi7uki/pvvc/internal/ga4"
	"github.com/4okimi7uki/pvvc/internal/vercel"
	"github.com/spf13/viper"
)

func RunMain(v *viper.Viper) error {
	ctx := context.Background()
	propertyID := v.GetString("ga4.property_id")

	g_client, err := ga4.New(ctx, propertyID, "./service-account.json")

	if err != nil {
		return err
	}

	_, err = g_client.FetchDailyPageViews(ctx, "2daysAgo", "yesterday")

	// for _, r := range report.Rows {
	// 	fmt.Printf("PV: %d, path: %s\n", r.Views, r.PagePath)
	// }

	vecel_client, err := vercel.New(
		v.GetString("vercel.token"),
		v.GetString("vercel.team_id"),
	)
	if err != nil {
		return fmt.Errorf("failed to create vercel client: %w", err)
	}
	charges, err := vecel_client.FetchBillingCharges(
		time.Now().AddDate(0, 0, -7),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to vercel fetching: %w", err)
	}

	for _, charge := range charges {
		if charge.BilledCost == 0.0 {
			continue
		}

		var start = charge.ChargePeriodStart.Format("2006/01/02")
		var end = charge.ChargePeriodEnd.Format("2006/01/02")

		fmt.Println("---")
		fmt.Printf("Period: %s - %s\n", start, end)
		fmt.Printf("ServiceName: %s\n BilledCost: %f USD\n", charge.ServiceName, charge.BilledCost)
	}

	return nil
}
