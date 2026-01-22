package config

import (
	"fmt"
	"os"

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

// LoadConfig loads the configuration from the .env file
func LoadConfig(filePath string) (*Config, error) {
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
