package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/auth/credentials"
	"github.com/4okimi7uki/pvvc/internal/ai/gemini"
	"github.com/4okimi7uki/pvvc/internal/config"
	"github.com/4okimi7uki/pvvc/internal/datasource/ga4"
	"github.com/4okimi7uki/pvvc/internal/datasource/vercel"
	"github.com/4okimi7uki/pvvc/internal/report"
	"github.com/4okimi7uki/pvvc/internal/ui"
	"github.com/spf13/viper"
	analyticsdata "google.golang.org/api/analyticsdata/v1beta"
)

func RunMain(v *viper.Viper, ctx context.Context, start, end time.Time, raw bool) ([]report.DailyReport, error) {
	if err := config.Validate(v); err != nil {
		return nil, err
	}

	propertyID := v.GetString("ga4.property_id")
	jsonStr := v.GetString("ga4.credential")

	creds, err := credentials.DetectDefault(&credentials.DetectOptions{
		CredentialsJSON: []byte(jsonStr),
		Scopes:          []string{analyticsdata.AnalyticsReadonlyScope},
	})
	if err != nil {
		return []report.DailyReport{}, fmt.Errorf("load ga4 credentials: %w", err)
	}

	ga4Client, err := ga4.New(ctx, propertyID, creds)
	if err != nil {
		return []report.DailyReport{}, err
	}
	ga4Client.Raw = raw

	vercelClient, err := vercel.New(
		v.GetString("vercel.token"),
		v.GetString("vercel.team_id"),
	)
	if err != nil {
		return []report.DailyReport{}, fmt.Errorf("failed to create vercel client: %w", err)
	}
	vercelClient.Raw = raw

	var reports []report.DailyReport
	err = ui.WithSpinner("Fetching...", func(update func(string), addDone func(string)) error {
		reports, err = report.FetchDailyReport(v, ctx, ga4Client, vercelClient, start, end, addDone)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return []report.DailyReport{}, err
	}

	if raw {
		printRawResponses(ga4Client, vercelClient)
	}

	return reports, nil
}

func printRawResponses(ga4Client *ga4.Client, vercelClient *vercel.Client) {
	for i, page := range ga4Client.RawPages {
		var buf bytes.Buffer
		if err := json.Indent(&buf, page, "", "  "); err == nil {
			fmt.Printf("\n=== GA4 Raw Response (page %d) ===\n%s\n", i+1, buf.String())
		}
	}
	if len(vercelClient.RawBody) > 0 {
		var buf bytes.Buffer
		if err := json.Indent(&buf, vercelClient.RawBody, "", "  "); err == nil {
			fmt.Printf("\n=== Vercel Raw Response ===\n%s\n", buf.String())
		}
	}
}

func RunAnalysis(v *viper.Viper, ctx context.Context, reports []report.DailyReport) (string, error) {
	// TODO:　AIを外から切り替えられるようにする
	var analysisResult string
	geminiKey := v.GetString("ai.gemini_key")
	if geminiKey != "" {
		aiClient := gemini.New(geminiKey, v.GetString("service.name"))
		err := ui.WithSpinner("Analyzing...", func(update func(string), addDone func(string)) error {
			var err error
			analysisResult, err = aiClient.Analyze(ctx, reports, update)
			return err
		})
		if err != nil {
			return "", fmt.Errorf("ai analysis: %w", err)
		}
	}
	return analysisResult, nil
}
