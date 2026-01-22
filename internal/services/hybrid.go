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
    // Get content-based recommendations
    contentRecs, err := s.contentService.GetContentBasedRecommendations(songID, limit*2)
    if err != nil {
        return nil, err
    }
    
    // Get collaborative recommendations
    collabRecs, err := s.collaborativeService.GetCollaborativeRecommendations(userID, limit*2)
    if err != nil {
        return nil, err
    }
    
    // Combine recommendations
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
    
    // Return top N recommendations
    if len(finalScores) > limit {
        finalScores = finalScores[:limit]
    }
    
    return finalScores, nil
}