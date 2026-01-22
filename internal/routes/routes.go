package routes

import (
	"back_music/internal/handlers"
	"back_music/internal/middleware"
	"back_music/internal/repository"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	authHandler *handlers.AuthHandler,
	songHandler *handlers.SongHandler,
	recommendationHandler *handlers.RecommendationHandler,
	userRepo repository.UserRepository, 
) *gin.Engine {
	router := gin.Default()

	// 1. CORS MIDDLEWARE
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5174", 
			"http://127.0.0.1:5174", 
			"http://192.168.1.7:5174",
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 2. Global middleware lainnya
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.OptionalJWTMiddleware())

	// API routes
	api := router.Group("/api")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)

			protectedAuth := auth.Group("/")
			protectedAuth.Use(middleware.JWTMiddleware())
			{
				protectedAuth.GET("/me", authHandler.Me)
			}
		}

		// PUBLIC SONG ROUTES
		songs := api.Group("/songs")
		{
			songs.GET("", songHandler.GetAllSongs)
			songs.GET("/search", songHandler.SearchSongs)
			songs.GET("/:id", songHandler.GetSongByID)
			songs.GET("/:id/audio", songHandler.GetAudioSource)
			songs.GET("/popular-id", songHandler.GetPopularIndonesianSongs)
			songs.POST("/seed", songHandler.SeedSongs)
			songs.GET("/:id/source", songHandler.GetAudioSource)
		}

		// PROTECTED ROUTES
		protected := api.Group("/")
		protected.Use(middleware.JWTMiddleware())
		{
			// User actions
			user := protected.Group("/user")
			{
				user.POST("/like/:song_id", songHandler.LikeSong)
				user.DELETE("/like/:song_id", songHandler.UnlikeSong)
				user.POST("/play/:song_id", songHandler.PlaySong)
				user.GET("/likes", songHandler.GetUserLikes)
				user.GET("/plays", songHandler.GetUserPlays)
			}

			// Recommendations
			recommendations := protected.Group("/recommendations")
			{
				recommendations.GET("/content/:song_id", recommendationHandler.GetContentBasedRecommendations)
				recommendations.GET("/collaborative", recommendationHandler.GetCollaborativeRecommendations)
				recommendations.GET("/hybrid", recommendationHandler.GetHybridRecommendations)
				recommendations.GET("/smart-hybrid", recommendationHandler.GetSmartHybridRecommendations)
				recommendations.GET("/popular", recommendationHandler.GetPopularSongs)
			}

			// ⭐⭐ PERBAIKAN: Admin routes dengan parameter yang benar
			// admin := protected.Group("/admin")
			// admin.Use(middleware.AdminMiddleware(userRepo)) // ⭐⭐ TAMBAH PARAMETER userRepo
			// {
			// 	admin.POST("/songs/:song_id/upload", songHandler.UploadCustomMP3)
			// 	admin.GET("/uploads/stats", func(c *gin.Context) {
			// 		c.JSON(200, gin.H{
			// 			"status": "success",
			// 			"data": gin.H{
			// 				"message": "Admin upload stats endpoint",
			// 				"admin_id": c.GetUint("admin_id"),
			// 			},
			// 		})
			// 	})
			// }
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "success", "message": "Server is running"})
	})

	// Root endpoint info
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"message": "Back Music API",
			"version": "1.0.0",
			"endpoints": gin.H{
				"auth": "/api/auth",
				"songs": "/api/songs",
				"recommendations": "/api/recommendations (protected)",
				"admin": "/api/admin (admin only)",
				"health": "/health",
			},
		})
	})

	return router
}