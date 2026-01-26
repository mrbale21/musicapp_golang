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
	// LOAD CONFIG
	// =========================
	if err := config.LoadConfig(); err != nil {
		log.Println("⚠️ Config load warning:", err)
		log.Println("⚠️ Using environment variables only")
	}

	// =========================
	// CONNECT DATABASE
	// =========================
	if err := database.ConnectDB(); err != nil {
		log.Println("❌ Database connection failed:", err)
		log.Println("❌ App WILL NOT start without database")
		// os.Exit(1) 
	}

	// =========================
	// DB KEEP ALIVE (SAFE)
	// =========================
	go func() {
		sqlDB, err := database.DB.DB()
		if err != nil {
			log.Println("⚠️ Cannot access sql.DB:", err)
			return
		}

		for {
			if err := sqlDB.Ping(); err != nil {
				log.Println("⚠️ Database ping failed:", err)
			}
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
	spotifyService := services.NewSpotifyService(songRepo)

	contentService := services.NewContentBasedService(songRepo)
	collaborativeService := services.NewCollaborativeService(userRepo, songRepo)
	hybridService := services.NewHybridService(contentService, collaborativeService)

	smartHybridService := services.NewSmartHybridService(
		contentService,
		collaborativeService,
		hybridService,
	)

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
	// PORT (RAILWAY)
	// =========================
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := "0.0.0.0:" + port

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// =========================
	// START SERVER (BLOCKING)
	// =========================
	log.Println("🎵 =======================================")
	log.Println("🎵   BACK MUSIC API SERVER")
	log.Println("🎵 =======================================")
	log.Printf("🎵   Running on: %s", addr)
	log.Println("🎵   Environment: Production (Railway)")
	log.Println("🎵 =======================================")
	log.Println("🚀 Server started")

	// Graceful shutdown listener
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		log.Println("🛑 Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Println("❌ Forced shutdown:", err)
		}
	}()

	// ⛔ BLOCK DI SINI (PENTING)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("❌ Server crashed:", err)
	}

	log.Println("✅ Server exited properly")
}
