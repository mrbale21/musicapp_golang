package services

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"back_music/internal/config"
	"back_music/internal/models"
	"back_music/internal/repository"
)

type ContentBasedService interface {
    GetContentBasedRecommendations(songID string, limit int) ([]models.RecommendationScore, error)
    CalculateSimilarity(song1, song2 *models.Song) float64
    BuildFeatureVector(song *models.Song) []float64
}

type contentBasedService struct {
    songRepo repository.SongRepository
    config   *config.Config
}

func NewContentBasedService(songRepo repository.SongRepository) ContentBasedService {
    return &contentBasedService{
        songRepo: songRepo,
        config:   config.GlobalConfig,
    }
}

func (s *contentBasedService) BuildFeatureVector(song *models.Song) []float64 {
    // Combine audio features into a vector for similarity calculation
    // Weights can be adjusted based on importance
    features := []float64{
        song.Danceability,
        song.Energy,
        float64(song.Key) / 11.0, // Normalize key (0-11) to 0-1
        (song.Loudness + 60) / 60.0, // Normalize loudness (-60 to 0) to 0-1
        float64(song.Mode),
        song.Speechiness,
        song.Acousticness,
        song.Instrumentalness,
        song.Liveness,
        song.Valence,
        song.Tempo / 250.0, // Normalize tempo (0-250) to 0-1
        float64(song.TimeSignature) / 7.0, // Normalize time signature (3-7) to 0-1
    }
    
    // Add popularity (normalized 0-100 ke 0-0.5) supaya
    // pengaruh popularitas tidak terlalu mendominasi dibanding fitur audio.
    features = append(features, (float64(song.Popularity)/100.0)*0.5)
    
    song.FeatureVector = features
    return features
}

func (s *contentBasedService) CalculateSimilarity(song1, song2 *models.Song) float64 {
    // Build feature vectors if not already built
    if song1.FeatureVector == nil {
        s.BuildFeatureVector(song1)
    }
    if song2.FeatureVector == nil {
        s.BuildFeatureVector(song2)
    }
    
    // Calculate cosine similarity
    var dotProduct, norm1, norm2 float64
    
    for i := range song1.FeatureVector {
        dotProduct += song1.FeatureVector[i] * song2.FeatureVector[i]
        norm1 += song1.FeatureVector[i] * song1.FeatureVector[i]
        norm2 += song2.FeatureVector[i] * song2.FeatureVector[i]
    }
    
    norm1 = math.Sqrt(norm1)
    norm2 = math.Sqrt(norm2)
    
    if norm1 == 0 || norm2 == 0 {
        return 0
    }
    
    similarity := dotProduct / (norm1 * norm2)
    
    // Boost similarity for same artist or genre
    if strings.EqualFold(song1.Artist, song2.Artist) {
        similarity += 0.3
    }
    
    if strings.EqualFold(song1.Genre, song2.Genre) && song1.Genre != "" {
        similarity += 0.2
    }
    
    // Ensure similarity doesn't exceed 1.0
    if similarity > 1.0 {
        similarity = 1.0
    }
    
    return similarity
}

// internal/services/content_based_service.go

func (s *contentBasedService) GetContentBasedRecommendations(songID string, limit int) ([]models.RecommendationScore, error) {
    // Get the target song
    targetSong, err := s.songRepo.GetSongByID(songID)
    if err != nil {
        return nil, err
    }
    
    // Get all songs
    allSongs, err := s.songRepo.GetAllSongs()
    if err != nil {
        return nil, err
    }
    
    // Calculate similarity scores
    scores := make([]models.RecommendationScore, 0, len(allSongs))
    
    for _, song := range allSongs {
        if song.ID == targetSong.ID {
            continue // Skip the target song itself
        }
        
        similarity := s.CalculateSimilarity(targetSong, &song)
        
        if similarity >= s.config.SimilarityThreshold {
            explanation := s.generateExplanation(targetSong, &song, similarity)
            
            scores = append(scores, models.RecommendationScore{
                Song:        song,
                Score:       similarity,
                ScoreType:   "content",
                Explanation: explanation,
            })
        }
    }
    
    // Sort by score (descending)
    sort.Slice(scores, func(i, j int) bool {
        return scores[i].Score > scores[j].Score
    })
    
    // Add rank based on sorted position
    for i := range scores {
        scores[i].Rank = i + 1
        
        // Format score to 2 decimal places for better readability
        scores[i].Score = math.Round(scores[i].Score*100) / 100
    }
    
    // Return top N recommendations
    if len(scores) > limit {
        scores = scores[:limit]
    }
    
    return scores, nil
}

// Helper function to generate explanation for content-based recommendations
func (s *contentBasedService) generateExplanation(targetSong, recommendedSong *models.Song, similarity float64) string {
    explanations := []string{}
    
    // Add base similarity explanation
    similarityPercent := int(math.Round(similarity * 100))
    explanations = append(explanations, fmt.Sprintf("Similarity score: %d%%", similarityPercent))
    
    // Check for same artist
    if strings.EqualFold(targetSong.Artist, recommendedSong.Artist) {
        explanations = append(explanations, "Same artist: " + targetSong.Artist)
    }
    
    // Check for same genre
    if strings.EqualFold(targetSong.Genre, recommendedSong.Genre) && targetSong.Genre != "" {
        explanations = append(explanations, "Same genre: " + targetSong.Genre)
    }
    
    // Check for similar audio features
    if math.Abs(targetSong.Danceability - recommendedSong.Danceability) < 0.1 {
        explanations = append(explanations, "Similar danceability")
    }
    
    if math.Abs(targetSong.Energy - recommendedSong.Energy) < 0.1 {
        explanations = append(explanations, "Similar energy level")
    }
    
    if math.Abs(targetSong.Valence - recommendedSong.Valence) < 0.15 {
        mood := "neutral"
        if targetSong.Valence > 0.7 {
            mood = "upbeat/positive"
        } else if targetSong.Valence < 0.3 {
            mood = "mellow/sad"
        }
        explanations = append(explanations, "Similar mood: " + mood)
    }
    
    // If no specific explanations, add generic ones
    if len(explanations) <= 1 { // Only has similarity score
        if similarity >= 0.8 {
            explanations = append(explanations, "Very high audio feature match")
        } else if similarity >= 0.6 {
            explanations = append(explanations, "Good audio feature match")
        } else {
            explanations = append(explanations, "Moderate audio feature match")
        }
    }
    
    return strings.Join(explanations, " • ")
}