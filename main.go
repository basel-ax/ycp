package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/basel-ax/ycp/redis"
	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	TotalLimit    int
	TimeLimit     int
	FinalComment  string
	APIConnection string
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	RedisCount    int
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
	config := &Config{}

	// Load the .env file
	if err := godotenv.Load(filePath); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
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
	config.RedisCount = getEnvAsInt("REDIS_COUNT", 5)

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
	fmt.Println("Parameters:")
	fmt.Printf("Total Limit: %d\n", config.TotalLimit)
	fmt.Printf("Time Limit: %d seconds\n", config.TimeLimit)
	fmt.Printf("Final Comment: %s\n", config.FinalComment)
	fmt.Printf("API Connection: %s\n", config.APIConnection)
	fmt.Printf("Redis Count: %d\n", config.RedisCount)
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
func processComment(comment string, config *Config, stats *Stats, logger *Logger, redisClient *redis.RedisClient, devMode bool) bool {
	// Log the comment
	if logger != nil {
		if err := logger.LogComment(comment); err != nil {
			log.Printf("Error logging comment: %v", err)
		}
	} else if devMode {
		fmt.Printf("Comment: %s\n", comment)
	}

	// Check for FINAL_COMMENT
	if config.FinalComment != "" && strings.Contains(comment, config.FinalComment) {
		if logger != nil {
			if err := logger.LogEvent("FINAL_COMMENT detected"); err != nil {
				log.Printf("Error logging event: %v", err)
			}
		}
		return true
	}

	// Check for double same letter or symbol
	for _, char := range comment {
		charStr := string(char)
		if strings.Count(comment, charStr) >= 2 {
			// Check if FINAL_COMMENT contains this letter/symbol
			if strings.Contains(config.FinalComment, charStr) {
				// Increment count in Redis
				if err := redisClient.IncrementButtonCount(charStr); err != nil {
					log.Printf("Error incrementing count for %s: %v", charStr, err)
					continue
				}

				// Get current count
				count, err := redisClient.GetButtonCount(charStr)
				if err != nil {
					log.Printf("Error getting count for %s: %v", charStr, err)
					continue
				}

				// Check if count > REDIS_COUNT
				if count > config.RedisCount {
					// Reset count to 0
					if err := redisClient.ResetButtonCount(charStr); err != nil {
						log.Printf("Error resetting count for %s: %v", charStr, err)
					}
					// Increase total limit
					config.TotalLimit++
					// Print the letter
					fmt.Printf("Letter: %s\n", charStr)
				}

				stats.LettersTyped++
				stats.CommandsSent++
			}
		}
	}

	stats.CommentsRead++
	return false
}

// main is the entry point of the application
func main() {
	devMode := flag.Bool("dev", false, "Enable development mode (print comments to console)")
	flag.Parse()

	// Load configuration
	config, err := loadConfig("example.env")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize Redis client
	redisClient, err := redis.NewRedisClient(config.RedisHost, config.RedisPort, config.RedisPassword, config.RedisDB)
	if err != nil {
		log.Fatalf("Error initializing Redis client: %v", err)
	}
	defer redisClient.Close()

	// Initialize logger
	var logger *Logger
	if *devMode {
		logger = nil // In dev mode, no file logging
	} else {
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		logFileName := fmt.Sprintf("comments_%s.log", timestamp)
		logger, err = NewLogger(logFileName)
		if err != nil {
			log.Fatalf("Error initializing logger: %v", err)
		}
		defer logger.Close()
	}

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
				if processComment(comment, config, stats, logger, redisClient, *devMode) {
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
