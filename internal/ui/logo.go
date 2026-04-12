package ui

import (
	"fmt"
	"strings"
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
		description = "Page View Vercel Cost"
		tagline     = "Compare Vercel spend with GA4 traffic :P"
		repoURL     = "https://github.com/4okimi7uki/pvvc"
	)
	width := max(len(tagline), len(repoURL)) + 3
	upperBar := strings.Repeat(".", width)
	belowBar := strings.Repeat("·", width)

	fmt.Println(Mastered(logo))
	items := []string{" " + Bold(description), " " + tagline, upperBar, " " + repoURL, belowBar}
	for _, item := range items {
		fmt.Println(Mastered(item))

	}
	fmt.Println()
}
