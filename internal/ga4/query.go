package ga4

import (
	"context"
	"fmt"
	"strconv"

	analyticsdata "google.golang.org/api/analyticsdata/v1beta"
)

func (c *Client) FetchDailyPageViews(ctx context.Context, startDate, endDate string) (*Report, error) {
	req := &analyticsdata.RunReportRequest{
		DateRanges: []*analyticsdata.DateRange{
			{StartDate: startDate, EndDate: endDate},
		},
		Dimensions: []*analyticsdata.Dimension{
			{Name: "pagePath"},
		},
		Metrics: []*analyticsdata.Metric{
			{Name: "screenPageViews"},
		},
		OrderBys: []*analyticsdata.OrderBy{
			{Dimension: &analyticsdata.DimensionOrderBy{DimensionName: "screenPageViews"}},
		},
	}

	resp, err := c.svc.Properties.RunReport(c.propertyID, req).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("ga4: RunReport failed: %w", err)
	}

	return parseReport(c.propertyID, resp)
}

func parseReport(propertyID string, resp *analyticsdata.RunReportResponse) (*Report, error) {
	report := &Report{
		PropertyID: propertyID,
		Rows:       make([]DailyPageViews, 0, len(resp.Rows)),
	}

	for _, row := range resp.Rows {
		if len(row.DimensionValues) == 0 || len(row.MetricValues) == 0 {
			continue
		}

		pagePath := row.DimensionValues[0].Value

		views, err := strconv.ParseInt(row.MetricValues[0].Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("ga4: failed to parse views %q: %w", row.MetricValues[0].Value, err)
		}

		report.Rows = append(report.Rows, DailyPageViews{PagePath: pagePath, Views: views})
	}

	return report, nil
}
