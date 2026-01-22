package handlers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"back_music/internal/models"
	"back_music/internal/repository"
	"back_music/internal/services"
)

type RecommendationHandler struct {
    contentService      services.ContentBasedService
    collaborativeService services.CollaborativeService
    hybridService       services.HybridService
    smartHybridService  services.SmartHybridService
    db                  *gorm.DB
     songRepo            repository.SongRepository
}

func NewRecommendationHandler(
    content services.ContentBasedService, 
    collaborative services.CollaborativeService, 
    hybrid services.HybridService, 
    smartHybrid services.SmartHybridService,
    db *gorm.DB, 
     songRepo repository.SongRepository,
) *RecommendationHandler {
    return &RecommendationHandler{
        contentService:      content,
        collaborativeService: collaborative,
        hybridService:       hybrid,
        smartHybridService:  smartHybrid,
        db:                  db, 
        songRepo:            songRepo,
    }
}

func (h *RecommendationHandler) checkIfSongLiked(songID string, userID uint) (bool, error) {
    var count int64
    err := h.db.Model(&models.UserLike{}).
        Where("song_id = ? AND user_id = ?", songID, userID).
        Count(&count).Error
    
    return count > 0, err
}

// ⭐⭐ NEW: Method untuk set like status pada array recommendations
func (h *RecommendationHandler) setLikeStatusForRecommendations(recommendations []models.RecommendationScore, userID uint) {
    if userID == 0 || len(recommendations) == 0 {
        return
    }
    
    // Collect all song IDs
    songIDs := make([]string, len(recommendations))
    for i, rec := range recommendations {
        songIDs[i] = rec.Song.ID
    }
    
    // Query which songs are liked by user
    var likedSongIDs []string
    h.db.Model(&models.UserLike{}).
        Where("user_id = ? AND song_id IN ?", userID, songIDs).
        Pluck("song_id", &likedSongIDs)
    
    // Create map for O(1) lookup
    likedMap := make(map[string]bool)
    for _, id := range likedSongIDs {
        likedMap[id] = true
    }
    
    // Set IsLiked field
    for i := range recommendations {
        recommendations[i].Song.IsLiked = likedMap[recommendations[i].Song.ID]
    }
}

func (h *RecommendationHandler) GetContentBasedRecommendations(c *gin.Context) {
    songID := c.Param("song_id")
    limitStr := c.DefaultQuery("limit", "10")
    userID := c.GetUint("user_id")
    
    limit, err := strconv.Atoi(limitStr)
    if err != nil || limit <= 0 {
        limit = 10
    }
    
    if limit > 20 {
        limit = 20 // Safety limit
    }
    
    // ⭐⭐ PERBAIKAN: Panggil service dulu untuk dapat recommendations
    recommendations, err := h.contentService.GetContentBasedRecommendations(songID, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Failed to generate recommendations",
            "error":   err.Error(),
        })
        return
    }
    
    // ⭐⭐ PERBAIKAN: Set like status untuk recommendations
    if userID > 0 {
        h.setLikeStatusForRecommendations(recommendations, userID)
    }
    
    // Format scores and add rank
    for i := range recommendations {
        recommendations[i].Rank = i + 1
        recommendations[i].Score = math.Round(recommendations[i].Score*100) / 100
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Content-based recommendations fetched",
        "data": gin.H{
            "song_id":         songID,
            "recommendations": recommendations,
            "count":           len(recommendations),
            "type":            "content-based",
            "metadata": gin.H{
                "max_recommendations": limit,
            },
        },
    })
}

func (h *RecommendationHandler) GetCollaborativeRecommendations(c *gin.Context) {
    userID := c.GetUint("user_id")
    limitStr := c.DefaultQuery("limit", "10")
    
    limit, err := strconv.Atoi(limitStr)
    if err != nil || limit <= 0 {
        limit = 10
    }
    
    if limit > 20 {
        limit = 20 // Safety limit
    }
    
    recommendations, err := h.collaborativeService.GetCollaborativeRecommendations(userID, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Failed to generate recommendations",
            "error":   err.Error(),
        })
        return
    }
    
    // ⭐⭐ PERBAIKAN: Set like status (meskipun ini collaborative, tetap bisa check)
    if userID > 0 {
        h.setLikeStatusForRecommendations(recommendations, userID)
    }
    
    // Format scores and add rank untuk collaborative
    for i := range recommendations {
        recommendations[i].Rank = i + 1
        recommendations[i].Score = math.Round(recommendations[i].Score*100) / 100
        
        // Tambahkan explanation jika kosong
        if recommendations[i].Explanation == "" {
            recommendations[i].Explanation = h.generateCollaborativeExplanation(&recommendations[i])
        }
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Collaborative recommendations fetched",
        "data": gin.H{
            "user_id":         userID,
            "recommendations": recommendations,
            "count":           len(recommendations),
            "type":            "collaborative",
        },
    })
}

