package cmd

import (
	"context"

	"github.com/4okimi7uki/pvvc/internal/app"
	"github.com/4okimi7uki/pvvc/internal/report"
	"github.com/4okimi7uki/pvvc/internal/slack"
	"github.com/spf13/cobra"
)

var notify bool

var analyzeCmd = &cobra.Command{
	Use:          "analyze",
	SilenceUsage: true,
	Short:        "Analyze traffic and cost with AI",
	Long:         "Analyze traffic and hosting cost with AI. This command fetches GA4 pageviews, Vercel costs, and FX rates, prepares the data, and sends it to an AI model for deeper interpretation and summary.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWith(func(ctx context.Context) error {

			// build report
			rep, err := app.RunMain(cfg, ctx, from, to)
			if err != nil {
				return err
			}

			// ai analyze
			analysisResult, err := app.RunAnalysis(cfg, ctx, rep)
			if err != nil {
				return err
			}
			summary := report.PrintSomeDayReports(from, to, rep, analysisResult)

			if notify {
				slackClient, err := slack.New(cfg.GetString("slack.webhook_url"), cfg.GetString("service.name"))
				if err != nil {
					return err
				}
				err = slackClient.Send(ctx, analysisResult, summary)

				if err != nil {
					return err
				}
			}

			return nil
		})
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().BoolVarP(&notify, "notify", "", false, "notify Slack with the analysis result")
}
