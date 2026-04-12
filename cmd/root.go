package cmd

import (
	"os"

	"github.com/4okimi7uki/pvvc/internal/config"
	"github.com/spf13/cobra"
)

var cfg = config.New()

var rootCmd = &cobra.Command{
	Use:          "pvvc",
	SilenceUsage: true,
	Short:        "Analyze Vercel cost against GA4 pageviews",
	Long:         "pvvc fetches GA4 pageviews, Vercel costs, and FX rates to help you report on and analyze the relationship between traffic and hosting cost.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
