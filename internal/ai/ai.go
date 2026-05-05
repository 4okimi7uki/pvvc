package ai

import (
	"context"

	"github.com/4okimi7uki/pvvc/internal/report"
)

type Analyzer interface {
	Analyze(ctx context.Context, req []report.DailyReport, update func(string)) (string, error)
}
