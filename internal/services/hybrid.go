package services

import (
	"sort"

	"back_music/internal/config"
	"back_music/internal/models"
)

type HybridService interface {
    GetHybridRecommendations(userID uint, songID string, limit int) ([]models.RecommendationScore, error)
}

type hybridService struct {
    contentService      ContentBasedService
    collaborativeService CollaborativeService
    config             *config.Config
}

func NewHybridService(content ContentBasedService, collaborative CollaborativeService) HybridService {
    return &hybridService{
        contentService:      content,
        collaborativeService: collaborative,
        config:             config.GlobalConfig,
    }
}

func (s *hybridService) GetHybridRecommendations(userID uint, songID string, limit int) ([]models.RecommendationScore, error) {
    // Selalu ambil content-based sebagai dasar hybrid
    contentRecs, err := s.contentService.GetContentBasedRecommendations(songID, limit*2)
    if err != nil {
        return nil, err
    }
    
    // Ambil collaborative recommendations hanya jika:
    // - userID valid
    // - dan weight collaborative > 0
    // Jika gagal, kita abaikan saja dan tetap kembalikan content-based.
    var collabRecs []models.RecommendationScore
    if userID != 0 && s.config.CollaborativeWeight > 0 {
        if cr, err := s.collaborativeService.GetCollaborativeRecommendations(userID, limit*2); err == nil {
            collabRecs = cr
        }
    }
    
    // Combine recommendations with improved diversity
    combinedScores := make(map[string]models.RecommendationScore)
    
    // Add content-based scores with weight
    for _, rec := range contentRecs {
        combined := combinedScores[rec.Song.ID]
        combined.Song = rec.Song
        combined.Score += rec.Score * s.config.ContentWeight
        combined.ScoreType = "hybrid"
        combinedScores[rec.Song.ID] = combined
    }
    
    // Add collaborative scores with weight
    for _, rec := range collabRecs {
        combined := combinedScores[rec.Song.ID]
        combined.Song = rec.Song
        combined.Score += rec.Score * s.config.CollaborativeWeight
        combined.ScoreType = "hybrid"
        combinedScores[rec.Song.ID] = combined
    }
    
    // Convert map to slice
    finalScores := make([]models.RecommendationScore, 0, len(combinedScores))
    for _, score := range combinedScores {
        finalScores = append(finalScores, score)
    }
    
    // Sort by combined score (descending)
    sort.Slice(finalScores, func(i, j int) bool {
        return finalScores[i].Score > finalScores[j].Score
    })
    
    // Add diversity spread to prevent very similar scores
    // This creates more variation in results
    if len(finalScores) > 1 {
        for i := 0; i < len(finalScores)-1; i++ {
            // If scores are too similar (within 0.03), add small variance
            if finalScores[i].Score-finalScores[i+1].Score < 0.03 && finalScores[i].Score > 0 {
                // Add position-based variance to create more separation
                finalScores[i+1].Score -= float64((i+1)%5) * 0.005
            }
        }
    }
    
    // Return top N recommendations
    if len(finalScores) > limit {
        finalScores = finalScores[:limit]
    }
    
    return finalScores, nil
}