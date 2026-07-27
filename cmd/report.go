package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/4okimi7uki/pvvc/internal/app"
	"github.com/4okimi7uki/pvvc/internal/report"
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

			if printResult() {
				report.PrintSomeDayReports(rootOpts.from, rootOpts.to, rep, "", "", topPages, rootOpts.topPagesLimit)
			}

			// report コマンドは AI 分析を伴わないので aiBody / llm は空。
			if err := notifySlack(ctx, rep, topPages, "", ""); err != nil {
				return err
			}
			if err := writeChart(rep); err != nil {
				return err
			}

			return nil
		})
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
	addCommonFlags(reportCmd)
}
