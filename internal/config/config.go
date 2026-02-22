package config

import (
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
    DBSSLMode  string
    
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
    
    // Check environment
    env := getEnv("ENV", "development") // default to development
    
    // Default tuning:
    // - SIMILARITY_THRESHOLD diturunkan agar content-based lebih variatif
    // - CONTENT_WEIGHT dibuat lebih besar agar hybrid/smart-hybrid lebih terasa beda dari pure collaborative
    similarityThreshold, _ := strconv.ParseFloat(getEnv("SIMILARITY_THRESHOLD", "0.6"), 64)
    contentWeight, _ := strconv.ParseFloat(getEnv("CONTENT_WEIGHT", "0.7"), 64)
    collaborativeWeight, _ := strconv.ParseFloat(getEnv("COLLABORATIVE_WEIGHT", "0.3"), 64)
    
    // Set DB defaults based on environment
    var dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode string
    if env == "production" {
        // Production defaults (Railway/Supabase)
        dbHost = getEnv("DB_HOST", "")
        dbPort = getEnv("DB_PORT", "5432")
        dbUser = getEnv("DB_USER", "")
        dbPassword = getEnv("DB_PASSWORD", "")
        dbName = getEnv("DB_NAME", "")
        dbSSLMode = getEnv("DB_SSLMODE", "require")
    } else {
        // Development defaults (local)
        dbHost = getEnv("DB_HOST", "localhost")
        dbPort = getEnv("DB_PORT", "5432")
        dbUser = getEnv("DB_USER", "postgres")
        dbPassword = getEnv("DB_PASSWORD", "password")
        dbName = getEnv("DB_NAME", "music_app")
        dbSSLMode = getEnv("DB_SSLMODE", "disable")
    }
    
    GlobalConfig = &Config{
        SpotifyClientID:     getEnv("SPOTIFY_CLIENT_ID", ""),
        SpotifyClientSecret: getEnv("SPOTIFY_CLIENT_SECRET", ""),
        RedirectURI:         getEnv("REDIRECT_URI", "http://localhost:8080/callback"),
        
        DBHost:     dbHost,
        DBPort:     dbPort,
        DBUser:     dbUser,
        DBPassword: dbPassword,
        DBName:     dbName,
        DBSSLMode:  dbSSLMode,
        
        ServerPort: getEnv("SERVER_PORT", "8080"),
        JWTSecret:  getEnv("JWT_SECRET", "default-jwt-secret-change-in-production"),
        
        SimilarityThreshold:  similarityThreshold,
        ContentWeight:       contentWeight,
        CollaborativeWeight: collaborativeWeight,
    }
    
    if GlobalConfig.SpotifyClientID == "" || GlobalConfig.SpotifyClientSecret == "" {
    log.Println("⚠️ Spotify API credentials not set, Spotify features disabled")
}

    return nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}