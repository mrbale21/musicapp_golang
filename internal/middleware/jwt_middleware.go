package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"back_music/internal/config"
)

func JWTMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "status":  "error",
                "message": "Authorization header required",
            })
            c.Abort()
            return
        }
        
        // Check if Bearer token
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "status":  "error",
                "message": "Invalid authorization format",
            })
            c.Abort()
            return
        }
        
        tokenString := parts[1]
        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "status":  "error",
                "message": "Token is empty",
            })
            c.Abort()
            return
        }
        
        // Parse and validate token
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, jwt.ErrSignatureInvalid
            }
            
            // Validate JWT secret is set
            if config.GlobalConfig.JWTSecret == "" {
                return nil, jwt.ErrSignatureInvalid
            }
            
            return []byte(config.GlobalConfig.JWTSecret), nil
        })
        
        if err != nil {
            // Check if token is expired
            if ve, ok := err.(*jwt.ValidationError); ok {
                if ve.Errors&jwt.ValidationErrorExpired != 0 {
                    c.JSON(http.StatusUnauthorized, gin.H{
                        "status":  "error",
                        "message": "Token has expired",
                    })
                } else {
                    c.JSON(http.StatusUnauthorized, gin.H{
                        "status":  "error",
                        "message": "Invalid token",
                    })
                }
            } else {
                c.JSON(http.StatusUnauthorized, gin.H{
                    "status":  "error",
                    "message": "Token validation failed",
                })
            }
            c.Abort()
            return
        }
        
        if !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{
                "status":  "error",
                "message": "Invalid token",
            })
            c.Abort()
            return
        }
        
        // Extract claims
        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            // Set user ID in context
            userID, ok := claims["user_id"].(float64)
            if !ok {
                c.JSON(http.StatusUnauthorized, gin.H{
                    "status":  "error",
                    "message": "Invalid token claims: user_id not found",
                })
                c.Abort()
                return
            }
            
            c.Set("user_id", uint(userID))
        } else {
            c.JSON(http.StatusUnauthorized, gin.H{
                "status":  "error",
                "message": "Invalid token claims",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

func OptionalJWTMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.Next()
            return
        }
        
        // Check if Bearer token
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.Next()
            return
        }
        
        tokenString := parts[1]
        
        // Try to parse token
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, jwt.ErrSignatureInvalid
            }
            return []byte(config.GlobalConfig.JWTSecret), nil
        })
        
        if err != nil || !token.Valid {
            c.Next()
            return
        }
        
        // Extract claims if valid
        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            userID, ok := claims["user_id"].(float64)
            if ok {
                c.Set("user_id", uint(userID))
            }
        }
        
        c.Next()
    }
}
