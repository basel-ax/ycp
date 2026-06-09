package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/basel-ax/ycp/config"
	"github.com/basel-ax/ycp/redis"
	"github.com/basel-ax/ycp/ui"
)

type Stats struct {
	CommentsRead int
	LettersTyped int
	CommandsSent int
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

func extractVideoID(streamURL string) (string, error) {
	if streamURL == "" {
		return "", fmt.Errorf("empty stream URL")
	}

	parsedURL, err := url.Parse(streamURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %v", err)
	}

	if parsedURL.Host == "youtu.be" {
		path := strings.TrimPrefix(parsedURL.Path, "/")
		return path, nil
	}

	query := parsedURL.Query()
	if v := query.Get("v"); v != "" {
		return v, nil
	}

	if after, ok := strings.CutPrefix(parsedURL.Path, "/live/"); ok {
		return after, nil
	}

	if after, ok := strings.CutPrefix(parsedURL.Path, "/shorts/"); ok {
		return after, nil
	}

	return "", fmt.Errorf("unable to extract video ID from URL: %s", streamURL)
}

func fetchYouTubeComments(ctx context.Context, videoID, apiKey string) <-chan string {
	comments := make(chan string)
	go func() {
		defer close(comments)

		if videoID == "" || apiKey == "" {
			log.Println("YouTube API: videoID or apiKey is empty, skipping")
			return
		}

		log.Printf("YouTube API: Fetching live chat for video %s", videoID)

		client := &http.Client{Timeout: 30 * time.Second}

		videoResp, err := getVideoDetails(videoID, apiKey, client)
		if err != nil {
			log.Printf("YouTube API: Failed to get video details: %v", err)
			return
		}

		liveChatID := videoResp.LiveChatID
		if liveChatID == "" {
			log.Println("YouTube API: No active live chat found for this video")
			return
		}

		log.Printf("YouTube API: Found liveChatId: %s", liveChatID)

		nextPageToken := ""
		for {
			select {
			case <-ctx.Done():
				log.Println("YouTube API: context cancelled, stopping fetch")
				return
			default:
			}

			msgResp, err := getLiveChatMessages(liveChatID, apiKey, nextPageToken, client)
			if err != nil {
				log.Printf("YouTube API: Error fetching messages: %v", err)
				time.Sleep(10 * time.Second)
				continue
			}

			for _, item := range msgResp.Items {
				if item.Snippet.DisplayMessage != "" {
					select {
					case comments <- item.Snippet.DisplayMessage:
					case <-ctx.Done():
						log.Println("YouTube API: context cancelled while sending")
						return
					}
				}
			}

			nextPageToken = msgResp.NextPageToken

			pollInterval := 2000 * time.Millisecond
			if msgResp.PollingIntervalMillis > 0 {
				pollInterval = time.Duration(msgResp.PollingIntervalMillis) * time.Millisecond
			}
			pollInterval += 500 * time.Millisecond

			time.Sleep(pollInterval)
		}
	}()

	return comments
}

type videoDetailsResponse struct {
	Items []struct {
		LiveStreamingDetails struct {
			ActiveLiveChatID string `json:"activeLiveChatId"`
		} `json:"liveStreamingDetails"`
	} `json:"items"`
	LiveChatID string
}

func getVideoDetails(videoID, apiKey string, client *http.Client) (videoDetailsResponse, error) {
	apiURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=liveStreamingDetails&id=%s&key=%s", videoID, apiKey)
	resp, err := client.Get(apiURL)
	if err != nil {
		return videoDetailsResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return videoDetailsResponse{}, err
	}

	var result videoDetailsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return videoDetailsResponse{}, err
	}

	if len(result.Items) > 0 && result.Items[0].LiveStreamingDetails.ActiveLiveChatID != "" {
		result.LiveChatID = result.Items[0].LiveStreamingDetails.ActiveLiveChatID
	}

	return result, nil
}

type liveChatMessagesResponse struct {
	Items                 []liveChatMessage `json:"items"`
	NextPageToken         string            `json:"nextPageToken"`
	PollingIntervalMillis int64             `json:"pollingIntervalMillis"`
}

type liveChatMessage struct {
	Snippet struct {
		DisplayMessage string `json:"displayMessage"`
	} `json:"snippet"`
}

