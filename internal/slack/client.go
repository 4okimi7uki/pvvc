package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/4okimi7uki/pvvc/internal/httpclient"
	rep "github.com/4okimi7uki/pvvc/internal/report"
	"github.com/4okimi7uki/pvvc/internal/retry"
)

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

func (c *Client) Send(ctx context.Context, aiBody string, end time.Time, report []rep.DailyReport, llm string) error {
	summary := rep.LatestDaySummary(report)
	costByService := rep.LatestServiceCosts(end, report)
	_, _, metricsRows := rep.Metrics(report)

	blocks := buildSlackBlocks(c.serviceName, llm, aiBody, c.vercelProjectURL, metricsRows, summary, costByService, end)

	body, err := json.Marshal(blockPayload{Blocks: blocks})
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
		if e != nil {
			return e
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("slack: unexpected status %d: %s", resp.StatusCode, string(body))
		}
		return nil
	}); err != nil {
		return fmt.Errorf("slack: request failed %w", err)
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
