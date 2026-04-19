package cmd

import (
	"github.com/4okimi7uki/pvvc/internal/app"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:          "init",
	SilenceUsage: true,
	Short:        "Initialize pvvc configuration",
	Long:         "Set up pvvc credentials interactively. This command guides you through configuring the credentials and access tokens required to use pvvc.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.RunSetup()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
