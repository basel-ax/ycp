package ui

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/Zebbeni/ansipx"
	"golang.org/x/term"
)

// DisplayHomeScreen displays the home screen with parameters
func DisplayHomeScreen(totalLimit, timeLimit, redisCount int, finalComment, apiConnection string) {
	fmt.Println("=== YouTube Stream Comments Processor ===")
	fmt.Println("Parameters:")
	fmt.Printf("Total Limit: %d\n", totalLimit)
	fmt.Printf("Time Limit: %d seconds\n", timeLimit)
	fmt.Printf("Final Comment: %s\n", finalComment)
	fmt.Printf("API Connection: %s\n", apiConnection)
	fmt.Printf("Redis Count: %d\n", redisCount)
	fmt.Println("Press Enter to clear the console and start reading comments...")
}

// DisplayFinalScreen displays the final statistics screen with an ANSI graphic
// on the left 2/3 and statistics on the right 1/3.
func DisplayFinalScreen(commentsRead, lettersTyped, commandsSent int, triggeredLetters []string) {
	fd := int(os.Stdout.Fd())
	width, _, err := term.GetSize(fd)
	if err != nil || width < 60 {
		width = 120
	}

	// If no graphics directory or no files, fall back to text
	graphicsDir := "graphics"
	entries, err := filepath.Glob(filepath.Join(graphicsDir, "*"))
	if err != nil || len(entries) == 0 {
		displayFinalText(commentsRead, lettersTyped, commandsSent, triggeredLetters)
		return
	}

	imgPath := entries[rand.Intn(len(entries))]

	// Image occupies left 2/3, stats the right 1/3
	imgWidth := width * 2 / 3
	if imgWidth < 20 {
		imgWidth = 20
	}

	opts := ansipx.DefaultOptions()
	opts.Width = imgWidth

	result, err := ansipx.RenderFile(imgPath, opts)
	if err != nil {
		displayFinalText(commentsRead, lettersTyped, commandsSent, triggeredLetters)
		return
	}

	// Print the ANSI image
	fmt.Print(result)

	// Build statistics text
	statsStr := formatStats(commentsRead, lettersTyped, commandsSent, triggeredLetters)
	statsLines := strings.Split(strings.TrimRight(statsStr, "\n"), "\n")

	// Count rows the rendered image occupies
	imgLines := strings.Split(strings.TrimRight(result, "\n"), "\n")
	imgRows := len(imgLines)

	// Reset colors then overlay stats on the right side via CUP (cursor positioning)
	rightCol := imgWidth + 2 // 2-character gap
	fmt.Print("\033[0m")

	for i, line := range statsLines {
		row := i + 1
		fmt.Printf("\033[%d;%dH%s", row, rightCol, line)
	}

	// If the image has more rows than stats, ensure the stats column is blank
	// for the remaining rows
	if imgRows > len(statsLines) {
		for i := len(statsLines); i < imgRows; i++ {
			fmt.Printf("\033[%d;%dH\033[0m", i+1, rightCol)
		}
	}

	// Position cursor below both areas
	bottomRow := imgRows
	if len(statsLines) > bottomRow {
		bottomRow = len(statsLines)
	}
	bottomRow += 2
	fmt.Printf("\033[%d;1H", bottomRow)
}

// formatStats builds the statistics text block.
func formatStats(commentsRead, lettersTyped, commandsSent int, triggeredLetters []string) string {
	var b strings.Builder
	b.WriteString("=== Statistics ===\n")
	b.WriteString(fmt.Sprintf("Comments Read: %d\n", commentsRead))
	b.WriteString(fmt.Sprintf("Letters Typed: %d\n", lettersTyped))
	b.WriteString(fmt.Sprintf("Commands Sent: %d\n", commandsSent))
	if len(triggeredLetters) > 0 {
		sorted := make([]string, len(triggeredLetters))
		copy(sorted, triggeredLetters)
		sort.Strings(sorted)
		b.WriteString("Triggered Letters:\n")
		for _, l := range sorted {
			b.WriteString(fmt.Sprintf("  %s\n", l))
		}
	}
	return b.String()
}

// displayFinalText is the text-only fallback when ANSI rendering is unavailable.
func displayFinalText(commentsRead, lettersTyped, commandsSent int, triggeredLetters []string) {
	fmt.Print("\033[0m")
	fmt.Print(formatStats(commentsRead, lettersTyped, commandsSent, triggeredLetters))
}

// ClearConsole clears the console
func ClearConsole() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "error clearing console: %v\n", err)
		}
	}
}
