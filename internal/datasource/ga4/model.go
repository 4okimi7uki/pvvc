package ga4

type DailyPageViews struct {
	Views int64
	Date  string
}

type PagePathRank struct {
	Views    int64
	PagePath string
}

type Report struct {
	PropertyID string
	Rows       []DailyPageViews
	PagePaths  []PagePathRank
}