func getLiveChatMessages(liveChatID, apiKey, pageToken string, client *http.Client) (liveChatMessagesResponse, error) {
	apiURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/liveChat/messages?part=snippet&liveChatId=%s&maxResults=10&key=%s", liveChatID, apiKey)
	if pageToken != "" {
		apiURL += "&pageToken=" + pageToken
	}

	resp, err := client.Get(apiURL)
	if err != nil {
		return liveChatMessagesResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return liveChatMessagesResponse{}, err
	}

	var result liveChatMessagesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return liveChatMessagesResponse{}, err
	}

	return result, nil
}

func readComments(ctx context.Context, cfg *config.Config) <-chan string {
	if cfg.StreamURL != "" && cfg.YouTubeAPIKey != "" {
		videoID, err := extractVideoID(cfg.StreamURL)
		if err != nil {
			log.Printf("Error extracting video ID: %v", err)
			log.Println("Falling back to mock comments")
		} else {
			ytCtx, ytCancel := context.WithCancel(ctx)

			ytComments := fetchYouTubeComments(ytCtx, videoID, cfg.YouTubeAPIKey)
			select {
			case msg, ok := <-ytComments:
				if ok {
					go func() {
						for range ytComments {
						}
						ytCancel()
					}()
					log.Printf("YouTube API connected with message: %s", msg)
					return ytComments
				}
			case <-time.After(3 * time.Second):
				log.Println("YouTube API timeout, falling back to mock comments")
			}

			ytCancel()
		}
	}

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

func logOrPrintComment(comment string, logger *Logger, devMode bool) {
	if logger != nil {
		if err := logger.LogComment(comment); err != nil {
			log.Printf("Error logging comment: %v", err)
		}
	} else if devMode {
		fmt.Printf("Comment: %s\n", comment)
	}
}

func checkFinalComment(comment string, cfg *config.Config, logger *Logger) bool {
	if cfg.FinalComment != "" && strings.Contains(comment, cfg.FinalComment) {
		if logger != nil {
			if err := logger.LogEvent("FINAL_COMMENT detected"); err != nil {
				log.Printf("Error logging event: %v", err)
			}
		}
		return true
	}
	return false
}

func processDoubleLetters(comment string, cfg *config.Config, stats *Stats, redisClient *redis.RedisClient) {
	seen := make(map[rune]bool)
	for _, char := range comment {
		if seen[char] {
			continue
		}
		seen[char] = true

		charStr := string(char)
		if strings.Count(comment, charStr) >= 2 && strings.Contains(cfg.FinalComment, charStr) {
			count, err := redisClient.IncrementButtonCount(charStr)
			if err != nil {
				log.Printf("Error incrementing count for %s: %v", charStr, err)
				continue
			}

			if count > cfg.RedisCount {
				if err := redisClient.ResetButtonCount(charStr); err != nil {
					log.Printf("Error resetting count for %s: %v", charStr, err)
				}
				cfg.TotalLimit++
				fmt.Printf("Letter: %s\n", charStr)
			}

			stats.LettersTyped++
			stats.CommandsSent++
		}
	}
}

func processComment(comment string, cfg *config.Config, stats *Stats, logger *Logger, redisClient *redis.RedisClient, devMode bool) bool {
	logOrPrintComment(comment, logger, devMode)

	if checkFinalComment(comment, cfg, logger) {
		return true
	}

	processDoubleLetters(comment, cfg, stats, redisClient)

	stats.CommentsRead++
	return false
}

func main() {
	devMode := flag.Bool("dev", false, "Enable development mode (print comments to console)")
	flag.Parse()

	cfg, err := config.LoadConfig("example.env")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	redisClient, err := redis.NewRedisClient(cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("Error initializing Redis client: %v", err)
	}
	defer redisClient.Close()

	var logger *Logger
	if *devMode {
		logger = nil
	} else {
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		logFileName := fmt.Sprintf("comments_%s.log", timestamp)
		logger, err = NewLogger(logFileName)
		if err != nil {
			log.Fatalf("Error initializing logger: %v", err)
		}
		defer logger.Close()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ui.DisplayHomeScreen(cfg.TotalLimit, cfg.TimeLimit, cfg.RedisCount, cfg.FinalComment, cfg.APIConnection)

	fmt.Scanln()
	ui.ClearConsole()

	stats := &Stats{}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	comments := readComments(ctx, cfg)

	go func() {
		defer cancel()
		for comment := range comments {
			if processComment(comment, cfg, stats, logger, redisClient, *devMode) {
				return
			}
			if stats.CommandsSent >= cfg.TotalLimit {
				return
			}
		}
	}()

	select {
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

	ui.DisplayFinalScreen(stats.CommentsRead, stats.LettersTyped, stats.CommandsSent)
}
