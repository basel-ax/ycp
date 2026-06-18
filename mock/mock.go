package mock

import (
	"context"
	"time"
)

// Source generates mock comments for development and testing.
type Source struct{}

// NewSource creates a new mock comment source.
func NewSource() *Source {
	return &Source{}
}

// Read returns a channel of mock comments for testing.
func (s *Source) Read(ctx context.Context) <-chan string {
	comments := make(chan string)
	go func() {
		defer close(comments)
		mockComments := []string{
			"what the fuck? help me! i am trapped inside a computer.",
			"ww", "hh", "aa", "tt", "ww", "hh", "aa", "tt",
			"exit",
		}
		for _, c := range mockComments {
			select {
			case comments <- c:
			case <-ctx.Done():
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()
	return comments
}
