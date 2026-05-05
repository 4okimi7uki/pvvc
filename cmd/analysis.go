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

type analyzeFlags struct {
	notify     bool
	promptPath string
	llm        string
}

var analyzeOpts analyzeFlags

var analyzeCmd = &cobra.Command{
	Use:          "analyze",
	SilenceUsage: true,
	Short:        "Analyze traffic and cost with AI",
	Long:         "Analyze traffic and hosting cost with AI. This command fetches GA4 pageviews, Vercel costs, and FX rates, prepares the data, and sends it to an AI model for deeper interpretation and summary.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWith(func(ctx context.Context) error {
			serviceName := cfg.GetString("service.name")

			// build report
			rep, err := app.RunMain(cfg, ctx, from, to, raw)
			if err != nil {
				return err
			}

			// ai analyze
			var analyzer ai.Analyzer
			switch analyzeOpts.llm {
			case "gemini", "":
				if key := cfg.GetString("ai.gemini_key"); key != "" {
					analyzer = gemini.New(key, serviceName, analyzeOpts.promptPath)
				} else {
					return fmt.Errorf("no AI key configured")
				}
			case "claude":
				if key := cfg.GetString("ai.claude_key"); key != "" {
					analyzer = claude.New(key, serviceName, analyzeOpts.promptPath)
				} else {
					return fmt.Errorf("no AI key configured")
				}
			default:
				return fmt.Errorf("unknown LLM: %s", analyzeOpts.llm)
			}

			analysisResult, err := app.RunAnalysis(analyzer, ctx, rep)
			if err != nil {
				return err
			}

			if !quiet {
				report.PrintSomeDayReports(from, to, rep, analysisResult, analyzeOpts.llm)
			}

			if analyzeOpts.notify {
				slackClient, err := slack.New(cfg.GetString("slack.webhook_url"), serviceName)
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
	analyzeCmd.Flags().BoolVar(&analyzeOpts.notify, "notify", false, "notify Slack with the analysis result")
	analyzeCmd.Flags().StringVarP(&analyzeOpts.promptPath, "prompt", "p", "", "path to a custom prompt template file")
	analyzeCmd.Flags().StringVar(
		&analyzeOpts.llm,
		"llm",
		"gemini",
		"LLM provider/model to use for AI analysis: gemini, claude",
	)
}
