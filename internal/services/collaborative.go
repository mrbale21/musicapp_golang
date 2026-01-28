package services

import (
	"math"
	"sort"

	"back_music/internal/config"
	"back_music/internal/models"
	"back_music/internal/repository"
)

type CollaborativeService interface {
    GetCollaborativeRecommendations(userID uint, limit int) ([]models.RecommendationScore, error)
    CalculateUserSimilarity(userID1, userID2 uint) (float64, error)
    FindSimilarUsers(userID uint, threshold float64) ([]uint, error)
}

type collaborativeService struct {
    songRepo repository.SongRepository
    userRepo repository.UserRepository
    config   *config.Config
}

func NewCollaborativeService(userRepo repository.UserRepository, songRepo repository.SongRepository) CollaborativeService {
    return &collaborativeService{
        userRepo: userRepo,
        songRepo: songRepo,
        config:   config.GlobalConfig,
    }
}

func (s *collaborativeService) CalculateUserSimilarity(userID1, userID2 uint) (float64, error) {
    user1, err := s.userRepo.FindUserByID(userID1)
    if err != nil {
        return 0, err
    }
    
    user2, err := s.userRepo.FindUserByID(userID2)
    if err != nil {
        return 0, err
    }
    
    // Create sets of liked and played songs for each user
    user1Likes := make(map[string]bool)
    user1Plays := make(map[string]int)
    
    for _, like := range user1.Likes {
        user1Likes[like.SongID] = true
    }
    
    for _, play := range user1.Plays {
        user1Plays[play.SongID] = play.PlayCount
    }
    
    // Calculate Jaccard similarity for likes
    var intersection, union float64
    
    user2LikedSongs := make(map[string]bool)
    for _, like := range user2.Likes {
        user2LikedSongs[like.SongID] = true
        union++
        
        if user1Likes[like.SongID] {
            intersection++
        }
    }
    
    for songID := range user1Likes {
        if !user2LikedSongs[songID] {
            union++
        }
    }
    
    likeSimilarity := 0.0
    if union > 0 {
        likeSimilarity = intersection / union
    }
    
    // Calculate similarity based on play counts (cosine similarity)
    var dotProduct, norm1, norm2 float64
    
    allSongIDs := make(map[string]bool)
    for songID := range user1Plays {
        allSongIDs[songID] = true
    }
    
    user2Plays := make(map[string]int)
    for _, play := range user2.Plays {
        user2Plays[play.SongID] = play.PlayCount
        allSongIDs[play.SongID] = true
    }
    
    for songID := range allSongIDs {
        play1 := float64(user1Plays[songID])
        play2 := float64(user2Plays[songID])
        
        dotProduct += play1 * play2
        norm1 += play1 * play1
        norm2 += play2 * play2
    }
    
    playSimilarity := 0.0
    if norm1 > 0 && norm2 > 0 {
        playSimilarity = dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
    }
    
    // Combine similarities (weighted average)
    totalSimilarity := (likeSimilarity*0.6 + playSimilarity*0.4)
    
    return totalSimilarity, nil
}

func (s *collaborativeService) FindSimilarUsers(userID uint, threshold float64) ([]uint, error) {
    // In production, you'd have a more efficient way to find similar users
    // This is a simplified version
    
    // For now, return empty - in real implementation, you'd query the database
    // for users with similar activity patterns
    return []uint{}, nil
}

func (s *collaborativeService) GetCollaborativeRecommendations(userID uint, limit int) ([]models.RecommendationScore, error) {
    user, err := s.userRepo.FindUserByID(userID)
    if err != nil {
        return nil, err
    }
    
    // Get all songs the user has liked or played
    userSongIDs := make(map[string]bool)
    for _, like := range user.Likes {
        userSongIDs[like.SongID] = true
    }
    for _, play := range user.Plays {
        userSongIDs[play.SongID] = true
    }
    
    // Precompute preferred genres dari lagu-lagu yang user like (sekali query saja)
    userGenres := make(map[string]int)
    totalLikes := len(user.Likes)
    if totalLikes > 0 {
        ids := make([]string, 0, totalLikes)
        for _, like := range user.Likes {
            ids = append(ids, like.SongID)
        }
        likedSongs, err := s.songRepo.GetSongsByIDs(ids)
        if err == nil {
            for _, ls := range likedSongs {
                if ls.Genre != "" {
                    userGenres[ls.Genre]++
                }
            }
        }
    }
    
    // Ambil subset lagu populer saja (lebih cepat daripada full table scan),
    // lalu filter yang belum pernah user dengar.
    allSongs, err := s.songRepo.GetPopularSongs(limit * 3)
    if err != nil {
        return nil, err
    }
    
    // For each song not interacted with, calculate a score
    scores := make([]models.RecommendationScore, 0, len(allSongs))
    
    for _, song := range allSongs {
        if userSongIDs[song.ID] {
            continue // Skip songs user already knows
        }
        
        // Simplified collaborative filtering:
        // Score based on popularity dan genre alignment dengan preferensi user.
        // Genre dibuat jauh lebih dominan supaya hasilnya lebih beda
        // dibanding content-based (yang fokus ke fitur audio).
        
        score := 0.0
        
        // Popularity component (0–1)
        popularityScore := float64(song.Popularity) / 100.0
        
        // GenreScore: seberapa sering genre ini muncul di lagu yang user like (0–1)
        genreScore := 0.0
        if totalLikes > 0 && song.Genre != "" {
            genreCount := userGenres[song.Genre]
            genreScore = float64(genreCount) / float64(totalLikes)
        }
        
        // Combine scores
        // Lebih berat ke genre (0.8) dibanding popularity (0.2)
        score = (popularityScore * 0.2) + (genreScore * 0.8)
        
        if score > 0.1 { // Threshold to avoid very low scores
            scores = append(scores, models.RecommendationScore{
                Song:      song,
                Score:     score,
                ScoreType: "collaborative",
            })
        }
    }
    
    // Sort by score (descending)
    sort.Slice(scores, func(i, j int) bool {
        return scores[i].Score > scores[j].Score
    })
    
    // Return top N recommendations
    if len(scores) > limit {
        scores = scores[:limit]
    }
    
    return scores, nil
}