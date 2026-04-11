package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

const barWith = 32

func printTable(rows [][]string) {
	colWidths := make([]int, len(rows[0]))

	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell) + 4
			}
		}
	}

	for _, row := range rows {
		fmt.Print(" ")
		for i, cell := range row {
			fmt.Printf("%-*s", colWidths[i], cell)
		}
		fmt.Println()
	}
	fmt.Println()
}

func printSection(label string) {
	line := strings.Repeat("─", barWith-len(label))
	fmt.Printf("\n%s %s\n", label, line)
}

func PrintReport(targetDay time.Time, _pv int64, _totalCost, _totalCostJPY, _rate float64) {
	_costPerPVUSD := _totalCost / float64(_pv)
	_costPerPVJPY := _totalCostJPY / float64(_pv)

	pv := humanize.Comma(_pv)
	totalCost := humanize.CommafWithDigits(_totalCost, 4)
	totalCostJPY := humanize.CommafWithDigits(_totalCostJPY, 2)
	costPerPVUSD := humanize.CommafWithDigits(_costPerPVUSD, 6)
	costPerPVJPY := humanize.CommafWithDigits(_costPerPVJPY, 4)
	rate := humanize.CommafWithDigits(_rate, 2)

	type row struct {
		label string
		value string
	}
	summaryRows := []row{
		{"Period", targetDay.Format("2006/01/02")},
		{"PV", fmt.Sprintf("%s", pv)},
		{"Rate", fmt.Sprintf("$1 = ¥%s", rate)},
	}

	printSection("Summary")
	fmt.Println()
	for _, s := range summaryRows {
		fmt.Printf(" %-8s %s\n", s.label, s.value)
	}

	printSection("Cost")
	fmt.Println()
	printTable([][]string{
		{"", "USD", "JPY"},
		{"Total", fmt.Sprintf("%s", totalCost), fmt.Sprintf("%s", totalCostJPY)},
		{"/ PV", fmt.Sprintf("%s", costPerPVUSD), fmt.Sprintf("%s", costPerPVJPY)},
	})

	fmt.Println(strings.Repeat("─", barWith))
}
