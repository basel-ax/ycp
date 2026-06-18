package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/basel-ax/ycp/config"
	"github.com/basel-ax/ycp/processor"
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

	expectedFinalComment := "what the fuck? help me! i am trapped inside a computer."
	if cfg.FinalComment != expectedFinalComment {
		t.Errorf("Expected FinalComment to be %q, got %q", expectedFinalComment, cfg.FinalComment)
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
	defer cleanupTestLog()

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

	proc := processor.New(cfg, stats, redisClient, logger, false)

	comment := "ww"
	shouldTerminate := proc.Process(comment)
	if shouldTerminate {
		t.Errorf("Expected shouldTerminate to be false, got true")
	}

	if stats.CommentsRead != 1 {
		t.Errorf("Expected CommentsRead to be 1, got %d", stats.CommentsRead)
	}

	comment = cfg.FinalComment
	shouldTerminate = proc.Process(comment)
	if !shouldTerminate {
		t.Errorf("Expected shouldTerminate to be true for FinalComment comment")
	}
}

func TestReadComments(t *testing.T) {
	cfg, err := config.LoadConfig("example.env")
	if err != nil {
		t.Fatalf("Error loading config: %v", err)
	}

	comments := readComments(context.Background(), cfg)
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

func TestSymbolDoubleLetters(t *testing.T) {
	cfg, err := config.LoadConfig("example.env")
	if err != nil {
		t.Fatalf("Error loading config: %v", err)
	}

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

	tests := []struct {
		name          string
		comment       string
		wantTerminate bool
		wantLetters   int
		wantCommands  int
		wantComments  int
	}{
		{
			name:         "double question marks trigger",
			comment:      "??",
			wantLetters:  1,
			wantCommands: 1,
			wantComments: 1,
		},
		{
			name:         "double exclamation marks trigger",
			comment:      "!!",
			wantLetters:  1,
			wantCommands: 1,
			wantComments: 1,
		},
		{
			name:         "double dots trigger",
			comment:      "..",
			wantLetters:  1,
			wantCommands: 1,
			wantComments: 1,
		},
		{
			name:         "question marks separated by space do NOT trigger",
			comment:      "? ?",
			wantLetters:  0,
			wantCommands: 0,
			wantComments: 1,
		},
		{
			name:         "exclamation marks separated by space do NOT trigger",
			comment:      "! !",
			wantLetters:  0,
			wantCommands: 0,
			wantComments: 1,
		},
		{
			name:         "dots separated by space do NOT trigger",
			comment:      ". .",
			wantLetters:  0,
			wantCommands: 0,
			wantComments: 1,
		},
		{
			name:         "different consecutive symbols do NOT trigger",
			comment:      "?!",
			wantLetters:  0,
			wantCommands: 0,
			wantComments: 1,
		},
		{
			name:         "triple question marks trigger twice",
			comment:      "???",
			wantLetters:  2,
			wantCommands: 2,
			wantComments: 1,
		},
		{
			name:         "single symbol does NOT trigger",
			comment:      "?",
			wantLetters:  0,
			wantCommands: 0,
			wantComments: 1,
		},
		{
			name:         "double comma does NOT trigger (not in FinalComment)",
			comment:      ",,",
			wantLetters:  0,
			wantCommands: 0,
			wantComments: 1,
		},
		{
			name:         "mixed text with double symbols triggers correctly",
			comment:      "What?? Great!! No..",
			wantLetters:  3,
			wantCommands: 3,
			wantComments: 1,
		},
		{
			name:          "final comment string triggers termination",
			comment:       cfg.FinalComment,
			wantTerminate: true,
			wantLetters:   0,
			wantCommands:  0,
			wantComments:  0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stats := &Stats{}
			proc := processor.New(cfg, stats, redisClient, nil, false)
			shouldTerminate := proc.Process(tc.comment)

			if shouldTerminate != tc.wantTerminate {
				t.Errorf("shouldTerminate: got %v, want %v", shouldTerminate, tc.wantTerminate)
			}
			if stats.LettersTyped != tc.wantLetters {
				t.Errorf("LettersTyped: got %d, want %d", stats.LettersTyped, tc.wantLetters)
			}
			if stats.CommandsSent != tc.wantCommands {
				t.Errorf("CommandsSent: got %d, want %d", stats.CommandsSent, tc.wantCommands)
			}
			if stats.CommentsRead != tc.wantComments {
				t.Errorf("CommentsRead: got %d, want %d", stats.CommentsRead, tc.wantComments)
			}
		})
	}
}

func cleanupTestLog() {
	os.Remove("test_comments.log")
}
