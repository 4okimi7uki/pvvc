package slack

import (
	"github.com/4okimi7uki/pvvc/internal/datasource/ga4"
	rep "github.com/4okimi7uki/pvvc/internal/report"
)

// sampleTopPages は ga4.PagePathRank のサンプル。
var sampleTopPages = []ga4.PagePathRank{
	{PagePath: "/", Views: 5000},
	{PagePath: "/blog", Views: 2500},
	{PagePath: "/about", Views: 1200},
	{PagePath: "/contact", Views: 800},
	{PagePath: "/pricing", Views: 600},
}
var formattedSmampleTopPages = rep.FormatTopPage(sampleTopPages)
