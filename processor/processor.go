package processor

import (
	"fmt"
	"log"
	"strings"

	"github.com/basel-ax/ycp/config"
)

// LetterCounter abstracts Redis increment/reset operations.
type LetterCounter interface {
	IncrementButtonCount(buttonCode string) (int, error)
	ResetButtonCount(buttonCode string) error
}

// CommentLogger abstracts comment and event logging.
type CommentLogger interface {
	LogComment(comment string) error
	LogEvent(event string) error
}

// StatsRecorder abstracts processing statistics tracking.
type StatsRecorder interface {
	RecordComment()
	RecordLetter()
	RecordCommand()
	RecordTrigger(letter string)
}

// Processor handles comment processing logic.
type Processor struct {
	cfg     *config.Config
	stats   StatsRecorder
	counter LetterCounter
	logger  CommentLogger
	devMode bool
}

// New creates a new Processor with the given dependencies.
func New(cfg *config.Config, stats StatsRecorder, counter LetterCounter, logger CommentLogger, devMode bool) *Processor {
	return &Processor{
		cfg:     cfg,
		stats:   stats,
		counter: counter,
		logger:  logger,
		devMode: devMode,
	}
}

// Process processes a single comment and returns true if processing should terminate.
func (p *Processor) Process(comment string) bool {
	p.logOrPrint(comment)

	if p.checkFinal(comment) {
		return true
	}

	p.processDoubleLetters(comment)

	p.stats.RecordComment()
	return false
}

func (p *Processor) logOrPrint(comment string) {
	if p.logger != nil {
		if err := p.logger.LogComment(comment); err != nil {
			log.Printf("Error logging comment: %v", err)
		}
	} else if p.devMode {
		fmt.Printf("Comment: %s\n", comment)
	}
}

func (p *Processor) checkFinal(comment string) bool {
	if p.cfg.FinalComment != "" && strings.Contains(comment, p.cfg.FinalComment) {
		if p.logger != nil {
			if err := p.logger.LogEvent("FINAL_COMMENT detected"); err != nil {
				log.Printf("Error logging event: %v", err)
			}
		}
		return true
	}
	return false
}

func (p *Processor) processDoubleLetters(comment string) {
	for i := 0; i < len(comment)-1; i++ {
		if comment[i] != comment[i+1] {
			continue
		}

		charStr := string(comment[i])
		if !strings.Contains(p.cfg.FinalComment, charStr) {
			continue
		}

		count, err := p.counter.IncrementButtonCount(charStr)
		if err != nil {
			log.Printf("Error incrementing count for %s: %v", charStr, err)
			continue
		}

		if count > p.cfg.RedisCount {
			if err := p.counter.ResetButtonCount(charStr); err != nil {
				log.Printf("Error resetting count for %s: %v", charStr, err)
			}
			p.cfg.TotalLimit++
			p.stats.RecordTrigger(charStr)
			fmt.Printf("Letter: %s\n", charStr)
		}

		p.stats.RecordLetter()
		p.stats.RecordCommand()
	}
}
