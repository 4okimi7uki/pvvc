package ga4

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/sync/errgroup"
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

		report.Rows = append(report.Rows, DailyPageViews{Views: views, Date: date})
	}

	return report, nil
}
func parsePagePath(propertyID string, resp *analyticsdata.RunReportResponse) (*Report, error) {
	report := &Report{
		PropertyID: propertyID,
		PagePaths:  make([]PagePathRank, 0, len(resp.Rows)),
	}

	for _, row := range resp.Rows {
		if len(row.DimensionValues) == 0 || len(row.MetricValues) == 0 {
			continue
		}

		pagePath := row.DimensionValues[0].Value
		// pagePath := row.DimensionValues[1].Value

		views, err := strconv.ParseInt(row.MetricValues[0].Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("ga4: failed to parse views %q: %w", row.MetricValues[0].Value, err)
		}

		report.PagePaths = append(report.PagePaths, PagePathRank{Views: views, PagePath: pagePath})
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

const (
	dateLayout = "2006-01-02"

	// chunkDays は日次PVを取りに行くときの 1 リクエストあたりの日数。
	//
	// 期間を広げるほど日次PVが実態より小さく出る。日次クエリは
	// date × sessionSourceMedium で行を持つので、期間が長いと組み合わせ数が
	// GA4 のカーディナリティ上限を超え、あふれた分が (other) 行に丸められる。
	// (other) 行は date の値まで (other) になるため、TotalPageViewByDay で
	// どの日にも紐づかず落ちる。実測で 87 日レンジは 26 日レンジの約 1/3 まで
	// PV が落ちた。期間を短く割って足し合わせることで組み合わせ数を抑える。
	chunkDays = 7

	// chunkConcurrency は分割したリクエストの同時実行数。GA4 の同時実行クォータに
	// 余裕を持たせる。上げすぎると 429 が返る。
	chunkConcurrency = 4
)

// dateChunks は [startDate, endDate] を先頭から最大 days 日ずつの区間に割る。
// 返す区間の端は両方とも含む（GA4 の DateRange と同じ閉区間）。
func dateChunks(startDate, endDate string, days int) ([][2]string, error) {
	start, err := time.Parse(dateLayout, startDate)
	if err != nil {
		return nil, fmt.Errorf("ga4: invalid start date %q: %w", startDate, err)
	}
	end, err := time.Parse(dateLayout, endDate)
	if err != nil {
		return nil, fmt.Errorf("ga4: invalid end date %q: %w", endDate, err)
	}
	if end.Before(start) {
		return nil, fmt.Errorf("ga4: end date %s is before start date %s", endDate, startDate)
	}
	if days < 1 {
		days = 1
	}

	var out [][2]string
	for cur := start; !cur.After(end); cur = cur.AddDate(0, 0, days) {
		last := cur.AddDate(0, 0, days-1)
		if last.After(end) {
			last = end
		}
		out = append(out, [2]string{cur.Format(dateLayout), last.Format(dateLayout)})
	}
	return out, nil
}

// FetchDailyPageViews は期間を chunkDays 日ずつに割って取得し、日付順に連結する。
// 1 回のリクエストで全期間を取ると PV が過少に出るため（chunkDays のコメント参照）。
func (c *Client) FetchDailyPageViews(ctx context.Context, startDate, endDate string) (*Report, error) {
	chunks, err := dateChunks(startDate, endDate, chunkDays)
	if err != nil {
		return nil, err
	}

	// 連結順を入力の期間順に保つため、結果はチャンクごとのスロットに書く。
	perChunk := make([][]DailyPageViews, len(chunks))

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(chunkConcurrency)
	for i, ch := range chunks {
		eg.Go(func() error {
			rows, err := c.fetchDailyRange(ctx, ch[0], ch[1])
			if err != nil {
				return err
			}
			perChunk[i] = rows
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	report := &Report{
		PropertyID: c.propertyID,
		Rows:       make([]DailyPageViews, 0),
	}
	for _, rows := range perChunk {
		report.Rows = append(report.Rows, rows...)
	}
	return report, nil
}

// fetchDailyRange は 1 区間ぶんの日次PVを、行数上限ぶんページングしながら取る。
func (c *Client) fetchDailyRange(ctx context.Context, startDate, endDate string) ([]DailyPageViews, error) {
	rows := make([]DailyPageViews, 0)

	var offset int64
	for {
		req := buildRunReportRequest(startDate, endDate, offset)

		resp, err := c.svc.Properties.RunReport(c.propertyID, req).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("ga4: RunReport failed (%s..%s, offset=%d): %w", startDate, endDate, offset, err)
		}

		if c.Raw {
			if b, e := json.Marshal(resp); e == nil {
				c.appendRaw(b)
			}
		}

		if len(resp.Rows) == 0 {
			break
		}

		page, err := parseReport(c.propertyID, resp)
		if err != nil {
			return nil, err
		}
		rows = append(rows, page.Rows...)

		offset += int64(len(resp.Rows))
		if offset >= resp.RowCount {
			break
		}
	}

	return rows, nil
}

func buildTopPagesRequest(startDate, endDate string, offset int64, limit int64) *analyticsdata.RunReportRequest {
	return &analyticsdata.RunReportRequest{
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
			{Metric: &analyticsdata.MetricOrderBy{MetricName: "screenPageViews"}, Desc: true},
		},
		Limit:  limit,
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

// FetchDailyTopPagePaths は期間内の各日について上位ページを取得し、
// 日付(YYYY-MM-DD)をキーに日ごとのランキングを返す。
//
// 1 日ずつ別リクエストで取るのは、date × pagePath を 1 クエリにまとめると
// FetchDailyPageViews と同じくカーディナリティ上限であふれた分が (other) 行に
// 丸められ、日別の精度が落ちるため（chunkDays のコメント参照）。日ごとに閉じた
// 単日リクエストなら組み合わせ数が増えないので取りこぼさない。
func (c *Client) FetchDailyTopPagePaths(ctx context.Context, startDate, endDate string, limit int64) (map[string][]PagePathRank, error) {
	days, err := dateChunks(startDate, endDate, 1)
	if err != nil {
		return nil, err
	}

	// マップへの書き込みは競合するので、いったん日付順のスロットに受ける。
	perDay := make([][]PagePathRank, len(days))

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(chunkConcurrency)
	for i, d := range days {
		eg.Go(func() error {
			report, err := c.FetchTopPagePaths(ctx, d[0], d[1], limit)
			if err != nil {
				return err
			}
			perDay[i] = report.PagePaths
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	out := make(map[string][]PagePathRank, len(days))
	for i, d := range days {
		out[d[0]] = perDay[i] // dateChunks(days=1) なので d[0] == d[1]
	}
	return out, nil
}

func (c *Client) FetchTopPagePaths(ctx context.Context, startDate, endDate string, limit int64) (*Report, error) {
	report := &Report{
		PropertyID: c.propertyID,
		PagePaths:  make([]PagePathRank, 0),
	}

	var offset int64
	for {
		req := buildTopPagesRequest(startDate, endDate, offset, limit)

		resp, err := c.svc.Properties.RunReport(c.propertyID, req).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("ga4: RunReport failed (offset=%d): %w", offset, err)
		}

		if c.Raw {
			if b, e := json.Marshal(resp); e == nil {
				c.appendRaw(b)
			}
		}

		if len(resp.Rows) == 0 {
			break
		}

		page, err := parsePagePath(c.propertyID, resp)
		if err != nil {
			return nil, err
		}
		report.PagePaths = append(report.PagePaths, page.PagePaths...)

		offset += int64(len(resp.Rows))
		if offset >= resp.RowCount || int64(len(report.PagePaths)) >= limit {
			break
		}
	}

	return report, nil
}
