package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/4okimi7uki/pvvc/internal/httpclient"
	rep "github.com/4okimi7uki/pvvc/internal/report"
	"github.com/4okimi7uki/pvvc/internal/retry"
)

const slackTextMaxLength = 3000

type Client struct {
	webhookURL       string
	httpClient       *http.Client
	serviceName      string
	vercelProjectURL string
}

func New(webhookURL, serviceName, vercelProjectURL string) (*Client, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("slack: webhook url is required")
	}
	return &Client{
		webhookURL:       webhookURL,
		httpClient:       httpclient.New(),
		serviceName:      serviceName,
		vercelProjectURL: vercelProjectURL,
	}, nil
}

type TextObject struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Emoji bool   `json:"emoji,omitempty"`
}

type Block struct {
	Type     string       `json:"type"`
	Text     *TextObject  `json:"text,omitempty"`
	Elements []TextObject `json:"elements,omitempty"`
}

type blockPayload struct {
	Blocks []Block `json:"blocks"`
}

func (c *Client) Send(ctx context.Context, text string, end time.Time, report []rep.DailyReport, llm string) error {
	summary := rep.LatestDaySummary(report)
	costByService := rep.LatestServiceCosts(end, report)

	var sb strings.Builder
	var costDetail strings.Builder
	var linkSection strings.Builder
	var aiTitle strings.Builder

	sb.WriteString("*Summary*\n")
	// この書き方の方が綺麗に並ぶ気がする
	for _, row := range summary {
		fmt.Fprintf(&sb, "%-*s %s\n", 25-len(row.Label), row.Label, row.Value)
	}

	fmt.Fprint(&costDetail, "```\n")
	rep.WriteTable(&costDetail, rep.RowsToCells(costByService))
	fmt.Fprint(&costDetail, "```")

	var linkList []rep.Row
	if c.vercelProjectURL != "" {
		linkList = append(linkList, []rep.Row{
			{Label: "Usage", Value: c.vercelProjectURL + "/usage"},
			{Label: "Logs", Value: c.vercelProjectURL + "/logs"},
		}...)
	}
	fmt.Fprint(&linkSection, "🔗 *Links*\n")
	for _, l := range linkList {
		fmt.Fprintf(&linkSection, " - <%s|%s>\n", l.Value, l.Label)
	}

	summaryText := sb.String()
	costDetailText := costDetail.String()
	linkSectionText := linkSection.String()
	headingTitle := fmt.Sprintf("📊 %s Daily Report", c.serviceName)
	switch llm {
	case "", "gemini":
		_, _ = fmt.Fprint(&aiTitle, ":google-gemini: *AI分析*")
	case "claude":
		_, _ = fmt.Fprint(&aiTitle, ":claude_ai_symbol: *AI分析*")
	default:
		_, _ = fmt.Fprint(&aiTitle, "🤖 *AI分析*")
	}

	body, err := json.Marshal(blockPayload{
		Blocks: []Block{
			{
				Type: "header",
				Text: &TextObject{
					Type:  "plain_text",
					Text:  headingTitle,
					Emoji: true,
				},
			},
			{
				Type: "context",
				Elements: []TextObject{
					{
						Type:  "plain_text",
						Text:  "Powered by P.V.V.C.",
						Emoji: true,
					},
				},
			},
			{
				Type: "divider",
			},
			{
				Type: "section",
				Text: &TextObject{
					Type: "mrkdwn",
					Text: summaryText,
				},
			},
			{
				Type: "divider",
			},
			{
				Type: "section",
				Text: &TextObject{
					Type: "mrkdwn",
					Text: aiTitle.String(),
				},
			},
			{
				Type: "section",
				Text: &TextObject{
					Type: "mrkdwn",
					Text: truncate(text, slackTextMaxLength),
				},
			},
			{
				Type: "divider",
			},
			{
				Type: "section",
				Text: &TextObject{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*:vercel: コスト内訳（%s）*", end.AddDate(0, 0, -1).Format("01/02")),
				},
			},
			{
				Type: "section",
				Text: &TextObject{
					Type: "mrkdwn",
					Text: truncate(costDetailText, slackTextMaxLength),
				},
			},
			{
				Type: "divider",
			},
			{
				Type: "section",
				Text: &TextObject{
					Type: "mrkdwn",
					Text: truncate(linkSectionText, slackTextMaxLength),
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("slack: failed to marshal payload: %w", err)
	}

	var resp *http.Response
	if err := retry.Do(ctx, 3, func() error {
		req, e := http.NewRequestWithContext(ctx, http.MethodPost, c.webhookURL, bytes.NewReader(body))
		if e != nil {
			return e
		}
		req.Header.Set("Content-Type", "application/json")
		resp, e = c.httpClient.Do(req)
		return e
	}); err != nil {
		return fmt.Errorf("slack: request failed %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack: unexpected status %d: %s", resp.StatusCode, string(body))
	}
	rep.PrintSection("Notification")
	fmt.Println()
	fmt.Println(" Sent the analysis result to Slack 🔔")
	fmt.Println()

	return nil
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max])
}
