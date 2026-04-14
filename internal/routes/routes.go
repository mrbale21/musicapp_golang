package routes

import (
	"os"
	"strings"
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
        // 🔒 PROD MODE: Ketat tapi flexible untuk testing
        corsConfig.AllowOriginFunc = func(origin string) bool {
            // Allow specific production domain jika di-set
            if frontendURL != "" {
                return origin == frontendURL
            }
            // Allow common production URLs & ngrok for testing
            return strings.Contains(origin, "ngrok") || 
                   strings.Contains(origin, "vercel.app") ||
                   strings.Contains(origin, "localhost") ||
                   strings.HasPrefix(origin, "http://localhost")
        }
    } else {
        // 🔓 DEV MODE: Accept semua origin
        corsConfig.AllowOriginFunc = func(origin string) bool {
            return true
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

	// Debug endpoint untuk cek routes
	router.GET("/api/debug/routes", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"message": "API routes working",
			"endpoints": gin.H{
				"auth": gin.H{
					"register": "POST /api/auth/register",
					"login":    "POST /api/auth/login",
					"me":       "GET /api/auth/me (protected)",
				},
				"songs": gin.H{
					"list":   "GET /api/songs",
					"search": "GET /api/songs/search",
				},
			},
		})
	})

	// Test POST endpoint
	router.POST("/api/debug/test-post", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"message": "POST endpoint working",
			"content_type": c.GetHeader("Content-Type"),
			"method": c.Request.Method,
		})
	})

	return router
}