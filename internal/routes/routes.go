package routes

import (
	"os"
	"time"

	"back_music/internal/handlers"
	"back_music/internal/middleware"
	"back_music/internal/repository"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	authHandler *handlers.AuthHandler,
	songHandler *handlers.SongHandler,
	recommendationHandler *handlers.RecommendationHandler,
	userRepo repository.UserRepository,
) *gin.Engine {

	router := gin.New()

	// =========================
	// GLOBAL MIDDLEWARE
	// =========================
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// =========================
	// CORS CONFIG (DEV / PROD)
	// =========================
	env := os.Getenv("ENV") // development | production
	frontendURL := os.Getenv("CORS_ORIGIN") // https://musicapp-frontend.vercel.app

	corsConfig := cors.Config{
        // JANGAN set AllowAllOrigins: true jika ingin pakai AllowCredentials
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }

    if env == "production" {
        // 🔒 PROD MODE: Menggunakan URL Vercel dari Railway
        // Jika frontendURL kosong, sebaiknya beri fallback atau log
        if frontendURL != "" {
            corsConfig.AllowOrigins = []string{frontendURL}
        } else {
            // Fallback jika lupa set env di production (opsional)
            corsConfig.AllowOrigins = []string{"https://musicapp-gules-pi.vercel.app"}
        }
    } else {
        // 🔓 DEV MODE: Anti CORS untuk lokal/mobile/network
        corsConfig.AllowOrigins = []string{
            "http://localhost:3000",
            "http://localhost:3001",
            "http://localhost:5173", // Vite default
            "http://127.0.0.1:3000",
            "http://127.0.0.1:3001",
            "http://127.0.0.1:5173",
        }
        // Untuk network access (HP di IP lokal), allow semua origin yang valid
        corsConfig.AllowOriginFunc = func(origin string) bool {
            // Allow localhost dan IP lokal network
            return true // Lebih fleksibel untuk development
        }
    }

	router.Use(cors.New(corsConfig))

	// =========================
	// API ROUTES
	// =========================
	api := router.Group("/api")
	{
		// ---------- AUTH ----------
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)

			authProtected := auth.Group("/")
			authProtected.Use(middleware.JWTMiddleware())
			{
				authProtected.GET("/me", authHandler.Me)
			}
		}

		// ---------- PUBLIC SONGS (optional JWT for like status when logged in) ----------
		songs := api.Group("/songs")
		songs.Use(middleware.OptionalJWTMiddleware())
		{
			songs.GET("", songHandler.GetAllSongs)
			songs.GET("/search", songHandler.SearchSongs)
			songs.GET("/popular-id", songHandler.GetPopularIndonesianSongs)
			songs.POST("/seed", songHandler.SeedSongs)
			songs.GET("/:id", songHandler.GetSongByID)
			songs.GET("/:id/audio", songHandler.GetAudioSource)
			songs.GET("/:id/source", songHandler.GetAudioSource)
		}

		// ---------- PROTECTED ----------
		protected := api.Group("/")
		protected.Use(middleware.JWTMiddleware())
		{
			// USER
			user := protected.Group("/user")
			{
				user.POST("/like/:song_id", songHandler.LikeSong)
				user.DELETE("/like/:song_id", songHandler.UnlikeSong)
				user.POST("/play/:song_id", songHandler.PlaySong)
				user.GET("/likes", songHandler.GetUserLikes)
				user.GET("/plays", songHandler.GetUserPlays)
			}

			// RECOMMENDATIONS
			recommendations := protected.Group("/recommendations")
			{
				recommendations.GET("/content/:song_id", recommendationHandler.GetContentBasedRecommendations)
				recommendations.GET("/collaborative", recommendationHandler.GetCollaborativeRecommendations)
				recommendations.GET("/hybrid", recommendationHandler.GetHybridRecommendations)
				recommendations.GET("/smart-hybrid", recommendationHandler.GetSmartHybridRecommendations)
				recommendations.GET("/popular", recommendationHandler.GetPopularSongs)
			}

			// ADMIN (OPTIONAL)
			// admin := protected.Group("/admin")
			// admin.Use(middleware.AdminMiddleware(userRepo))
			// {
			// 	admin.POST("/songs/:song_id/upload", songHandler.UploadCustomMP3)
			// }
		}
	}

	// =========================
	// HEALTH & ROOT
	// =========================
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"message": "Server is running",
		})
	})

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "Back Music API",
			"version": "1.0.0",
		})
	})

	return router
}
