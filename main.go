package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/basel-ax/ycp/config"
	"github.com/basel-ax/ycp/mock"
	"github.com/basel-ax/ycp/processor"
	"github.com/basel-ax/ycp/redis"
	"github.com/basel-ax/ycp/ui"
	"github.com/basel-ax/ycp/youtube"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

type Stats struct {
	CommentsRead     int
	LettersTyped     int
	CommandsSent     int
	TriggeredLetters []string
}

func (s *Stats) RecordComment()         { s.CommentsRead++ }
func (s *Stats) RecordLetter()          { s.LettersTyped++ }
func (s *Stats) RecordCommand()         { s.CommandsSent++ }
func (s *Stats) RecordTrigger(l string) { s.TriggeredLetters = append(s.TriggeredLetters, l) }

// commentSource unifies mock and YouTube comment streams.
type commentSource interface {
	Read(ctx context.Context) <-chan string
}

type mockSource struct{}

func (m *mockSource) Read(ctx context.Context) <-chan string {
	return mock.NewSource().Read(ctx)
}

type youtubeSource struct {
	videoID string
	apiKey  string
}

func (y *youtubeSource) Read(ctx context.Context) <-chan string {
	return youtube.FetchComments(ctx, y.videoID, y.apiKey)
}

type Logger struct {
	file *os.File
}

func NewLogger(filePath string) (*Logger, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	return &Logger{file: file}, nil
}

func (l *Logger) LogComment(comment string) error {
	_, err := l.file.WriteString(comment + "\n")
	return err
}

func (l *Logger) LogEvent(event string) error {
	_, err := l.file.WriteString("[EVENT] " + event + "\n")
	return err
}

func (l *Logger) Close() error {
	return l.file.Close()
}

func readComments(ctx context.Context, cfg *config.Config) <-chan string {
	var src commentSource
	if cfg.StreamURL != "" && cfg.YouTubeAPIKey != "" {
		videoID, err := youtube.ExtractVideoID(cfg.StreamURL)
		if err != nil {
			log.Printf("Error extracting video ID: %v", err)
			log.Println("Falling back to mock comments")
		} else {
			ytCtx, ytCancel := context.WithCancel(ctx)
			defer ytCancel()

			src = &youtubeSource{videoID: videoID, apiKey: cfg.YouTubeAPIKey}
			ytComments := src.Read(ytCtx)

			select {
			case msg, ok := <-ytComments:
				if ok {
					log.Printf("YouTube API connected with message: %s", msg)
					// Forward first message + remaining on a merged channel
					ch := make(chan string, 1)
					go func() {
						defer close(ch)
						defer ytCancel()
						ch <- msg
						for m := range ytComments {
							ch <- m
						}
					}()
					return ch
				}
			case <-time.After(3 * time.Second):
				log.Println("YouTube API timeout, falling back to mock comments")
			}
		}
	}

	return (&mockSource{}).Read(ctx)
}

// listenForESC watches for ESC key press and cancels context to stop processing.
// Uses raw terminal mode for single-keypress detection without Enter.
func listenForESC(ctx context.Context, cancel context.CancelFunc, escPressed chan<- struct{}) {
	fd := int(os.Stdin.Fd())

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		log.Printf("Warning: ESC listener not available: %v", err)
		return
	}
	defer term.Restore(fd, oldState)

	// Re-enable output processing — term.MakeRaw disables OPOST,
	// which breaks \n → CR+LF translation and causes shifted output.
	if ts, err := unix.IoctlGetTermios(fd, unix.TCGETS); err == nil {
		ts.Oflag |= unix.OPOST
		unix.IoctlSetTermios(fd, unix.TCSETS, ts)
	}

	buf := make([]byte, 1)
	readCh := make(chan struct {
		n   int
		err error
	}, 1)

	go func() {
		n, err := os.Stdin.Read(buf)
		readCh <- struct {
			n   int
			err error
		}{n, err}
	}()

	select {
	case <-ctx.Done():
		return
	case result := <-readCh:
		if result.err != nil || result.n == 0 {
			return
		}
		if buf[0] == 27 { // ESC key
			close(escPressed)
			cancel()
		}
		return
	}
}

func main() {
	devMode := flag.Bool("dev", false, "Enable development mode (print comments to console)")
	flag.Parse()

	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	redisClient, err := redis.NewRedisClient(cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("Error initializing Redis client: %v", err)
	}
	defer redisClient.Close()

	var logger processor.Logger
	if *devMode {
		logger = nil
	} else {
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		logFileName := fmt.Sprintf("comments_%s.log", timestamp)
		realLogger, err := NewLogger(logFileName)
		if err != nil {
			log.Fatalf("Error initializing logger: %v", err)
		}
		logger = realLogger
		defer realLogger.Close()
	}

	// Errors still go into the log file via the Logger.
	if !*devMode {
		log.SetOutput(io.Discard)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ui.DisplayHomeScreen(cfg.TotalLimit, cfg.TimeLimit, cfg.RedisCount, cfg.FinalComment, cfg.APIConnection)

	fmt.Scanln()
	ui.ClearConsole()

	stats := &Stats{}

	proc := processor.New(cfg, stats, redisClient, logger, *devMode)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	comments := readComments(ctx, cfg)

	escPressed := make(chan struct{})
	go listenForESC(ctx, cancel, escPressed)

	go func() {
		defer cancel()
		for comment := range comments {
			terminate, triggeredLetter := proc.Process(comment)
			if terminate {
				return
			}
			// Increment TotalLimit when a threshold was triggered (processor no longer mutates config)
			if triggeredLetter != "" {
				cfg.TotalLimit++
			}
			if stats.CommandsSent >= cfg.TotalLimit {
				return
			}
		}
	}()

	select {
	case <-escPressed:
		fmt.Println("\nESC pressed. Stopping...")
	case <-ctx.Done():
		if stats.CommandsSent >= cfg.TotalLimit || stats.CommentsRead > 0 {
			fmt.Println("\nProcessing completed.")
		}
	case <-signals:
		cancel()
		fmt.Println("\nReceived interrupt signal. Shutting down...")
	case <-time.After(time.Duration(cfg.TimeLimit) * time.Second):
		cancel()
		fmt.Println("\nTime limit reached. Shutting down...")
	}

	ui.DisplayFinalScreen(stats.CommentsRead, stats.LettersTyped, stats.CommandsSent, stats.TriggeredLetters)
}
