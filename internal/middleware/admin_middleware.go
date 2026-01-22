package middleware

import (
	"back_music/internal/repository"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AdminMiddleware(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ⭐⭐ PERBAIKAN: Dapatkan user ID dari token/JWT
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token (format: "Bearer <token>")
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(401, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Get user ID from claims
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(401, gin.H{"error": "User ID not found in token"})
			c.Abort()
			return
		}

		userID := uint(userIDFloat)

		// ⭐⭐ PERBAIKAN: Set user ID ke context dengan KEY YANG TEPAT
		c.Set("user_id", userID)
		c.Set("admin_id", userID) // Tambah ini jika handler pakai "admin_id"

		log.Printf("🔐 AdminMiddleware: User %d authenticated", userID)

		// Check if user is admin
		user, err := userRepo.FindUserByID(userID)
		if err != nil || user.Role != "admin" {
			c.JSON(403, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		log.Printf("👑 AdminMiddleware: User %d is admin (%s)", userID, user.Username)
		c.Next()
	}
}