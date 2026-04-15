package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/4okimi7uki/pvvc/internal/report"
	"github.com/4okimi7uki/pvvc/internal/retry"
)

type Client struct {
	webhookURL  string
	httpClient  *http.Client
	serviceName string
}

func New(webhookURL string, serviceName string) (*Client, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("slack: webhook url is required")
	}
	return &Client{
		webhookURL:  webhookURL,
		httpClient:  &http.Client{},
		serviceName: serviceName,
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

func (c *Client) Send(ctx context.Context, text string, summary []report.Row) error {
	var sb strings.Builder
	sb.WriteString("*Summary*\n")

	for _, row := range summary {
		fmt.Fprintf(&sb, "%-*s %s\n", 25-len(row.Label), row.Label, row.Value)
	}
	summaryText := sb.String()
	headingTitle := fmt.Sprintf("📊 %s Daily Report", c.serviceName)

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
					Text: truncate(text, 3000),
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
	report.PrintSection("Notification")
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
