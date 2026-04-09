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
		tagline = "Compare Vercel spend with GA4 traffic."
		credit  = "crafted by 4okimi7uki :P"
	)
	width := max(len(tagline), len(credit)) + 3
	upperBar := strings.Repeat(".", width)
	belowBar := strings.Repeat("·", width)

	fmt.Println(Mastered(logo))
	items := []string{upperBar, " " + tagline, " " + credit, belowBar}
	for _, item := range items {
		fmt.Println(Mastered(item))

	}
	fmt.Println()
}
