package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

type YouTubeService interface {
	SearchAudio(query string) (string, error)
}

type YoutubeService struct {
	apiKey string
}

func NewYouTubeService() YouTubeService {
	return &YoutubeService{
		apiKey: os.Getenv("YOUTUBE_API_KEY"),
	}
}

func (s *YoutubeService) SearchAudio(query string) (string, error) {
	baseURL := "https://www.googleapis.com/youtube/v3/search"
	searchQuery := query + " official audio"

	// Setup Query Parameters
	params := url.Values{}
	params.Add("part", "id")
	params.Add("q", searchQuery)
	params.Add("type", "video")
	params.Add("maxResults", "1")
	params.Add("key", s.apiKey)
	params.Add("videoCategoryId", "10") // Music Category

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// HTTP Request
	resp, err := http.Get(fullURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("youtube api error: status %d", resp.StatusCode)
	}

	// Parsing Response
	var result struct {
		Items []struct {
			Id struct {
				VideoId string `json:"videoId"`
			} `json:"id"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Items) > 0 {
		return result.Items[0].Id.VideoId, nil
	}

	return "", fmt.Errorf("no audio found on youtube")
}