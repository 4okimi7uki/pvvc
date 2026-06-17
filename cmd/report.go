package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/4okimi7uki/pvvc/internal/app"
	"github.com/4okimi7uki/pvvc/internal/report"
	"github.com/4okimi7uki/pvvc/internal/slack"
)

var reportCmd = &cobra.Command{
	Use:          "report",
	SilenceUsage: true,
	Short:        "Generate a traffic vs cost report",
	Long:         "Fetch GA4 pageviews, Vercel costs, and FX rates, then print a traffic-and-cost report to the terminal.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWith(func(ctx context.Context) error {
			rep, topPages, err := app.RunMain(cfg, ctx, rootOpts.from, rootOpts.to, raw, rootOpts.topPagesLimit)
			if err != nil {
				return err
			}

			if !quiet {
				report.PrintSomeDayReports(rootOpts.from, rootOpts.to, rep, "", "", topPages, rootOpts.topPagesLimit)
			}

			serviceName := cfg.GetString("service.name")
			if rootOpts.notify {
				slackClient, err := slack.New(cfg.GetString("slack.webhook_url"), serviceName, cfg.GetString("vercel.project_url"), cfg.GetString("service.url"))
				if err != nil {
					return err
				}
				analysisResult := ""
				llm := ""

				err = slackClient.Send(ctx, analysisResult, rootOpts.to, rep, llm, topPages, rootOpts.topPagesLimit)

				if err != nil {
					return err
				}
			}

			return nil
		})
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
	addCommonFlags(reportCmd)
}
