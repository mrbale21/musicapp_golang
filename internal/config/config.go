package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
    SpotifyClientID     string
    SpotifyClientSecret string
    RedirectURI         string
    
    DBHost     string
    DBPort     string
    DBUser     string
    DBPassword string
    DBName     string
    
    ServerPort string
    JWTSecret  string
    
    SimilarityThreshold  float64
    ContentWeight       float64
    CollaborativeWeight float64
}

var GlobalConfig *Config

func LoadConfig() error {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using environment variables")
    }
    
    similarityThreshold, _ := strconv.ParseFloat(getEnv("SIMILARITY_THRESHOLD", "0.7"), 64)
    contentWeight, _ := strconv.ParseFloat(getEnv("CONTENT_WEIGHT", "0.5"), 64)
    collaborativeWeight, _ := strconv.ParseFloat(getEnv("COLLABORATIVE_WEIGHT", "0.5"), 64)
    
    GlobalConfig = &Config{
        SpotifyClientID:     getEnv("SPOTIFY_CLIENT_ID", ""),
        SpotifyClientSecret: getEnv("SPOTIFY_CLIENT_SECRET", ""),
        RedirectURI:         getEnv("REDIRECT_URI", "http://localhost:8080/callback"),
        
        DBHost:     getEnv("DB_HOST", "localhost"),
        DBPort:     getEnv("DB_PORT", "5432"),
        DBUser:     getEnv("DB_USER", "postgres"),
        DBPassword: getEnv("DB_PASSWORD", ""),
        DBName:     getEnv("DB_NAME", "music_app"),
        
        ServerPort: getEnv("SERVER_PORT", "8080"),
        JWTSecret:  getEnv("JWT_SECRET", "default-jwt-secret-change-in-production"),
        
        SimilarityThreshold:  similarityThreshold,
        ContentWeight:       contentWeight,
        CollaborativeWeight: collaborativeWeight,
    }
    
    if GlobalConfig.SpotifyClientID == "" || GlobalConfig.SpotifyClientSecret == "" {
        return fmt.Errorf("Spotify API credentials are required")
    }
    
    return nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}