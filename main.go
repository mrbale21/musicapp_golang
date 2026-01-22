// main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"back_music/internal/config"
	"back_music/internal/database"
	"back_music/internal/handlers"
	"back_music/internal/repository"
	"back_music/internal/routes"
	"back_music/internal/services"
)

func main() {
	// Load configuration
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	if err := database.ConnectDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// log.Println("🔄 Running database migration...")
	// database.RunMigrations() // Atau database.AutoMigrate() sesuai kode Anda
	// log.Println("✅ Database migration completed")

	// Initialize repositories
	userRepo := repository.NewUserRepository()
	songRepo := repository.NewSongRepository()

	// Initialize services
	spotifyService := services.NewSpotifyService(songRepo)
	
	// ⭐⭐ BARU: Initialize Upload Service
	// uploadService, err := services.NewUploadService(songRepo)
	// if err != nil {
	// 	log.Printf("⚠️ Warning: Upload service initialization failed: %v", err)
	// 	log.Println("ℹ️  MP3 upload feature will be disabled")
	// } else {
	// 	log.Println("✅ Upload service initialized (Cloudinary)")
	// }

	// Other services
	contentService := services.NewContentBasedService(songRepo)
	collaborativeService := services.NewCollaborativeService(userRepo, songRepo)
	hybridService := services.NewHybridService(contentService, collaborativeService)

	smartHybridService := services.NewSmartHybridService(
		contentService,
		collaborativeService,
		hybridService,
	)
	log.Println("✅ Smart Hybrid service initialized")

	youtubeSvc := services.NewYouTubeService()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo)
	songHandler := handlers.NewSongHandler(songRepo, userRepo, spotifyService, youtubeSvc)
	recommendationHandler := handlers.NewRecommendationHandler(
		contentService,
		collaborativeService,
		hybridService,
		smartHybridService,
		 database.DB,
		songRepo, 
	)

	// ⭐⭐ PERBAIKAN: Setup routes dengan 4 parameter
	router := routes.SetupRoutes(
		authHandler, 
		songHandler, 
		recommendationHandler,
		userRepo, // ⭐⭐ PARAMETER KE-4 YANG DIBUTUHKAN
	)

	// Bind ke semua interface
	port := config.GlobalConfig.ServerPort
	if port == "" {
		port = "8080"
	}
	bindAddr := "0.0.0.0:" + port

	// Print startup info
	log.Println("🎵 =======================================")
	log.Println("🎵   BACK MUSIC API SERVER")
	log.Println("🎵 =======================================")
	log.Printf("🎵   Port: %s", bindAddr)
	log.Println("🎵   Features:")
	log.Println("🎵   - Spotify Integration")
	log.Println("🎵   - Indonesian Popular Songs")
	// if uploadService != nil {
	// 	log.Println("🎵   - Custom MP3 Upload (Cloudinary)")
	// }
	log.Println("🎵   - Content-Based Recommendations")
	log.Println("🎵   - Collaborative Filtering")
	log.Println("🎵   - Hybrid Recommendations")
	log.Println("🎵   - Admin Role-based Access Control")
	log.Println("🎵 =======================================")

	// Create server
	server := &http.Server{
		Addr:         bindAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("🚀 Server starting on %s", bindAddr)
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited properly")
}