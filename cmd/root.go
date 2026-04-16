package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/4okimi7uki/pvvc/internal/config"
	"github.com/4okimi7uki/pvvc/internal/gh"
	"github.com/4okimi7uki/pvvc/internal/ui"
	"github.com/spf13/cobra"
)

var cfg = config.New()
var (
	showVersion bool
	version     = "v0.0.0-dev"
	quiet       bool
	raw         bool
)
var (
	from time.Time
	to   time.Time
)

var rootCmd = &cobra.Command{
	Use:          "pvvc",
	SilenceUsage: true,
	Short:        "Analyze Vercel cost against GA4 pageviews",
	Long:         "pvvc fetches GA4 pageviews, Vercel costs, and FX rates to help you report on and analyze the relationship between traffic and hosting cost.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if from.After(to) || from.Equal(to) {
			return fmt.Errorf("--from must be before --to")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			resolvedVersion := gh.ResolvedVersion(version)
			fmt.Printf("%s %s\n", resolvedVersion, ui.Mastered("(PVVC)"))

			// check latest version
			PrintCheckLatestVersion(version)
			return nil
		}
		return cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "print version information")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "no print result")
	rootCmd.PersistentFlags().BoolVar(&raw, "raw", false, "print raw API responses from GA4 and Vercel")
	_ = rootCmd.PersistentFlags().MarkHidden("raw")

	// Default: 1 week
	rootCmd.PersistentFlags().TimeVar(&from, "from", time.Now().AddDate(0, 0, -7), []string{
		"2006-01-02",
		time.RFC3339,
	}, "start date of the report period (e.g. 2006-01-02)")
	rootCmd.PersistentFlags().TimeVar(&to, "to", time.Now(), []string{
		"2006-01-02",
		time.RFC3339,
	}, "end date of the report period (e.g. 2006-01-03)")
}

func runWith(fn func(ctx context.Context) error) error {
	executeTime := time.Now()

	ctx := context.Background()
	ui.PrintLogo()

	for _, w := range config.Warnings(cfg) {
		fmt.Fprintf(os.Stderr, "%s %s\n", ui.Yellow("⚠"), ui.Yellow(w))
	}
	fmt.Fprintln(os.Stderr)

	err := fn(ctx)
	if err != nil {
		return err
	}

	elapsed := time.Since(executeTime)
	fmt.Printf("───\nDone in %.1fs 🕊️\n\n", elapsed.Seconds())

	return nil
}

func PrintCheckLatestVersion(version string) {
	resolvedVersion := gh.ResolvedVersion(version)
	if msg, err := gh.CheckLatestVersion("4okimi7uki", "pvvc", resolvedVersion); err == nil && msg != "" {
		_, _ = fmt.Fprintf(os.Stdout, "%s\n", ui.LimeYellow(msg))
		_, _ = fmt.Fprintf(os.Stdout, "%s\n\n", "https://github.com/4okimi7uki/pvvc/releases")
	}
}
