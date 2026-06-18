package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func ExtractVideoID(streamURL string) (string, error) {
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

	if parsedURL.Host == "" && parsedURL.Scheme == "" {
		return parsedURL.Path, nil
	}

	return "", fmt.Errorf("unable to extract video ID from URL: %s", streamURL)
}

func FetchComments(ctx context.Context, videoID, apiKey string) <-chan string {
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
