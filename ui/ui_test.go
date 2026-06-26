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

	if !strings.Contains(out, "=== Statistics ===") {
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

	if !strings.Contains(out, "=== Statistics ===") {
		t.Errorf("expected header in final screen, got:\n%s", out)
	}
	if !strings.Contains(out, "Triggered Letters:") {
		t.Errorf("expected triggered letters section in output, got:\n%s", out)
	}
	// Triggered letters should be sorted: !, ?, t
	if !strings.Contains(out, "  !") {
		t.Errorf("expected sorted triggered letter '!' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "  ?") {
		t.Errorf("expected sorted triggered letter '?' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "  t") {
		t.Errorf("expected sorted triggered letter 't' in output, got:\n%s", out)
	}
}

// Test that ClearConsole does not panic
func TestClearConsole_NoPanic(t *testing.T) {
	// Simply call the function; if it panics the test will fail
	ClearConsole()
}

// Test that formatStats returns the expected text
func TestFormatStats(t *testing.T) {
	result := formatStats(5, 10, 3, []string{"a", "c", "b"})

	if !strings.Contains(result, "=== Statistics ===") {
		t.Errorf("expected '=== Statistics ===', got:\n%s", result)
	}
	if !strings.Contains(result, "Comments Read: 5") {
		t.Errorf("expected 'Comments Read: 5', got:\n%s", result)
	}
	if !strings.Contains(result, "Letters Typed: 10") {
		t.Errorf("expected 'Letters Typed: 10', got:\n%s", result)
	}
	if !strings.Contains(result, "Commands Sent: 3") {
		t.Errorf("expected 'Commands Sent: 3', got:\n%s", result)
	}
	if !strings.Contains(result, "Triggered Letters:") {
		t.Errorf("expected triggered letters header, got:\n%s", result)
	}

	// Verify sorted order: a, b, c
	if !strings.Contains(result, "  a\n  b\n  c") {
		t.Errorf("expected triggered letters sorted as 'a, b, c', got:\n%s", result)
	}
}

// Test that formatStats handles empty triggered letters
func TestFormatStats_NoTriggered(t *testing.T) {
	result := formatStats(1, 2, 3, nil)

	if strings.Contains(result, "Triggered") {
		t.Errorf("expected no triggered letters section, got:\n%s", result)
	}
}
