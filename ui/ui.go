package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
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

// DisplayFinalScreen displays the final statistics screen
func DisplayFinalScreen(commentsRead, lettersTyped, commandsSent int) {
	fmt.Println("=== Final Statistics ===")
	fmt.Printf("Comments Read: %d\n", commentsRead)
	fmt.Printf("Letters Typed: %d\n", lettersTyped)
	fmt.Printf("Commands Sent: %d\n", commandsSent)
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
		cmd.Run()
	}
}
