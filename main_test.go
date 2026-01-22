package main

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/basel-ax/ycp/redis"
)

// TestLoadConfig tests the loadConfig function
func TestLoadConfig(t *testing.T) {
	config, err := loadConfig("example.env")
	if err != nil {
		t.Fatalf("Error loading config: %v", err)
	}

	if config.TotalLimit != 100 {
		t.Errorf("Expected TotalLimit to be 100, got %d", config.TotalLimit)
	}

	if config.TimeLimit != 3600 {
		t.Errorf("Expected TimeLimit to be 3600, got %d", config.TimeLimit)
	}

	if config.FinalComment != "exit" {
		t.Errorf("Expected FinalComment to be 'exit', got %s", config.FinalComment)
	}

	if config.RedisCount != 5 {
		t.Errorf("Expected RedisCount to be 5, got %d", config.RedisCount)
	}
}

// TestProcessComment tests the processComment function
func TestProcessComment(t *testing.T) {
	config, err := loadConfig("example.env")
	if err != nil {
		t.Fatalf("Error loading config: %v", err)
	}

	stats := &Stats{}
	logger, err := NewLogger("test_comments.log")
	if err != nil {
		t.Fatalf("Error initializing logger: %v", err)
	}
	defer logger.Close()

	// Create a mini Redis server for testing
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Error starting mini Redis server: %v", err)
	}
	defer server.Close()

	// Create a Redis client using the custom RedisClient
	redisClient, err := redis.NewRedisClient("localhost", server.Port(), "", 0)
	if err != nil {
		t.Fatalf("Error creating Redis client: %v", err)
	}
	defer redisClient.Close()

	// Test with a comment containing double letters in FINAL_COMMENT
	comment := "ww"
	shouldTerminate := processComment(comment, config, stats, logger, redisClient, false)
	if shouldTerminate {
		t.Errorf("Expected shouldTerminate to be false, got true")
	}

	if stats.CommentsRead != 1 {
		t.Errorf("Expected CommentsRead to be 1, got %d", stats.CommentsRead)
	}

	// Test with FINAL_COMMENT
	comment = "exit"
	shouldTerminate = processComment(comment, config, stats, logger, redisClient, false)
	if !shouldTerminate {
		t.Errorf("Expected shouldTerminate to be true, got false")
	}
}

// TestRedisIntegration tests the Redis integration
func TestRedisIntegration(t *testing.T) {
	// Create a mini Redis server for testing
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Error starting mini Redis server: %v", err)
	}
	defer server.Close()

	// Create a Redis client using the custom RedisClient
	redisClient, err := redis.NewRedisClient("localhost", server.Port(), "", 0)
	if err != nil {
		t.Fatalf("Error creating Redis client: %v", err)
	}
	defer redisClient.Close()

	// Test incrementing button count
	err = redisClient.IncrementButtonCount("w")
	if err != nil {
		t.Fatalf("Error incrementing button count: %v", err)
	}

	count, err := redisClient.GetButtonCount("w")
	if err != nil {
		t.Fatalf("Error getting button count: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected button count to be 1, got %d", count)
	}

	// Test incrementing total commands
	err = redisClient.IncrementTotalCommands()
	if err != nil {
		t.Fatalf("Error incrementing total commands: %v", err)
	}

	totalCommands, err := redisClient.GetTotalCommands()
	if err != nil {
		t.Fatalf("Error getting total commands: %v", err)
	}

	if totalCommands != 1 {
		t.Errorf("Expected total commands to be 1, got %d", totalCommands)
	}
}

// TestReadComments tests the readComments function
func TestReadComments(t *testing.T) {
	comments := readComments()
	timeout := time.After(5 * time.Second)
	count := 0

	for {
		select {
		case <-comments:
			count++
		case <-timeout:
			if count == 0 {
				t.Error("No comments received")
			}
			return
		}
	}
}
