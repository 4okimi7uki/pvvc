package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/4okimi7uki/pvvc/internal/ai"
	"github.com/4okimi7uki/pvvc/internal/ai/claude"
	"github.com/4okimi7uki/pvvc/internal/ai/gemini"
	"github.com/4okimi7uki/pvvc/internal/app"
	"github.com/4okimi7uki/pvvc/internal/report"
)

type analyzeFlags struct {
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
			rep, topPages, err := app.RunMain(cfg, ctx, rootOpts.from, rootOpts.to, raw, rootOpts.topPagesLimit)
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

			if printResult() {
				report.PrintSomeDayReports(rootOpts.from, rootOpts.to, rep, analysisResult, analyzeOpts.llm, topPages, rootOpts.topPagesLimit)
			}

			if err := notifySlack(ctx, rep, topPages, analysisResult, analyzeOpts.llm); err != nil {
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
	rootCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().StringVarP(&analyzeOpts.promptPath, "prompt", "p", "", "path to a custom prompt template file")
	analyzeCmd.Flags().StringVar(
		&analyzeOpts.llm,
		"llm",
		"gemini",
		"LLM provider/model to use for AI analysis: gemini, claude",
	)
	addCommonFlags(analyzeCmd)
}
