package ga4

import (
	"context"
	"fmt"
	"strconv"

	analyticsdata "google.golang.org/api/analyticsdata/v1beta"
)

func parseReport(propertyID string, resp *analyticsdata.RunReportResponse) (*Report, error) {
	report := &Report{
		PropertyID: propertyID,
		Rows:       make([]DailyPageViews, 0, len(resp.Rows)),
	}

	for _, row := range resp.Rows {
		if len(row.DimensionValues) == 0 || len(row.MetricValues) == 0 {
			continue
		}

		date := row.DimensionValues[0].Value
		// pagePath := row.DimensionValues[1].Value

		views, err := strconv.ParseInt(row.MetricValues[0].Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("ga4: failed to parse views %q: %w", row.MetricValues[0].Value, err)
		}

		report.Rows = append(report.Rows, DailyPageViews{PagePath: "", Views: views, Date: date})
	}

	return report, nil
}

const pageSize = 10000

func buildRunReportRequest(startDate, endDate string, offset int64) *analyticsdata.RunReportRequest {
	return &analyticsdata.RunReportRequest{
		DateRanges: []*analyticsdata.DateRange{
			{StartDate: startDate, EndDate: endDate},
		},
		Dimensions: []*analyticsdata.Dimension{
			{Name: "date"},
			{Name: "sessionSourceMedium"},
			// {Name: "pagePath"},
		},
		Metrics: []*analyticsdata.Metric{
			{Name: "screenPageViews"},
		},
		OrderBys: []*analyticsdata.OrderBy{
			{Dimension: &analyticsdata.DimensionOrderBy{DimensionName: "date"}},
		},
		Limit:  pageSize,
		Offset: offset,
		DimensionFilter: &analyticsdata.FilterExpression{
			AndGroup: &analyticsdata.FilterExpressionList{
				Expressions: []*analyticsdata.FilterExpression{
					{
						NotExpression: &analyticsdata.FilterExpression{
							Filter: &analyticsdata.Filter{
								FieldName: "sessionSourceMedium",
								StringFilter: &analyticsdata.StringFilter{
									MatchType: "PARTIAL_REGEXP",
									Value:     "SmartNews / app",
								},
							},
						},
					},
					{
						Filter: &analyticsdata.Filter{
							FieldName: "pageTitle",
							StringFilter: &analyticsdata.StringFilter{
								MatchType: "CONTAINS",
								Value:     "ゴルフ総合サイト ALBA Net",
							},
						},
					},
				},
			},
		},
	}
}

func (c *Client) FetchDailyPageViews(ctx context.Context, startDate, endDate string) (*Report, error) {
	report := &Report{
		PropertyID: c.propertyID,
		Rows:       make([]DailyPageViews, 0),
	}

	var offset int64
	for {
		req := buildRunReportRequest(startDate, endDate, offset)

		resp, err := c.svc.Properties.RunReport(c.propertyID, req).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("ga4: RunReport failed (offset=%d): %w", offset, err)
		}

		if len(resp.Rows) == 0 {
			break
		}

		page, err := parseReport(c.propertyID, resp)
		if err != nil {
			return nil, err
		}
		report.Rows = append(report.Rows, page.Rows...)

		offset += int64(len(resp.Rows))
		if offset >= resp.RowCount {
			break
		}
	}

	return report, nil
}
