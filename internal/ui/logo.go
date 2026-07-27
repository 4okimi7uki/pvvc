package ui

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// FprintLogo はロゴの出力先を選べる PrintLogo。
// SVG を標準出力に流すときに stderr へ逃がすために使う。
func FprintLogo(w io.Writer, version string) {
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
	var startTime = fmt.Sprintf(" pvvc/%s | started at %s\n", version, time.Now().Format("2006-01-02 15:04:05 MST"))
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
