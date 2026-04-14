package cmd

import (
	"context"

	"github.com/4okimi7uki/pvvc/internal/app"
	"github.com/4okimi7uki/pvvc/internal/report"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:          "report",
	SilenceUsage: true,
	Short:        "Generate a traffic vs cost report",
	Long:         "Fetch GA4 pageviews, Vercel costs, and FX rates, then print a traffic-and-cost report to the terminal.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWith(func(ctx context.Context) error {
			rep, err := app.RunMain(cfg, ctx, from, to)
			if err != nil {
				return err
			}
			_ = report.PrintSomeDayReports(from, to, rep, "")

			return nil
		})
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
