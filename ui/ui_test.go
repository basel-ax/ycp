package ui

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// Test that DisplayHomeScreen prints the expected formatting and values
func TestDisplayHomeScreen_Output(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	// Ensure we restore stdout after test
	defer func() {
		os.Stdout = old
	}()

	// Call the function with sample values
	DisplayHomeScreen(100, 3600, 5, "exit", "https://api.local")
	// Finish writing to the pipe and read output
	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	out := buf.String()

	// Validate key substrings are present
	if !strings.Contains(out, "Total Limit: 100") {
		t.Errorf("expected output to contain 'Total Limit: 100', got:\n%s", out)
	}
	if !strings.Contains(out, "Time Limit: 3600 seconds") {
		t.Errorf("expected output to contain 'Time Limit: 3600 seconds', got:\n%s", out)
	}
	if !strings.Contains(out, "Final Comment: exit") {
		t.Errorf("expected output to contain 'Final Comment: exit', got:\n%s", out)
	}
	if !strings.Contains(out, "Redis Count: 5") {
		t.Errorf("expected output to contain 'Redis Count: 5', got:\n%s", out)
	}
	if !strings.Contains(out, "Press Enter to clear the console") {
		t.Errorf("expected output to indicate console will clear, got:\n%s", out)
	}
}

// Test that DisplayFinalScreen prints the expected final statistics
func TestDisplayFinalScreen_Output(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = old }()

	DisplayFinalScreen(12, 34, 56, nil)
	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	out := buf.String()

	if !strings.Contains(out, "=== Final Statistics ===") {
		t.Errorf("expected header in final screen, got:\n%s", out)
	}
	if !strings.Contains(out, "Comments Read: 12") {
		t.Errorf("expected final screen to include 'Comments Read: 12', got:\n%s", out)
	}
	if !strings.Contains(out, "Letters Typed: 34") {
		t.Errorf("expected final screen to include 'Letters Typed: 34', got:\n%s", out)
	}
	if !strings.Contains(out, "Commands Sent: 56") {
		t.Errorf("expected final screen to include 'Commands Sent: 56', got:\n%s", out)
	}
}

// Test that DisplayFinalScreen shows triggered letters
func TestDisplayFinalScreen_WithTriggeredLetters(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = old }()

	DisplayFinalScreen(10, 5, 5, []string{"t", "?", "!"})
	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	out := buf.String()

	if !strings.Contains(out, "Triggered Letters: t, ?, !") {
		t.Errorf("expected triggered letters in output, got:\n%s", out)
	}
}

// Test that ClearConsole does not panic
func TestClearConsole_NoPanic(t *testing.T) {
	// Simply call the function; if it panics the test will fail
	ClearConsole()
}
