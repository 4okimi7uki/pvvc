package ai

type ReportRow struct {
	Line string
}

type PromptData struct {
	ServiceName    string
	Today          string
	TableHeader    string
	Rows           []ReportRow
	NewsURLs       []string
	IsBeforeCutoff bool
	HasAnomaly     bool
}
