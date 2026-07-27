package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/4okimi7uki/pvvc/internal/config"
	"github.com/4okimi7uki/pvvc/internal/gh"
	"github.com/4okimi7uki/pvvc/internal/ui"
)

var cfg = config.New()
var (
	showVersion bool
	quiet       bool
	raw         bool
)

type rootFlags struct {
	from          time.Time
	to            time.Time
	notify        bool
	topPagesLimit int64
	svgPath       string
}

var rootOpts rootFlags

var rootCmd = &cobra.Command{
	Use:          "pvvc",
	SilenceUsage: true,
	Short:        "Analyze Vercel cost against GA4 pageviews",
	Long:         "pvvc fetches GA4 pageviews, Vercel costs, and FX rates to help you report on and analyze the relationship between traffic and hosting cost.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if rootOpts.from.After(rootOpts.to) || rootOpts.from.Equal(rootOpts.to) {
			return fmt.Errorf("--from must be before --to")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			resolvedVersion := gh.ResolvedVersion()
			fmt.Printf("%s %s\n", resolvedVersion, ui.Mastered("(PVVC)"))

			// check latest version
			PrintCheckLatestVersion()
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
}

func runWith(fn func(ctx context.Context) error) error {
	executeTime := time.Now()

	resolvedVersion := gh.ResolvedVersion()
	ctx := context.Background()
	ui.FprintLogo(chromeOut(), resolvedVersion)

	configWarnings := config.Warnings(cfg)
	if len(configWarnings) > 0 {
		for _, w := range configWarnings {
			fmt.Fprintf(os.Stderr, "%s %s\n", ui.Yellow("⚠"), ui.Yellow(w))
		}
		fmt.Fprintln(os.Stderr)
	}

	err := fn(ctx)
	if err != nil {
		return err
	}

	elapsed := time.Since(executeTime)
	_, _ = fmt.Fprintf(chromeOut(), "\n───\nDone in %.1fs 🕊️\n\n", elapsed.Seconds())

	return nil
}

func addCommonFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "no print result")
	cmd.PersistentFlags().BoolVar(&raw, "raw", false, "print raw API responses from GA4 and Vercel")
	_ = cmd.PersistentFlags().MarkHidden("raw")
	cmd.PersistentFlags().BoolVar(&rootOpts.notify, "notify", false, "notify Slack with report")
	cmd.PersistentFlags().Int64Var(&rootOpts.topPagesLimit, "top-pages", 20, "access top pages limit")
	cmd.PersistentFlags().StringVar(&rootOpts.svgPath, "svg", "", "write the chart as SVG to this path (default: pvvc-<from>_<to>.svg)")
	cmd.PersistentFlags().Lookup("svg").NoOptDefVal = "auto"

	// Default: 1 week. Use local calendar date stored as UTC midnight to match
	// bare-date parsing (time.Parse("2006-01-02", ...) always returns UTC midnight).
	now := time.Now()
	yesterday := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC)
	cmd.PersistentFlags().TimeVar(&rootOpts.from, "from", yesterday.AddDate(0, 0, -6), []string{
		"2006-01-02",
		time.RFC3339,
	}, "start date of the report period (e.g. 2006-01-02)")
	cmd.PersistentFlags().TimeVar(&rootOpts.to, "to", yesterday, []string{
		"2006-01-02",
		time.RFC3339,
	}, "end date of the report period (e.g. 2006-01-03)")
}

func PrintCheckLatestVersion() {
	ctx := context.Background()
	resolvedVersion := gh.ResolvedVersion()
	if msg, err := gh.CheckLatestVersion(ctx, "4okimi7uki", "pvvc", resolvedVersion); err == nil && msg != "" {
		_, _ = fmt.Fprintf(os.Stdout, "%s\n", ui.LimeYellow(msg))
		_, _ = fmt.Fprintf(os.Stdout, "%s\n\n", "https://github.com/4okimi7uki/pvvc/releases")
	}
}
