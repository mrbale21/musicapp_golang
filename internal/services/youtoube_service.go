package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
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
	
	// Multiple search strategies for better results
	searchQueries := []string{
		query + " official audio",
		query + " official music video",
		query + " lyrics",
		query + " audio only",
		query, // fallback to original query
	}

	var lastErr error
	
	for _, searchQuery := range searchQueries {
		// Setup Query Parameters
		params := url.Values{}
		params.Add("part", "id,snippet")
		params.Add("q", searchQuery)
		params.Add("type", "video")
		params.Add("maxResults", "5") // Get more results to choose from
		params.Add("key", s.apiKey)
		params.Add("videoCategoryId", "10") // Music Category
		params.Add("order", "relevance")
		params.Add("safeSearch", "moderate")

		fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

		// HTTP Request
		resp, err := http.Get(fullURL)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("youtube api error: status %d", resp.StatusCode)
			continue
		}

		// Parsing Response with snippet for better filtering
		var result struct {
			Items []struct {
				Id struct {
					VideoId string `json:"videoId"`
				} `json:"id"`
				Snippet struct {
					Title        string `json:"title"`
					Description  string `json:"description"`
					ChannelTitle string `json:"channelTitle"`
				} `json:"snippet"`
			} `json:"items"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			lastErr = err
			continue
		}

		// Filter and prioritize results
		var fallbackID string
		for _, item := range result.Items {
			title := item.Snippet.Title
			description := item.Snippet.Description
			channel := item.Snippet.ChannelTitle

			// Skip if title contains unwanted keywords
			if containsUnwantedKeywords(title) {
				continue
			}

			// Store first valid result as fallback
			if fallbackID == "" {
				fallbackID = item.Id.VideoId
			}

			// Prioritize official channels and good titles
			if isOfficialChannel(channel) || isGoodTitle(title, description) {
				return item.Id.VideoId, nil
			}
		}

		// If no perfect match found, return first fallback (if exists)
		if fallbackID != "" {
			return fallbackID, nil
		}
	}

	if lastErr != nil {
		return "", lastErr
	}
	return "", fmt.Errorf("no audio found on youtube after trying multiple search strategies")
}

// Helper functions for better filtering
func containsUnwantedKeywords(title string) bool {
	// Relaxed filter - only exclude most problematic content
	unwanted := []string{
		"karaoke", "reaction", "tutorial", "review",
		"parody", "mashup",
	}
	
	titleLower := strings.ToLower(title)
	for _, word := range unwanted {
		if strings.Contains(titleLower, word) {
			return true
		}
	}
	return false
}

func isOfficialChannel(channel string) bool {
	official := []string{
		"official", "vevo", "records", "music", "entertainment",
		"tv", "radio", "fm",
	}
	
	channelLower := strings.ToLower(channel)
	for _, word := range official {
		if strings.Contains(channelLower, word) {
			return true
		}
	}
	return false
}

func isGoodTitle(title, description string) bool {
	content := strings.ToLower(title + " " + description)
	
	good := []string{
		"official", "audio", "music video", "lyric", "lyrics",
		"original", "album", "single", "track",
	}
	
	for _, word := range good {
		if strings.Contains(content, word) {
			return true
		}
	}
	return false
}