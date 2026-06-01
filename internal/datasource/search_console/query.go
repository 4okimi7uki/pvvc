package searchconsole

import (
	"context"
	"fmt"

	"google.golang.org/api/searchconsole/v1"
)

func (c *Client) FetchSearchConsole(ctx context.Context, startDate, endDate string) error {
	req := &searchconsole.SearchAnalyticsQueryRequest{
		StartDate:  startDate,
		EndDate:    endDate,
		Dimensions: []string{"query"},
		RowLimit:   5,
	}

	resp, err := c.svc.Searchanalytics.Query(c.siteURL, req).Do()
	if err != nil {
		return fmt.Errorf("search console: Query failed: %w", err)
	}

	for _, row := range resp.Rows {
		fmt.Printf("query: %s, clicks: %.0f, position: %.1f\n",
			row.Keys[0], row.Clicks, row.Position)
	}
	return nil
}
