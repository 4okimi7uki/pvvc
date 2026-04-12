package app

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/auth/credentials"
	"github.com/4okimi7uki/pvvc/internal/ga4"
	"github.com/4okimi7uki/pvvc/internal/report"
	"github.com/4okimi7uki/pvvc/internal/ui"
	"github.com/4okimi7uki/pvvc/internal/vercel"
	"github.com/spf13/viper"
	analyticsdata "google.golang.org/api/analyticsdata/v1beta"
)

func RunMain(v *viper.Viper) error {
	ui.PrintLogo()

	end := time.Now()
	start := end.AddDate(0, 0, -14)

	propertyID := v.GetString("ga4.property_id")
	jsonStr := v.GetString("ga4.credential")

	creds, err := credentials.DetectDefault(&credentials.DetectOptions{
		CredentialsJSON: []byte(jsonStr),
		Scopes:          []string{analyticsdata.AnalyticsReadonlyScope},
	})
	if err != nil {
		return fmt.Errorf("load ga4 credentials: %w", err)
	}

	ctx := context.Background()
	ga4Client, err := ga4.New(ctx, propertyID, creds)
	if err != nil {
		return err
	}

	vercelClient, err := vercel.New(
		v.GetString("vercel.token"),
		v.GetString("vercel.team_id"),
	)
	if err != nil {
		return fmt.Errorf("failed to create vercel client: %w", err)
	}

	var reports []report.DailyReport
	err = ui.WithSpinner("Fetching...", func(update func(string), addDone func(string)) error {
		reports, err = report.FetchDailyReport(ctx, ga4Client, vercelClient, start, end, addDone)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	report.PrintSomeDayReports(start, end, reports)
	return nil
}
