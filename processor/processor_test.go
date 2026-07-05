package processor

import (
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/basel-ax/ycp/config"
	ycpredis "github.com/basel-ax/ycp/redis"
)

type mockStats struct {
	commentsRead     int
	lettersTyped     int
	commandsSent     int
	triggeredLetters []string
}

func (m *mockStats) RecordComment() { m.commentsRead++ }
func (m *mockStats) RecordLetter()  { m.lettersTyped++ }
func (m *mockStats) RecordCommand() { m.commandsSent++ }
func (m *mockStats) RecordTrigger(letter string) {
	m.triggeredLetters = append(m.triggeredLetters, letter)
}

func TestMotivatingMessageShownWhenCountNotReached(t *testing.T) {
	cfg := &config.Config{
		FinalComment: "what the fuck? help me! i am trapped inside a computer.",
		RedisCount:   5,
		TotalLimit:   100,
	}

	server := miniredis.RunT(t)
	defer server.Close()

	redisClient, err := ycpredis.NewRedisClient("localhost", server.Port(), "", 0)
	if err != nil {
		t.Fatalf("error creating Redis client: %v", err)
	}
	defer redisClient.Close()

	stats := &mockStats{}
	p := New(cfg, stats, redisClient, nil, false)

	_ = p.Process("??")

	if stats.lettersTyped != 1 {
		t.Errorf("expected LettersTyped to be 1, got %d", stats.lettersTyped)
	}
	if stats.commandsSent != 1 {
		t.Errorf("expected CommandsSent to be 1, got %d", stats.commandsSent)
	}
}

func TestMotivatingMessagesArrayLength(t *testing.T) {
	if len(motivatingMessages) != 10 {
		t.Errorf("expected 10 motivating messages, got %d", len(motivatingMessages))
	}
}

func TestMotivatingMessagesNotEmpty(t *testing.T) {
	for i, msg := range motivatingMessages {
		if strings.TrimSpace(msg) == "" {
			t.Errorf("motivating message at index %d is empty", i)
		}
	}
}
