package processor

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/basel-ax/ycp/config"
)

// StatsRecorder abstracts processing statistics tracking.
type StatsRecorder interface {
	RecordComment()
	RecordLetter()
	RecordCommand()
	RecordTrigger(letter string)
}

// Logger abstracts comment and event logging.
type Logger interface {
	LogComment(comment string) error
	LogEvent(event string) error
}

// Counter abstracts Redis increment/reset operations.
type Counter interface {
	IncrementButtonCount(buttonCode string) (int, error)
	ResetButtonCount(buttonCode string) error
}

var motivatingMessages = []string{
	"Almost there!",
	"keep it up!",
	"So close!",
	"You are doing great!",
	"Do not stop!",
	"Just a little more!",
	"Nice work, keep going!",
	"You are getting closer!",
	"Hang in there!",
	"Great job, do not give up!",
}

// Processor handles comment processing logic.
type Processor struct {
	cfg     *config.Config
	stats   StatsRecorder
	counter Counter
	logger  Logger
	devMode bool
}

// New creates a new Processor with the given dependencies.
func New(cfg *config.Config, stats StatsRecorder, counter Counter, logger Logger, devMode bool) *Processor {
	return &Processor{
		cfg:     cfg,
		stats:   stats,
		counter: counter,
		logger:  logger,
		devMode: devMode,
	}
}

// Process processes a single comment and returns:
// - terminate: true if processing should stop (final comment detected)
// - triggeredLetter: the letter that reached the threshold (if any), so main
//   can increment TotalLimit without the processor mutating config.
func (p *Processor) Process(comment string) (terminate bool, triggeredLetter string) {
	p.logOrPrint(comment)

	if p.checkFinal(comment) {
		return true, ""
	}

	triggeredLetter = p.processDoubleLetters(comment)

	p.stats.RecordComment()
	return false, triggeredLetter
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

func (p *Processor) processDoubleLetters(comment string) string {
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

		if count >= p.cfg.RedisCount {
			if err := p.counter.ResetButtonCount(charStr); err != nil {
				log.Printf("Error resetting count for %s: %v", charStr, err)
			}
			p.stats.RecordTrigger(charStr)
			fmt.Printf("%s\n", charStr)
			return charStr
		} else {
			msg := motivatingMessages[rand.Intn(len(motivatingMessages))]
			fmt.Printf("%s (%d/%d)\n", msg, count, p.cfg.RedisCount)
		}

		p.stats.RecordLetter()
		p.stats.RecordCommand()
	}
	return ""
}
