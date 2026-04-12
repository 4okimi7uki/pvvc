package ga4

type DailyPageViews struct {
	PagePath string // Note: 一旦空の値が入るようにしている
	Views    int64
	Date     string
}

type Report struct {
	PropertyID string
	Rows       []DailyPageViews
}
