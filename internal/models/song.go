package models

import (
	"time"
)

// package models/song.go
type Song struct {
    ID           string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
    SpotifyID    string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"spotify_id"`
    Title        string    `gorm:"type:varchar(255);not null" json:"title"`
    Artist       string    `gorm:"type:varchar(255);not null" json:"artist"`
    Album        string    `gorm:"type:varchar(255)" json:"album"`
    Genre        string    `gorm:"type:varchar(100)" json:"genre"`
    Popularity   int       `json:"popularity"`
    DurationMs   int       `json:"duration_ms"`
    Danceability float64   `gorm:"default:0" json:"danceability"`
    Energy       float64   `gorm:"default:0" json:"energy"`
    Key          int       `gorm:"default:0" json:"key"`
    Loudness     float64   `gorm:"default:0" json:"loudness"`
    Mode         int       `gorm:"default:0" json:"mode"`
    Speechiness  float64   `gorm:"default:0" json:"speechiness"`
    Acousticness float64   `gorm:"default:0" json:"acousticness"`
    Instrumentalness float64 `gorm:"default:0" json:"instrumentalness"`
    Liveness     float64   `gorm:"default:0" json:"liveness"`
    Valence      float64   `gorm:"default:0" json:"valence"`
    Tempo        float64   `gorm:"default:0" json:"tempo"`
    TimeSignature int      `gorm:"default:0" json:"time_signature"`
    IsLiked      bool      `gorm:"-" json:"is_liked"`
    PreviewURL   string    `json:"preview_url"`
    ImageURL     string    `json:"image_url"`
    CreatedAt    time.Time `json:"created_at"`
    
    // For similarity calculations
    FeatureVector []float64 `gorm:"-" json:"-"`
    YoutubeID        string    `json:"youtube_id"`
}

type AudioFeatures struct {
    Danceability    float64 `json:"danceability"`
    Energy          float64 `json:"energy"`
    Key             int     `json:"key"`
    Loudness        float64 `json:"loudness"`
    Mode            int     `json:"mode"`
    Speechiness     float64 `json:"speechiness"`
    Acousticness    float64 `json:"acousticness"`
    Instrumentalness float64 `json:"instrumentalness"`
    Liveness        float64 `json:"liveness"`
    Valence         float64 `json:"valence"`
    Tempo           float64 `json:"tempo"`
    TimeSignature   int     `json:"time_signature"`
}