package report

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"github.com/4okimi7uki/pvvc/internal/config"
	"github.com/4okimi7uki/pvvc/internal/datasource/fx"
	"github.com/4okimi7uki/pvvc/internal/datasource/ga4"
	"github.com/4okimi7uki/pvvc/internal/datasource/vercel"
	"github.com/4okimi7uki/pvvc/internal/ui"
)

func FetchDailyReport(
	v *viper.Viper,
	ctx context.Context,
	ga4Client *ga4.Client,
	vercelClient *vercel.Client,
	start time.Time,
	end time.Time,
	addDone func(string),
	topPageLimit int64,
) ([]DailyReport, []ga4.PagePathRank, error) {
	var pvs map[string]decimal.Decimal
	var topPages []ga4.PagePathRank
	var totalCosts map[string]decimal.Decimal
	var dailyCostByService map[string][]vercel.ServiceCost
	var rates map[string]decimal.Decimal

	eg, ctx := errgroup.WithContext(ctx)

	// GA4 PV
	eg.Go(func() error {
		report, err := ga4Client.FetchDailyPageViews(ctx, start.Format("2006-01-02"), end.Format("2006-01-02"))
		if err != nil {
			addDone(ui.Red("  ✗ ") + "GA4 PV")
			return err
		}
		pvs = report.TotalPageViewByDay()
		addDone(ui.Green("  ✔ ") + "GA4 PV")

		return nil
	})
	// GA4 TopPath
	eg.Go(func() error {
		report, err := ga4Client.FetchTopPagePaths(ctx, end.Format("2006-01-02"), end.Format("2006-01-02"), topPageLimit)
		if err != nil {
			addDone(ui.Red("  ✗ ") + "GA4 Top page path")
			return err
		}
		topPages = report.PagePaths
		addDone(ui.Green("  ✔ ") + "GA4 Top page path")

		return nil
	})

	// Vercel cost
	eg.Go(func() error {
		cost, err := vercelClient.FetchBillingCharges(
			ctx,
			start,
			end,
		)
		if err != nil {
			addDone(ui.Red("  ✗ ") + "Vercel")
			return fmt.Errorf("failed to fetch Vercel billing: %w", err)
		}

		projectIds := config.GetProjectIDs(v)
		ieg, _ := errgroup.WithContext(ctx)

		ieg.Go(func() error {
			totalCosts = cost.TotalCostByDay(projectIds)
			return nil
		})
		ieg.Go(func() error {
			dailyCostByService = cost.DailyCostByService(projectIds)
			return nil
		})
		if err = ieg.Wait(); err != nil {
			addDone(ui.Red("  ✗ ") + "Vercel")
			return fmt.Errorf("failed to aggregate costs: %w", err)
		}

		addDone(ui.Green("  ✔ ") + "Vercel")
		return nil
	})

	// FX
	eg.Go(func() error {
		var err error
		rates, err = fx.FetchUSDToJPY(ctx, start, end)
		if err != nil {
			addDone(ui.Red("  ✗ ") + "FX")
			return err
		}

		addDone(ui.Green("  ✔ ") + "FX")
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, nil, err
	}

	var reports []DailyReport
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		key := d.Format("20060102")
		cost := totalCosts[key]
		rate := rates[key]
		totalCostJPY := cost.Mul(rate)
		pv := pvs[key]

		var costJPYPerPV decimal.Decimal
		if pv.IsZero() {
			costJPYPerPV = decimal.Zero
		} else {
			costJPYPerPV = totalCostJPY.Div(pv)
		}

		reports = append(reports, DailyReport{
			Date:           d,
			PV:             pv,
			TotalCost:      cost,
			TotalCostJPY:   totalCostJPY,
			CostJPYPerPV:   costJPYPerPV,
			Rate:           rate,
			CostByServices: dailyCostByService,
		})
	}
	return reports, topPages, nil
}
