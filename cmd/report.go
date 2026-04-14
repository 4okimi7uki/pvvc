package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/4okimi7uki/pvvc/internal/app"
	"github.com/4okimi7uki/pvvc/internal/report"
	"github.com/4okimi7uki/pvvc/internal/ui"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:          "report",
	SilenceUsage: true,
	Short:        "Generate a traffic vs cost report",
	Long:         "Fetch GA4 pageviews, Vercel costs, and FX rates, then print a traffic-and-cost report to the terminal.",
	RunE: func(cmd *cobra.Command, args []string) error {
		excuteTime := time.Now()

		end := time.Now()
		start := end.AddDate(0, 0, -7)
		ctx := context.Background()
		ui.PrintLogo()

		rep, err := app.RunMain(cfg, ctx, start, end)
		if err != nil {
			return err
		}
		_ = report.PrintSomeDayReports(start, end, rep, "")

		elapsed := time.Since(excuteTime)
		fmt.Printf("───\nDone in %.1fs 🍭\n\n", elapsed.Seconds())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
