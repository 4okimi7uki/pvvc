package cmd

import (
	"os"

	"github.com/4okimi7uki/pvvc/internal/app"
	"github.com/4okimi7uki/pvvc/internal/config"
	"github.com/spf13/cobra"
)

var cfg = config.New()

var rootCmd = &cobra.Command{
	Use:          "pvvc",
	SilenceUsage: true,
	Short:        "aaaa",
	Long:         "long.",
	RunE: func(cmd *cobra.Command, args []string) error {

		if err := app.RunMain(cfg); err != nil {
			return err
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
