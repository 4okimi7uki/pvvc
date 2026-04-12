package ga4

import "time"

type DailyPageViews struct {
	PagePath string // Note: 一旦空の値が入るようにしている
	Views    int64
	Date     time.Time
}

type Report struct {
	PropertyID string
	Rows       []DailyPageViews
}
