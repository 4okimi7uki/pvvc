package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/4okimi7uki/pvvc/internal/ui"
)

const barWidth = 100

func printTable(rows [][]string) {
	colWidths := make([]int, len(rows[0]))

	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell) + 3
			}
		}
	}

	for _, row := range rows {
		fmt.Print(" ")
		for i, cell := range row {
			fmt.Printf("%-*s  ", colWidths[i], cell)
		}
		fmt.Println()
	}
	fmt.Println()
}

func PrintSection(label string) {
	line := strings.Repeat(ui.MossGray("─"), barWidth-len(label))
	fmt.Printf("\n%s %s\n", label, line)
}

func PrintSomeDayReports(start, end time.Time, reports []DailyReport, aiResponse string, llm string) {
	var (
		allPv, allCost, metricsRows = metrics(reports)
		summaryRows                 = summary(start, end, reports, llm, allPv, allCost)
		costByService               = LatestServiceCosts(end, reports)
	)

	PrintSection("Summary")
	fmt.Println()
	for _, s := range summaryRows {
		fmt.Printf(" %-10s %s\n", s.Label, s.Value)
	}

	PrintSection("Metrics")
	fmt.Println()
	printTable(metricsRows)

	PrintSection("Service Costs on Latest Date")
	fmt.Println()
	for _, s := range costByService {
		fmt.Printf(" %-10s %s\n", s.Label, s.Value)
	}

	if aiResponse != "" {
		PrintSection("AI Analytics")
		fmt.Println()
		fmt.Println(aiResponse)
	}
}
