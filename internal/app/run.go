package app

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/auth/credentials"
	"github.com/4okimi7uki/pvvc/internal/fx"
	"github.com/4okimi7uki/pvvc/internal/ga4"
	"github.com/4okimi7uki/pvvc/internal/ui"
	"github.com/4okimi7uki/pvvc/internal/vercel"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	analyticsdata "google.golang.org/api/analyticsdata/v1beta"
)

func RunMain(v *viper.Viper) error {
	ui.PrintLogo()

	var pv int64
	var totalCost float64
	var rate float64
	start := time.Now().AddDate(0, 0, -1)
	end := time.Now()

	err := ui.WithSpinner("Fetching...", func(update func(string), addDone func(string)) error {
		ctx := context.Background()
		eg, ctx := errgroup.WithContext(ctx)

		// GA4 PV
		eg.Go(func() error {
			propertyID := v.GetString("ga4.property_id")
			jsonStr := v.GetString("ga4.credential")

			creds, err := credentials.DetectDefault(&credentials.DetectOptions{
				CredentialsJSON: []byte(jsonStr),
				Scopes:          []string{analyticsdata.AnalyticsReadonlyScope},
			})
			if err != nil {
				return fmt.Errorf("load ga4 credentials: %w", err)
			}

			client, err := ga4.New(ctx, propertyID, creds)
			if err != nil {
				addDone(ui.Red("  ✗ ") + "GA4")
				return err
			}
			report, err := client.FetchDailyPageViews(ctx, "yesterday", "today")
			if err != nil {
				addDone(ui.Red("  ✗ ") + "GA4")
				return err
			}
			pv = report.TotalPageView()
			addDone(ui.Green("  ✓ ") + "GA4")

			return nil
		})

		// Vercel cost
		eg.Go(func() error {
			client, err := vercel.New(
				v.GetString("vercel.token"),
				v.GetString("vercel.team_id"),
			)
			if err != nil {
				addDone(ui.Red("  ✗ ") + "Vercel")
				return fmt.Errorf("failed to create vercel client: %w", err)
			}

			cost, err := client.FetchBillingCharges(
				start,
				end,
			)
			if err != nil {
				addDone(ui.Red("  ✗ ") + "Vercel")
				return fmt.Errorf("failed to vercel fetching: %w", err)
			}
			totalCost = cost.TotalCost()
			addDone(ui.Green("  ✓ ") + "Vercel")

			return nil
		})

		// FX
		eg.Go(func() error {
			var err error
			rate, err = fx.FetchUSDToJPY(end)
			if err != nil {
				addDone(ui.Red("  ✗ ") + "FX")
				return err
			}
			addDone(ui.Green("  ✓ ") + "FX")

			return nil
		})

		if err := eg.Wait(); err != nil {
			return err
		}
		if pv == 0 {
			return fmt.Errorf("PV is 0, cannot calculate cost per PV")
		}

		return nil
	})
	if err != nil {
		return err
	}

	totalCostJP := totalCost * rate

	ui.PrintReport(start, pv, totalCost, totalCostJP, rate)

	return nil
}
