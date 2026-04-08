package ga4

type DailyPageViews struct {
	PagePath string
	Views    int64
}

type Report struct {
	PropertyID string
	Rows       []DailyPageViews
}
