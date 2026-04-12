package cmd

import (
	"context"
	"time"

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
		end := time.Now()
		start := end.AddDate(0, 0, -14)
		ctx := context.Background()

		rep, err := app.RunMain(cfg, ctx, start, end)
		if err != nil {
			return err
		}
		report.PrintSomeDayReports(start, end, rep, "")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
