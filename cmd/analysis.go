package cmd

import (
	"context"
	"fmt"

	"github.com/4okimi7uki/pvvc/internal/ai"
	"github.com/4okimi7uki/pvvc/internal/ai/claude"
	"github.com/4okimi7uki/pvvc/internal/ai/gemini"
	"github.com/4okimi7uki/pvvc/internal/app"
	"github.com/4okimi7uki/pvvc/internal/report"
	"github.com/4okimi7uki/pvvc/internal/slack"
	"github.com/spf13/cobra"
)

var (
	notify     bool
	promptPath string
)

var analyzeCmd = &cobra.Command{
	Use:          "analyze",
	SilenceUsage: true,
	Short:        "Analyze traffic and cost with AI",
	Long:         "Analyze traffic and hosting cost with AI. This command fetches GA4 pageviews, Vercel costs, and FX rates, prepares the data, and sends it to an AI model for deeper interpretation and summary.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWith(func(ctx context.Context) error {

			// build report
			rep, err := app.RunMain(cfg, ctx, from, to, raw)
			if err != nil {
				return err
			}

			// ai analyze
			var analyzer ai.Analyzer
			serviceName := cfg.GetString("service.name")
			if key := cfg.GetString("ai.claude_key"); key != "" {
				analyzer = claude.New(key, serviceName, promptPath)
			} else if key := cfg.GetString("ai.gemini_key"); key != "" {
				analyzer = gemini.New(key, serviceName, promptPath)
			} else {
				return fmt.Errorf("no AI key configured")
			}

			analysisResult, err := app.RunAnalysis(analyzer, ctx, rep)
			if err != nil {
				return err
			}

			if !quiet {
				report.PrintSomeDayReports(from, to, rep, analysisResult)
			}

			if notify {
				slackClient, err := slack.New(cfg.GetString("slack.webhook_url"), cfg.GetString("service.name"))
				if err != nil {
					return err
				}
				summary := report.LatestDaySummary(to, rep)
				err = slackClient.Send(ctx, analysisResult, summary)

				if err != nil {
					return err
				}
			}

			return nil
		})
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().BoolVar(&notify, "notify", false, "notify Slack with the analysis result")
	analyzeCmd.Flags().StringVarP(&promptPath, "prompt", "p", "", "path to a custom prompt template file")

}
