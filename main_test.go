package main

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/basel-ax/ycp/config"
	"github.com/basel-ax/ycp/redis"
)

func TestLoadConfig(t *testing.T) {
	cfg, err := config.LoadConfig("example.env")
	if err != nil {
		t.Fatalf("Error loading config: %v", err)
	}

	if cfg.TotalLimit != 100 {
		t.Errorf("Expected TotalLimit to be 100, got %d", cfg.TotalLimit)
	}

	if cfg.TimeLimit != 3600 {
		t.Errorf("Expected TimeLimit to be 3600, got %d", cfg.TimeLimit)
	}

	if cfg.FinalComment != "exit" {
		t.Errorf("Expected FinalComment to be 'exit', got %s", cfg.FinalComment)
	}

	if cfg.RedisCount != 5 {
		t.Errorf("Expected RedisCount to be 5, got %d", cfg.RedisCount)
	}
}

func TestProcessComment(t *testing.T) {
	cfg, err := config.LoadConfig("example.env")
	if err != nil {
		t.Fatalf("Error loading config: %v", err)
	}

	stats := &Stats{}
	logger, err := NewLogger("test_comments.log")
	if err != nil {
		t.Fatalf("Error initializing logger: %v", err)
	}
	defer logger.Close()

	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Error starting mini Redis server: %v", err)
	}
	defer server.Close()

	redisClient, err := redis.NewRedisClient("localhost", server.Port(), "", 0)
	if err != nil {
		t.Fatalf("Error creating Redis client: %v", err)
	}
	defer redisClient.Close()

	comment := "ww"
	shouldTerminate := processComment(comment, cfg, stats, logger, redisClient, false)
	if shouldTerminate {
		t.Errorf("Expected shouldTerminate to be false, got true")
	}

	if stats.CommentsRead != 1 {
		t.Errorf("Expected CommentsRead to be 1, got %d", stats.CommentsRead)
	}

	comment = "exit"
	shouldTerminate = processComment(comment, cfg, stats, logger, redisClient, false)
	if !shouldTerminate {
		t.Errorf("Expected shouldTerminate to be true, got false")
	}
}

func TestReadComments(t *testing.T) {
	cfg, err := config.LoadConfig("example.env")
	if err != nil {
		t.Fatalf("Error loading config: %v", err)
	}

	comments := readComments(cfg)
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
