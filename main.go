package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	ButtonCodes      map[string]string
	TotalLimit       int
	TimeLimit        int
	FinalComment     string
	APIConnection    string
	RedisHost        string
	RedisPort        string
	RedisPassword    string
	RedisDB          int
}

// Stats holds the application statistics
type Stats struct {
	CommentsRead int
	LettersTyped int
	CommandsSent int
}

// Logger handles logging comments and events
type Logger struct {
	file *os.File
}

// NewLogger creates a new Logger instance
func NewLogger(filePath string) (*Logger, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	return &Logger{file: file}, nil
}

// LogComment logs a comment to the file
func (l *Logger) LogComment(comment string) error {
	_, err := l.file.WriteString(comment + "\n")
	return err
}

// LogEvent logs an event to the file
func (l *Logger) LogEvent(event string) error {
	_, err := l.file.WriteString("[EVENT] " + event + "\n")
	return err
}

// Close closes the logger file
func (l *Logger) Close() error {
	return l.file.Close()
}

// loadConfig loads the configuration from the .env file
func loadConfig(filePath string) (*Config, error) {
	config := &Config{
		ButtonCodes: make(map[string]string),
	}

	// Load the .env file
	if err := godotenv.Load(filePath); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	// Load button codes
	buttonCodes := []string{"BUTTON_WW", "BUTTON_HH", "BUTTON_AA", "BUTTON_TT", "BUTTON_SPACE", "BUTTON_DOT", "BUTTON_QUESTION", "BUTTON_EXCLAMATION"}
	for _, code := range buttonCodes {
		if value, exists := os.LookupEnv(code); exists {
			config.ButtonCodes[code] = value
		}
	}

	// Load other configurations
	config.TotalLimit = getEnvAsInt("TOTAL_LIMIT", 100)
	config.TimeLimit = getEnvAsInt("TIME_LIMIT", 3600)
	config.FinalComment = os.Getenv("FINAL_COMMENT")
	config.APIConnection = os.Getenv("API_CONNECTION")
	config.RedisHost = os.Getenv("REDIS_HOST")
	config.RedisPort = os.Getenv("REDIS_PORT")
	config.RedisPassword = os.Getenv("REDIS_PASSWORD")
	config.RedisDB = getEnvAsInt("REDIS_DB", 0)

	return config, nil
}

// getEnvAsInt gets an environment variable as an integer
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	var value int
	_, err := fmt.Sscanf(valueStr, "%d", &value)
	if err != nil {
		return defaultValue
	}
	return value
}

// displayHomeScreen displays the home screen with buttons and parameters
func displayHomeScreen(config *Config) {
	fmt.Println("=== YouTube Stream Comments Processor ===")
	fmt.Println("Buttons and Parameters:")
	for code, value := range config.ButtonCodes {
		fmt.Printf("  %s: %s\n", code, value)
	}
	fmt.Printf("Total Limit: %d\n", config.TotalLimit)
	fmt.Printf("Time Limit: %d seconds\n", config.TimeLimit)
	fmt.Printf("Final Comment: %s\n", config.FinalComment)
	fmt.Printf("API Connection: %s\n", config.APIConnection)
	fmt.Println("Press Enter to clear the console and start reading comments...")
}

// displayFinalScreen displays the final statistics screen
func displayFinalScreen(stats *Stats) {
	fmt.Println("=== Final Statistics ===")
	fmt.Printf("Comments Read: %d\n", stats.CommentsRead)
	fmt.Printf("Letters Typed: %d\n", stats.LettersTyped)
	fmt.Printf("Commands Sent: %d\n", stats.CommandsSent)
}

// clearConsole clears the console
func clearConsole() {
	fmt.Print("\033[H\033[2J")
}

// readComments reads comments from the stream (mock implementation)
func readComments() <-chan string {
	comments := make(chan string)
	go func() {
		defer close(comments)
		mockComments := []string{
			"what the fuck? help me! i am trapped inside a computer.",
			"ww", "hh", "aa", "tt", "ww", "hh", "aa", "tt",
			"exit",
		}
		for _, comment := range mockComments {
			comments <- comment
			time.Sleep(1 * time.Second)
		}
	}()
	return comments
}

// processComment processes a comment and updates the stats
func processComment(comment string, config *Config, stats *Stats, logger *Logger) bool {
	// Log the comment
	if err := logger.LogComment(comment); err != nil {
		log.Printf("Error logging comment: %v", err)
	}

	// Check for FINAL_COMMENT
	if config.FinalComment != "" && strings.Contains(comment, config.FinalComment) {
		if err := logger.LogEvent("FINAL_COMMENT detected"); err != nil {
			log.Printf("Error logging event: %v", err)
		}
		return true
	}

	// Process the comment to find button codes
	for code, word := range config.ButtonCodes {
		if strings.Contains(comment, word) {
			stats.LettersTyped++
			stats.CommandsSent++
			fmt.Printf("Button pressed: %s (Word: %s)\n", code, word)
		}
	}

	stats.CommentsRead++
	return false
}

// main is the entry point of the application
func main() {
	// Load configuration
	config, err := loadConfig("example.env")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize logger
	logger, err := NewLogger("comments.log")
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}
	defer logger.Close()

	// Display home screen
	displayHomeScreen(config)

	// Wait for user to press Enter
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	clearConsole()

	// Initialize stats
	stats := &Stats{}

	// Set up signal handling for graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Start reading comments
	comments := readComments()
	done := make(chan bool)

	go func() {
		for comment := range comments {
			select {
			case <-done:
				return
			default:
				if processComment(comment, config, stats, logger) {
					close(done)
					return
				}
				if stats.CommandsSent >= config.TotalLimit {
					close(done)
					return
				}
			}
		}
	}()

	// Wait for completion or signal
	select {
	case <-done:
		fmt.Println("\nProcessing completed.")
	case <-signals:
		fmt.Println("\nReceived interrupt signal. Shutting down...")
	case <-time.After(time.Duration(config.TimeLimit) * time.Second):
		fmt.Println("\nTime limit reached. Shutting down...")
	}

	// Display final screen
	displayFinalScreen(stats)
}