package slack

import (
	"fmt"
	"slices"
	"strings"
	"time"

	rep "github.com/4okimi7uki/pvvc/internal/report"
)

const slackTextMaxLength = 3000

func buildHeader(serviceName string, summary []rep.Row) []Block {
	var headerBlock []Block
	var summaryText strings.Builder
	headingTitle := fmt.Sprintf("📊 %s Daily Report", serviceName)

	summaryText.WriteString("*Summary*\n")
	// この書き方の方が綺麗に並ぶ気がする
	for _, row := range summary {
		fmt.Fprintf(&summaryText, "%-*s %s\n", 25-len(row.Label), row.Label, row.Value)
	}

	headerBlock = append(headerBlock,
		Block{
			Type: "header",
			Text: &TextObject{
				Type:  "plain_text",
				Text:  headingTitle,
				Emoji: true,
			},
		},
		Block{
			Type: "context",
			Elements: []TextObject{
				{
					Type:  "plain_text",
					Text:  "Powered by P.V.V.C.",
					Emoji: true,
				},
			},
		},
		Block{Type: "divider"},
		Block{
			Type: "section",
			Text: &TextObject{Type: "mrkdwn", Text: summaryText.String()},
		},
		Block{Type: "divider"},
	)
	return headerBlock
}

func buildAiAnalyzeSection(llm string, aiBody string) []Block {
	var aiAnalyzeBlock []Block
	var aiTitle strings.Builder

	switch llm {
	case "", "gemini":
		_, _ = fmt.Fprint(&aiTitle, ":google-gemini: *AI分析*")
	case "claude":
		_, _ = fmt.Fprint(&aiTitle, ":claude_ai_symbol: *AI分析*")
	default:
		_, _ = fmt.Fprint(&aiTitle, "🤖 *AI分析*")
	}

	aiAnalyzeBlock = append(aiAnalyzeBlock,
		Block{
			Type: "section",
			Text: &TextObject{Type: "mrkdwn", Text: aiTitle.String()},
		},
		Block{
			Type: "section",
			Text: &TextObject{Type: "mrkdwn", Text: truncate(aiBody, slackTextMaxLength)},
		},
		Block{Type: "divider"},
	)

	return aiAnalyzeBlock
}

func buildMetricsSection(metricsRows [][]string) []Block {
	var metricsTable strings.Builder
	var metricsBlock []Block

	rep.WriteTableFn(&metricsTable, metricsRows, rep.StrLen)

	metricsBlock = append(metricsBlock,
		Block{
			Type: "section",
			Text: &TextObject{
				Type: "mrkdwn",
				Text: fmt.Sprintln("🗓️ *直近の推移データ*"),
			},
		},
		Block{
			Type: "section",
			Text: &TextObject{
				Type: "mrkdwn",
				Text: fmt.Sprint("```\n" + metricsTable.String() + "```"),
			},
		},
		Block{Type: "divider"},
	)

	return metricsBlock
}

func buildVercelCostSection(end time.Time, costByService []rep.Row) []Block {
	var vercelCostBlock []Block
	var costDetail strings.Builder

	fmt.Fprint(&costDetail, "```\n")
	rep.WriteTableFn(&costDetail, rep.RowsToCells(costByService), rep.StrLen)
	fmt.Fprint(&costDetail, "```")

	vercelCostBlock = append(vercelCostBlock,
		Block{
			Type: "section",
			Text: &TextObject{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*:vercel: コスト内訳（%s）*", end.Format("01/02")),
			},
		},
		Block{
			Type: "section",
			Text: &TextObject{Type: "mrkdwn", Text: truncate(costDetail.String(), slackTextMaxLength)},
		},
		Block{Type: "divider"},
	)

	return vercelCostBlock
}

func buildGa4TopPathSection(end time.Time, topPages []rep.Row, topPageLimit int64, serviceURL string) []Block {
	var ga4TopPathBlock []Block
	topPageWithHeader := append([]rep.Row{{Label: "PATH", Value: "PV"}}, toLinkedRows(topPages, serviceURL)...)

	chunkedPaths := chunkRowBySize(topPageWithHeader, slackTextMaxLength)

	ga4TopPathBlock = append(ga4TopPathBlock,
		Block{
			Type: "section",
			Text: &TextObject{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*:google_analytics_アナリティクス: Top %d アクセスパス（%s）*", topPageLimit, end.Format("01/02")),
			},
		},
	)
	for _, cp := range chunkedPaths {
		var paths strings.Builder
		fmt.Fprint(&paths, "```\n")
		rep.WriteTableFn(&paths, rep.RowsToCells(cp), slackDisplayLen)
		fmt.Fprint(&paths, "```")

		ga4TopPathBlock = append(ga4TopPathBlock,
			Block{
				Type: "section",
				Text: &TextObject{
					Type: "mrkdwn",
					Text: paths.String(),
				},
			},
		)
	}
	ga4TopPathBlock = append(ga4TopPathBlock, Block{Type: "divider"})

	return ga4TopPathBlock
}

func buildLinkSection(vercelProjectURL string) []Block {
	if vercelProjectURL == "" {
		return nil
	}
	var linkSection []Block
	var linkBody strings.Builder

	fmt.Fprint(&linkBody, "🔗 *Links*\n")
	fmt.Fprintf(&linkBody, " - <%s/usage|Vercel Usage>\n", vercelProjectURL)
	fmt.Fprintf(&linkBody, " - <%s/logs|Vercel Logs>\n", vercelProjectURL)

	linkSection = append(linkSection,
		Block{
			Type: "section",
			Text: &TextObject{Type: "mrkdwn", Text: truncate(linkBody.String(), slackTextMaxLength)},
		},
	)
	return linkSection
}

func buildSlackBlocks(serviceName, llm, aiBody, vercelProjectURL, serviceURL string, metrics [][]string, summary, costByService []rep.Row, end time.Time, topPages []rep.Row, topPageLimit int64) []Block {
	var blocks []Block
	var mainSection []Block

	header := buildHeader(serviceName, summary)
	if aiBody != "" {
		mainSection = buildAiAnalyzeSection(llm, aiBody)
	} else {
		mainSection = buildMetricsSection(metrics)
	}
	vercelCostSection := buildVercelCostSection(end, costByService)

	topPathSection := buildGa4TopPathSection(end, topPages, topPageLimit, serviceURL)

	blocks = slices.Concat(header, mainSection, vercelCostSection, topPathSection)

	// 末尾に結合する
	blocks = slices.Concat(blocks, buildLinkSection(vercelProjectURL))

	return blocks
}
