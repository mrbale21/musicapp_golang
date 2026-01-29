package models

import (
	"time"
)

type UserLike struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    UserID    uint      `gorm:"not null;index" json:"user_id"`
    SongID    string    `gorm:"not null;index" json:"song_id"`
    CreatedAt time.Time `json:"created_at"`
    
    // Relationships
    User User `gorm:"foreignKey:UserID" json:"-"`
    Song Song `gorm:"foreignKey:SongID" json:"song"`
}

type UserPlay struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    UserID    uint      `gorm:"not null;index" json:"user_id"`
    SongID    string    `gorm:"not null;index" json:"song_id"`
    PlayCount int       `gorm:"default:1" json:"play_count"`
    LastPlayed time.Time `json:"last_played"`
    CreatedAt time.Time `json:"created_at"`
    
    // Relationships
    User User `gorm:"foreignKey:UserID" json:"-"`
    Song Song `gorm:"foreignKey:SongID" json:"song"`
}

type RecommendationScore struct {
    Song        Song    `json:"song"`
    Score       float64 `json:"score"`
    ScoreType   string  `json:"score_type"` // "content", "collaborative", "hybrid"
    Explanation string  `json:"explanation,omitempty"` // ⭐⭐ TAMBAH INI
    Rank        int     `json:"rank,omitempty"`        // ⭐⭐ TAMBAH INI
}