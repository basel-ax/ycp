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
	width, height, err := term.GetSize(fd)
	if err != nil || width < 60 {
		width = 120
		height = 40
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

	// Build statistics text
	statsStr := formatStats(commentsRead, lettersTyped, commandsSent, triggeredLetters)
	statsLines := strings.Split(strings.TrimRight(statsStr, "\n"), "\n")

	imgLines := strings.Split(strings.TrimRight(result, "\n"), "\n")

	fmt.Print("\033[2J\033[H")
	fmt.Print("\033[0m")

	rightCol := imgWidth + 2

	maxRows := max(len(imgLines), len(statsLines))
	if maxRows > height {
		maxRows = height
	}

	for i := 0; i < maxRows; i++ {
		fmt.Print("\033[2K")

		if i < len(imgLines) {
			fmt.Print(imgLines[i])
			fmt.Print("\033[0m")
		}

		if i < len(statsLines) {
			fmt.Printf("\033[%dG%s", rightCol, statsLines[i])
		}

		fmt.Print("\r\n")
	}

	fmt.Print("\r\n")
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
		b.WriteString(fmt.Sprintf("Triggered: %s\n", strings.Join(sorted, "")))
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
