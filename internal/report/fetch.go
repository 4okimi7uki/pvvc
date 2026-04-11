package report

import (
	"context"
	"fmt"
	"time"

	"github.com/4okimi7uki/pvvc/internal/fx"
	"github.com/4okimi7uki/pvvc/internal/ga4"
	"github.com/4okimi7uki/pvvc/internal/ui"
	"github.com/4okimi7uki/pvvc/internal/vercel"
	"golang.org/x/sync/errgroup"
)

func FetchDailyReport(
	ctx context.Context,
	ga4Client *ga4.Client,
	vercelClient *vercel.Client,
	start time.Time,
	end time.Time,
	addDone func(string),
) ([]DailyReport, error) {
	var pvs map[string]int64
	var totalCosts map[string]float64
	var rates map[string]float64

	eg, ctx := errgroup.WithContext(ctx)
	// end := start.AddDate(0, 0, 1)

	// GA4 PV
	eg.Go(func() error {
		report, err := ga4Client.FetchDailyPageViews(ctx, start.Format("2006-01-02"), end.Format("2006-01-02"))
		if err != nil {
			addDone(ui.Red("  ✗ ") + "GA4")
			return err
		}
		pvs = report.TotalPageViewByDay()
		addDone(ui.Green("  ✔ ") + "GA4")

		return nil
	})

	// Vercel cost
	eg.Go(func() error {
		cost, err := vercelClient.FetchBillingCharges(
			start,
			end,
		)
		if err != nil {
			addDone(ui.Red("  ✗ ") + "Vercel")
			return fmt.Errorf("failed to vercel fetching: %w", err)
		}

		totalCosts = cost.TotalCostByDay()
		addDone(ui.Green("  ✔ ") + "Vercel")

		return nil
	})

	// FX
	eg.Go(func() error {
		var err error
		rates, err = fx.FetchUSDToJPY(start, end)
		if err != nil {
			addDone(ui.Red("  ✗ ") + "FX")
			return err
		}

		addDone(ui.Green("  ✔ ") + "FX")
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	var reports []DailyReport
	for d := start; !d.After(end.AddDate(0, 0, -1)); d = d.AddDate(0, 0, 1) {
		key := d.Format("20060102")
		cost := totalCosts[key]
		rate := rates[key]
		reports = append(reports, DailyReport{
			Date:         d,
			PV:           pvs[key],
			TotalCost:    cost,
			TotalCostJPY: cost * rate,
			Rate:         rate,
		})
	}
	return reports, nil
}
