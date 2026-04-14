package main

import (
	"bufio"
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
)

type Config = config.Config

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

func displayHomeScreen(cfg *Config) {
	fmt.Println("=== YouTube Stream Comments Processor ===")
	fmt.Println("Parameters:")
	fmt.Printf("Total Limit: %d\n", cfg.TotalLimit)
	fmt.Printf("Time Limit: %d seconds\n", cfg.TimeLimit)
	fmt.Printf("Final Comment: %s\n", cfg.FinalComment)
	fmt.Printf("API Connection: %s\n", cfg.APIConnection)
	fmt.Printf("Stream URL: %s\n", cfg.StreamURL)
	fmt.Printf("Redis Count: %d\n", cfg.RedisCount)
	fmt.Println("Press Enter to clear the console and start reading comments...")
}

func displayFinalScreen(stats *Stats) {
	fmt.Println("=== Final Statistics ===")
	fmt.Printf("Comments Read: %d\n", stats.CommentsRead)
	fmt.Printf("Letters Typed: %d\n", stats.LettersTyped)
	fmt.Printf("Commands Sent: %d\n", stats.CommandsSent)
}

func clearConsole() {
	fmt.Print("\033[H\033[2J")
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

	if strings.HasPrefix(parsedURL.Path, "/live/") {
		return strings.TrimPrefix(parsedURL.Path, "/live/"), nil
	}

	if strings.HasPrefix(parsedURL.Path, "/shorts/") {
		return strings.TrimPrefix(parsedURL.Path, "/shorts/"), nil
	}

	return "", fmt.Errorf("unable to extract video ID from URL: %s", streamURL)
}

func fetchYouTubeComments(videoID, apiKey string) <-chan string {
	comments := make(chan string)
	go func() {
		defer close(comments)

		if videoID == "" || apiKey == "" {
			log.Println("YouTube API: videoID or apiKey is empty, skipping")
			return
		}

		log.Printf("YouTube API: Fetching live chat for video %s", videoID)

		client := &http.Client{}

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
			msgResp, err := getLiveChatMessages(liveChatID, apiKey, nextPageToken, client)
			if err != nil {
				log.Printf("YouTube API: Error fetching messages: %v", err)
				time.Sleep(10 * time.Second)
				continue
			}

			for _, item := range msgResp.Items {
				if item.Snippet.DisplayMessage != "" {
					comments <- item.Snippet.DisplayMessage
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

func readComments(cfg *Config) <-chan string {
	if cfg.StreamURL != "" && cfg.YouTubeAPIKey != "" {
		videoID, err := extractVideoID(cfg.StreamURL)
		if err != nil {
			log.Printf("Error extracting video ID: %v", err)
			log.Println("Falling back to mock comments")
		} else {
			ytComments := fetchYouTubeComments(videoID, cfg.YouTubeAPIKey)
			select {
			case msg, ok := <-ytComments:
				if ok {
					go func() {
						for range ytComments {
						}
					}()
					log.Printf("YouTube API connected with message: %s", msg)
					return ytComments
				}
			case <-time.After(3 * time.Second):
				log.Println("YouTube API timeout, falling back to mock comments")
			}
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
			comments <- c
			time.Sleep(1 * time.Second)
		}
	}()
	return comments
}

func processComment(comment string, cfg *Config, stats *Stats, logger *Logger, redisClient *redis.RedisClient, devMode bool) bool {
	if logger != nil {
		if err := logger.LogComment(comment); err != nil {
			log.Printf("Error logging comment: %v", err)
		}
	} else if devMode {
		fmt.Printf("Comment: %s\n", comment)
	}

	if cfg.FinalComment != "" && strings.Contains(comment, cfg.FinalComment) {
		if logger != nil {
			if err := logger.LogEvent("FINAL_COMMENT detected"); err != nil {
				log.Printf("Error logging event: %v", err)
			}
		}
		return true
	}

	seen := make(map[rune]bool)
	for _, char := range comment {
		if seen[char] {
			continue
		}
		seen[char] = true

		charStr := string(char)
		if strings.Count(comment, charStr) >= 2 {
			if strings.Contains(cfg.FinalComment, charStr) {
				if err := redisClient.IncrementButtonCount(charStr); err != nil {
					log.Printf("Error incrementing count for %s: %v", charStr, err)
					continue
				}

				count, err := redisClient.GetButtonCount(charStr)
				if err != nil {
					log.Printf("Error getting count for %s: %v", charStr, err)
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

	displayHomeScreen(cfg)

	bufio.NewReader(os.Stdin).ReadBytes('\n')
	clearConsole()

	stats := &Stats{}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	comments := readComments(cfg)
	done := make(chan bool)

	go func() {
		for comment := range comments {
			select {
			case <-done:
				return
			default:
				if processComment(comment, cfg, stats, logger, redisClient, *devMode) {
					close(done)
					return
				}
				if stats.CommandsSent >= cfg.TotalLimit {
					close(done)
					return
				}
			}
		}
	}()

	select {
	case <-done:
		fmt.Println("\nProcessing completed.")
	case <-signals:
		fmt.Println("\nReceived interrupt signal. Shutting down...")
	case <-time.After(time.Duration(cfg.TimeLimit) * time.Second):
		fmt.Println("\nTime limit reached. Shutting down...")
	}

	displayFinalScreen(stats)
}
