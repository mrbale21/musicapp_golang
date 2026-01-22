package services

import (
	"log"

	"back_music/internal/config"
	"back_music/internal/database"
	"back_music/internal/models"
)

type SmartHybridService interface {
    GetSmartHybridRecommendations(userID uint, limit int) ([]models.RecommendationScore, error)
}

// ⭐⭐ PERUBAHAN: Interface lebih sederhana, cukup 1 method
type smartHybridService struct {
    contentService      ContentBasedService
    collaborativeService CollaborativeService
    hybridService       HybridService  // ⭐⭐ TAMBAH INI
    config             *config.Config
}

func NewSmartHybridService(content ContentBasedService, collaborative CollaborativeService, hybrid HybridService) SmartHybridService {
    return &smartHybridService{
        contentService:      content,
        collaborativeService: collaborative,
        hybridService:       hybrid,  // ⭐⭐ TAMBAH INI
        config:             config.GlobalConfig,
    }
}

func (s *smartHybridService) GetSmartHybridRecommendations(userID uint, limit int) ([]models.RecommendationScore, error) {
    log.Printf("🔄 Smart hybrid for user %d, limit %d", userID, limit)
    
    // 1. Cek jika user baru (no likes/plays)
    var likeCount, playCount int64
    database.DB.Model(&models.UserLike{}).Where("user_id = ?", userID).Count(&likeCount)
    database.DB.Model(&models.UserPlay{}).Where("user_id = ?", userID).Count(&playCount)
    
    log.Printf("📊 User stats: %d likes, %d plays", likeCount, playCount)
    
    if likeCount == 0 && playCount == 0 {
        // User baru: return popular songs
        log.Println("👤 New user detected, returning popular songs")
        return s.getPopularSongsFallback(limit)
    }
    
    // 2. Cari seed song berdasarkan user behavior
    seedSongID, strategy := s.findBestSeedSong(userID)
    
    if seedSongID == "" {
        log.Println("⚠️ No suitable seed song found, using collaborative")
        return s.collaborativeService.GetCollaborativeRecommendations(userID, limit)
    }
    
    log.Printf("🎯 Using seed song: %s (strategy: %s)", seedSongID, strategy)
    
    // 3. Panggil hybrid service dengan seed yang ditemukan
    return s.hybridService.GetHybridRecommendations(userID, seedSongID, limit)
}

// ⭐⭐ FUNGSI BARU: Cari seed song terbaik
func (s *smartHybridService) findBestSeedSong(userID uint) (string, string) {
    // Priority 1: Last liked song
    var lastLike models.UserLike
    if err := database.DB.Where("user_id = ?", userID).
        Order("created_at DESC").
        First(&lastLike).Error; err == nil {
        return lastLike.SongID, "last_liked"
    }
    
    // Priority 2: Most played song
    var mostPlayed models.UserPlay
    if err := database.DB.Where("user_id = ?", userID).
        Order("play_count DESC, last_played DESC").
        First(&mostPlayed).Error; err == nil && mostPlayed.PlayCount > 1 {
        return mostPlayed.SongID, "most_played"
    }
    
    // Priority 3: Last played song
    var lastPlay models.UserPlay
    if err := database.DB.Where("user_id = ?", userID).
        Order("last_played DESC").
        First(&lastPlay).Error; err == nil {
        return lastPlay.SongID, "last_played"
    }
    
    // Priority 4: Random from likes
    var randomLike models.UserLike
    if err := database.DB.Where("user_id = ?", userID).
        Order("RANDOM()").
        First(&randomLike).Error; err == nil {
        return randomLike.SongID, "random_liked"
    }
    
    return "", "none"
}

// ⭐⭐ FUNGSI HELPER: Popular songs fallback
func (s *smartHybridService) getPopularSongsFallback(limit int) ([]models.RecommendationScore, error) {
    // Get popular songs dari database
    var songs []models.Song
    if err := database.DB.Order("popularity DESC").Limit(limit).Find(&songs).Error; err != nil {
        return nil, err
    }
    
    // Convert ke RecommendationScore
    recommendations := make([]models.RecommendationScore, 0, len(songs))
    for _, song := range songs {
        recommendations = append(recommendations, models.RecommendationScore{
            Song:      song,
            Score:     float64(song.Popularity) / 100.0, // Normalize 0-1
            ScoreType: "popular_fallback",
        })
    }
    
    return recommendations, nil
}

// ⭐⭐ HAPUS method yang tidak perlu dari interface
// Tidak perlu GetHybridBasedOnRecentActivity dan GetHybridBasedOnMostLikedGenre
// Cukup GetSmartHybridRecommendations saja