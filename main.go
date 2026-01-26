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

	// =========================
	// LOAD CONFIG (SAFE)
	// =========================
	if err := config.LoadConfig(); err != nil {
		log.Println("⚠️ Config load warning:", err)
		log.Println("⚠️ Using environment variables only")
	}

	// =========================
	// CONNECT DATABASE (SAFE)
	// =========================
	if err := database.ConnectDB(); err != nil {
		log.Println("⚠️ Database connection failed:", err)
		log.Println("⚠️ App will continue running without database")
	}

	go func() {
	sqlDB, _ := database.DB.DB()
	for {
		sqlDB.Ping()
		time.Sleep(5 * time.Minute)
	}
}()


	// =========================
	// INIT REPOSITORIES
	// =========================
	userRepo := repository.NewUserRepository()
	songRepo := repository.NewSongRepository()

	// =========================
	// INIT SERVICES
	// =========================

	// Spotify (optional)
	spotifyService := services.NewSpotifyService(songRepo)

	// Recommendation services
	contentService := services.NewContentBasedService(songRepo)
	collaborativeService := services.NewCollaborativeService(userRepo, songRepo)
	hybridService := services.NewHybridService(contentService, collaborativeService)

	smartHybridService := services.NewSmartHybridService(
		contentService,
		collaborativeService,
		hybridService,
	)
	log.Println("✅ Smart Hybrid service initialized")

	// Youtube service
	youtubeSvc := services.NewYouTubeService()

	// =========================
	// INIT HANDLERS
	// =========================
	authHandler := handlers.NewAuthHandler(userRepo)

	songHandler := handlers.NewSongHandler(
		songRepo,
		userRepo,
		spotifyService,
		youtubeSvc,
	)

	recommendationHandler := handlers.NewRecommendationHandler(
		contentService,
		collaborativeService,
		hybridService,
		smartHybridService,
		database.DB,
		songRepo,
	)

	// =========================
	// ROUTES
	// =========================
	router := routes.SetupRoutes(
		authHandler,
		songHandler,
		recommendationHandler,
		userRepo,
	)

	// =========================
	// PORT (RAILWAY SAFE)
	// =========================
	port := os.Getenv("PORT")
	if port == "" {
		port = config.GlobalConfig.ServerPort
	}
	if port == "" {
		port = "8080"
	}

	bindAddr := "0.0.0.0:" + port

	// =========================
	// SERVER CONFIG
	// =========================
	server := &http.Server{
		Addr:         bindAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// =========================
	// START SERVER
	// =========================
	go func() {
		log.Println("🎵 =======================================")
		log.Println("🎵   BACK MUSIC API SERVER")
		log.Println("🎵 =======================================")
		log.Printf("🎵   Running on: %s", bindAddr)
		log.Println("🎵   Environment: Production (Railway)")
		log.Println("🎵 =======================================")
		log.Println("🚀 Server started")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("❌ Server error:", err)
		}
	}()

	// =========================
	// GRACEFUL SHUTDOWN
	// =========================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Println("❌ Forced shutdown:", err)
	}

	log.Println("✅ Server exited properly")
}
