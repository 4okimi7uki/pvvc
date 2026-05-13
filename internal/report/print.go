package report

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/4okimi7uki/pvvc/internal/ui"
)

const barWidth = 100

func WriteTable(w io.Writer, rows [][]string) {
	colWidths := make([]int, len(rows[0]))

	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell) + 3
			}
		}
	}

	for _, row := range rows {
		fmt.Fprint(w, " ")
		for i, cell := range row {
			fmt.Fprintf(w, "%-*s  ", colWidths[i], cell)
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w)
}

func RowsToCells(rows []Row) [][]string {
	cells := make([][]string, len(rows))
	for i, r := range rows {
		cells[i] = []string{r.Label, r.Value}
	}
	return cells
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
	WriteTable(os.Stdout, RowsToCells(summaryRows))

	PrintSection("Metrics")
	fmt.Println()
	WriteTable(os.Stdout, metricsRows)

	PrintSection("Service Costs on Latest Date")
	fmt.Println()
	WriteTable(os.Stdout, RowsToCells(costByService))

	if aiResponse != "" {
		PrintSection("AI Analytics")
		fmt.Println()
		fmt.Println(aiResponse)
	}
}
