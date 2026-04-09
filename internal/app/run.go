package app

import (
	"context"
	"fmt"
	"time"

	"github.com/4okimi7uki/pvvc/internal/fx"
	"github.com/4okimi7uki/pvvc/internal/ga4"
	"github.com/4okimi7uki/pvvc/internal/ui"
	"github.com/4okimi7uki/pvvc/internal/vercel"
	"github.com/spf13/viper"
)

func RunMain(v *viper.Viper) error {
	ui.PrintLogo()

	err := ui.WithSpinner("Fetching GA4...", func(update func(string)) error {

		ctx := context.Background()
		propertyID := v.GetString("ga4.property_id")

		g_client, err := ga4.New(ctx, propertyID, "./service-account.json")

		if err != nil {
			return err
		}

		report, err := g_client.FetchDailyPageViews(ctx, "2daysAgo", "yesterday")
		pv := report.TotalPageView()
		if err != nil {
			return err
		}

		update("Fetching Vercel...")
		vecel_client, err := vercel.New(
			v.GetString("vercel.token"),
			v.GetString("vercel.team_id"),
		)
		if err != nil {
			return fmt.Errorf("failed to create vercel client: %w", err)
		}

		start := time.Now().AddDate(0, 0, -2)
		end := time.Now().AddDate(0, 0, -1)

		cost, err := vecel_client.FetchBillingCharges(
			start,
			end,
		)
		if err != nil {
			return fmt.Errorf("failed to vercel fetching: %w", err)
		}

		update("Aggregating...")
		totalCost := cost.TotalCost()
		rate, err := fx.FetchUSDToJPY(end)
		if err != nil {
			return err
		}
		totalCostJP := totalCost * rate

		fmt.Println("\n\n---")
		fmt.Println("Cost/PV")
		fmt.Printf("  USD: %f\n", totalCost/float64(pv))
		fmt.Printf("  JPY: %f\n", totalCostJP/float64(pv))
		fmt.Println("---")
		fmt.Printf(" Period: %s - %s\n", start.Format("2006/01/02"), end.Format("2006/01/02"))
		fmt.Printf(" PV: %d\n", pv)
		fmt.Println(" Cost")
		fmt.Printf("   USD: %.2f\n", totalCost)
		fmt.Printf("   JPY: %.2f\n", totalCost*totalCostJP)
		fmt.Printf(" Rate: $1 = ¥%.2f\n", rate)

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
