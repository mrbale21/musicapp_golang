package routes

import (
	"log"
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
	frontendURL := os.Getenv("CORS_ORIGIN")

	corsConfig := cors.Config{
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
        ExposeHeaders:    []string{"Content-Length", "Authorization"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }

    if env == "production" {
        // 🔒 PROD MODE: Require CORS_ORIGIN to be explicitly set
        if frontendURL == "" {
            log.Fatal("❌ CORS_ORIGIN environment variable is NOT set in production!")
        }
        corsConfig.AllowOrigins = []string{frontendURL}
        log.Printf("✅ CORS configured for production: %s", frontendURL)
    } else {
        // 🔓 DEV MODE: Allow flexible origins
        allowedOrigins := []string{
            "http://localhost:3000",
            "http://localhost:3001",
            "http://localhost:5173",
            "http://127.0.0.1:3000",
            "http://127.0.0.1:3001",
            "http://127.0.0.1:5173",
        }
        
        // If CORS_ORIGIN set in dev, also add it
        if frontendURL != "" {
            allowedOrigins = append(allowedOrigins, frontendURL)
        }
        
        corsConfig.AllowOriginFunc = func(origin string) bool {
            // Allow all localhost/127.0.0.1 variants
            for _, allowed := range allowedOrigins {
                if origin == allowed {
                    return true
                }
            }
            // Allow local network IPs (192.168.x.x, 10.x.x.x)
            if strings.HasPrefix(origin, "http://192.168.") || strings.HasPrefix(origin, "http://10.") {
                return true
            }
            return false
        }
        log.Printf("✅ CORS configured for development with %d allowed origins", len(allowedOrigins))
    }


	router.Use(cors.New(corsConfig))

	// =========================
	// SECURITY HEADERS MIDDLEWARE
	// =========================
	router.Use(func(c *gin.Context) {
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")
		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		// Content Security Policy (lenient untuk API)
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	})

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
