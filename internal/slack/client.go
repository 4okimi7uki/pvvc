package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/4okimi7uki/pvvc/internal/datasource/ga4"
	"github.com/4okimi7uki/pvvc/internal/httpclient"
	rep "github.com/4okimi7uki/pvvc/internal/report"
	"github.com/4okimi7uki/pvvc/internal/retry"
)

type Client struct {
	webhookURL       string
	httpClient       *http.Client
	serviceName      string
	vercelProjectURL string
	serviceURL       string
	chartURL         string
}

func New(webhookURL, serviceName, vercelProjectURL, serviceURL, chartURL string) (*Client, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("slack: webhook url is required")
	}
	return &Client{
		webhookURL:       webhookURL,
		httpClient:       httpclient.New(),
		serviceName:      serviceName,
		vercelProjectURL: vercelProjectURL,
		serviceURL:       serviceURL,
		chartURL:         chartURL,
	}, nil
}

func (c *Client) Send(ctx context.Context, aiBody string, end time.Time, report []rep.DailyReport, llm string, topPages []ga4.PagePathRank, topPageLimit int64) error {
	var (
		summary           = rep.LatestDaySummary(report)
		costByService     = rep.LatestServiceCosts(end, report)
		_, _, metricsRows = rep.Metrics(report)
		topPath           = rep.FormatTopPage(topPages)
	)

	blocks := buildSlackBlocks(c.serviceName, llm, aiBody, c.vercelProjectURL, c.serviceURL, c.chartURL, metricsRows, summary, costByService, end, topPath, topPageLimit)

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

	return nil
}
