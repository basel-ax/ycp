package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	TotalLimit    int
	TimeLimit     int
	FinalComment  string
	APIConnection string
	StreamURL     string
	YouTubeAPIKey string
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	RedisCount    int
}

func LoadConfig(filePath string) (*Config, error) {
	config := &Config{}

	if err := loadEnvFile(filePath); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	config.TotalLimit = getEnvAsInt("TOTAL_LIMIT", 100)
	config.TimeLimit = getEnvAsInt("TIME_LIMIT", 3600)
	config.FinalComment = os.Getenv("FINAL_COMMENT")
	config.APIConnection = os.Getenv("API_CONNECTION")
	config.StreamURL = os.Getenv("STREAM_URL")
	config.YouTubeAPIKey = os.Getenv("YOUTUBE_API_KEY")
	config.RedisHost = os.Getenv("REDIS_HOST")
	config.RedisPort = os.Getenv("REDIS_PORT")
	config.RedisPassword = os.Getenv("REDIS_PASSWORD")
	config.RedisDB = getEnvAsInt("REDIS_DB", 0)
	config.RedisCount = getEnvAsInt("REDIS_COUNT", 5)

	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

func validateConfig(config *Config) error {
	if config.TotalLimit <= 0 {
		return fmt.Errorf("TOTAL_LIMIT must be greater than 0, got %d", config.TotalLimit)
	}
	if config.TimeLimit <= 0 {
		return fmt.Errorf("TIME_LIMIT must be greater than 0, got %d", config.TimeLimit)
	}
	if config.RedisCount <= 0 {
		return fmt.Errorf("REDIS_COUNT must be greater than 0, got %d", config.RedisCount)
	}
	return nil
}

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

func loadEnvFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"")
		value = strings.Trim(value, "'")

		os.Setenv(key, value)
	}

	return nil
}
