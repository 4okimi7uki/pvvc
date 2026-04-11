package ga4

type DailyPageViews struct {
	PagePath string
	Views    int64
	Date     string
}

type Report struct {
	PropertyID string
	Rows       []DailyPageViews
}
