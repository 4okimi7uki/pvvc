package ui

import (
	"fmt"
	"strings"
	"time"
)

func PrintLogo() {
	const logo = `
 ██████╗ ██╗   ██╗ ██╗   ██╗   ██████╗
 ██╔══██╗██║   ██║ ██║   ██║  ██╔════╝
 ██████╔╝██║   ██║ ██║   ██║  ██║
 ██╔═══╝ ╚██╗ ██╔╝ ╚██╗ ██╔╝  ██║
 ██║██╗   ╚████╔╝██╗╚████╔╝██╗╚██████╗
 ╚═╝╚═╝    ╚═══╝ ╚═╝ ╚═══╝ ╚═╝ ╚═════╝
                                     `
	const (
		description = "Page Views Vercel Cost"
		tagline     = "Compare Vercel spend with GA4 traffic :P"
		repoURL     = "https://github.com/4okimi7uki/pvvc"
	)
	var startTime = fmt.Sprintf(" %s %s\n", DeepTeal("▶︎")+" Started at", time.Now().Format("2006-01-02 15:04:05"))

	width := max(len(tagline), len(repoURL)) + 3
	upperBar := strings.Repeat(".", width)
	belowBar := strings.Repeat("·", width)

	fmt.Println(Mastered(logo))
	items := []string{" " + Bold(description), " " + tagline, upperBar, " " + repoURL, belowBar}
	for _, item := range items {
		fmt.Println(Mastered(item))

	}
	fmt.Println()
	fmt.Print(startTime)
	fmt.Println()
}