func (h *RecommendationHandler) GetHybridRecommendations(c *gin.Context) {
    userID := c.GetUint("user_id")
    songID := c.Query("song_id")
    limitStr := c.DefaultQuery("limit", "10")
    
    limit, err := strconv.Atoi(limitStr)
    if err != nil || limit <= 0 {
        limit = 10
    }
    
    if limit > 20 {
        limit = 20 // Safety limit
    }
    
    if songID == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "Song ID is required for hybrid recommendations",
        })
        return
    }
    
    recommendations, err := h.hybridService.GetHybridRecommendations(userID, songID, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Failed to generate recommendations",
            "error":   err.Error(),
        })
        return
    }
    
    // ⭐⭐ PERBAIKAN: Set like status untuk recommendations
    if userID > 0 {
        h.setLikeStatusForRecommendations(recommendations, userID)
    }
    
    // Format scores and add rank untuk hybrid
    for i := range recommendations {
        recommendations[i].Rank = i + 1
        recommendations[i].Score = math.Round(recommendations[i].Score*100) / 100
        
        // Tambahkan explanation jika kosong
        if recommendations[i].Explanation == "" {
            recommendations[i].Explanation = h.generateHybridExplanation(&recommendations[i])
        }
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Hybrid recommendations fetched",
        "data": gin.H{
            "user_id":         userID,
            "song_id":         songID,
            "recommendations": recommendations,
            "count":           len(recommendations),
            "type":            "hybrid",
        },
    })
}

func (h *RecommendationHandler) GetSmartHybridRecommendations(c *gin.Context) {
    userID := c.GetUint("user_id")
    limitStr := c.DefaultQuery("limit", "10")
    
    limit, err := strconv.Atoi(limitStr)
    if err != nil || limit <= 0 {
        limit = 10
    }
    
    if limit > 20 {
        limit = 20
    }
    
    recommendations, err := h.smartHybridService.GetSmartHybridRecommendations(userID, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Failed to generate recommendations",
            "error":   err.Error(),
        })
        return
    }
    
    // ⭐⭐ PERBAIKAN: Set like status untuk recommendations
    if userID > 0 {
        h.setLikeStatusForRecommendations(recommendations, userID)
    }
    
    // Format scores and add rank untuk smart hybrid
    for i := range recommendations {
        recommendations[i].Rank = i + 1
        recommendations[i].Score = math.Round(recommendations[i].Score*100) / 100
        
        // Tambahkan explanation jika kosong
        if recommendations[i].Explanation == "" {
            recommendations[i].Explanation = h.generateSmartHybridExplanation(&recommendations[i])
        }
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Smart hybrid recommendations fetched",
        "data": gin.H{
            "user_id":         userID,
            "recommendations": recommendations,
            "count":           len(recommendations),
            "type":            "smart-hybrid",
            "algorithm_info": "Combines content-based, collaborative, and popularity factors",
        },
    })
}

func (h *RecommendationHandler) GetPopularSongs(c *gin.Context) {
    limitStr := c.DefaultQuery("limit", "20")
    userID := c.GetUint("user_id")
    
    limit, err := strconv.Atoi(limitStr)
    if err != nil || limit <= 0 {
        limit = 20
    }
    
    // ⭐⭐ PERBAIKAN: Ambil data dari database via songRepo
    // Kita perlu akses ke songRepo di RecommendationHandler
    songs, err := h.songRepo.GetPopularSongs(limit) // Anda perlu tambah field ini di struct
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Failed to fetch popular songs",
            "error":   err.Error(),
        })
        return
    }
    
    // ⭐⭐ PERBAIKAN: Check like status jika user logged in
    if userID > 0 && len(songs) > 0 {
        // Collect all song IDs
        songIDs := make([]string, len(songs))
        for i, song := range songs {
            songIDs[i] = song.ID
        }
        
        // Query which songs are liked by user
        var likedSongIDs []string
        h.db.Model(&models.UserLike{}).
            Where("user_id = ? AND song_id IN ?", userID, songIDs).
            Pluck("song_id", &likedSongIDs)
        
        // Create map for O(1) lookup
        likedMap := make(map[string]bool)
        for _, id := range likedSongIDs {
            likedMap[id] = true
        }
        
        // Set IsLiked field
        for i := range songs {
            songs[i].IsLiked = likedMap[songs[i].ID]
        }
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Popular songs fetched",
        "data": gin.H{
            "songs": songs,
            "limit": limit,
            "total": len(songs),
        },
    })
}
// Helper functions untuk generate explanation di handler
func (h *RecommendationHandler) generateCollaborativeExplanation(rec *models.RecommendationScore) string {
    explanations := []string{}
    
    scorePercent := int(math.Round(rec.Score * 100))
    explanations = append(explanations, fmt.Sprintf("Match score: %d%%", scorePercent))
    
    // Berdasarkan score type
    if rec.ScoreType == "collaborative" {
        explanations = append(explanations, "Based on users with similar tastes")
    }
    
    // Tambahkan info popularity
    if rec.Song.Popularity > 80 {
        explanations = append(explanations, "Highly popular")
    } else if rec.Song.Popularity > 60 {
        explanations = append(explanations, "Popular")
    }
    
    return strings.Join(explanations, " • ")
}

func (h *RecommendationHandler) generateHybridExplanation(rec *models.RecommendationScore) string {
    explanations := []string{}
    
    scorePercent := int(math.Round(rec.Score * 100))
    explanations = append(explanations, fmt.Sprintf("Hybrid score: %d%%", scorePercent))
    
    // Berdasarkan score type
    if rec.ScoreType == "hybrid" {
        explanations = append(explanations, "Combines content similarity and user preferences")
    }
    
    return strings.Join(explanations, " • ")
}

func (h *RecommendationHandler) generateSmartHybridExplanation(rec *models.RecommendationScore) string {
    explanations := []string{}
    
    scorePercent := int(math.Round(rec.Score * 100))
    explanations = append(explanations, fmt.Sprintf("Smart score: %d%%", scorePercent))
    
    // Berdasarkan score type
    if rec.ScoreType == "smart-hybrid" {
        explanations = append(explanations, "Optimized blend of multiple recommendation strategies")
    }
    
    if rec.Song.Popularity > 75 {
        explanations = append(explanations, "Trending now")
    }
    
    return strings.Join(explanations, " • ")
}