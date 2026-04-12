package cmd

import (
	"github.com/4okimi7uki/pvvc/internal/app"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:          "report",
	SilenceUsage: true,
	Short:        "Generate a traffic vs cost report",
	Long:         "Fetch GA4 pageviews, Vercel costs, and FX rates, then print a traffic-and-cost report to the terminal.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := app.RunMain(cfg); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
