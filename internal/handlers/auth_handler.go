package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"back_music/internal/config"
	"back_music/internal/models"
	"back_music/internal/repository"
)

type AuthHandler struct {
    userRepo repository.UserRepository
    config   *config.Config
}

func NewAuthHandler(userRepo repository.UserRepository) *AuthHandler {
    return &AuthHandler{
        userRepo: userRepo,
        config:   config.GlobalConfig,
    }
}

func (h *AuthHandler) Register(c *gin.Context) {
    var req models.UserRegister
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "Invalid request body",
            "error":   err.Error(),
        })
        return
    }
    
    // Check if user already exists
     existingUser, err := h.userRepo.FindUserByEmail(req.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Database error",
            "error":   err.Error(),
        })
        return
    }
    
    if existingUser != nil {
        c.JSON(http.StatusConflict, gin.H{
            "status":  "error",
            "message": "User already exists",
        })
        return
    }
    
    // Hash password
    hashedPassword, err := h.userRepo.HashPassword(req.Password)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Failed to process password",
        })
        return
    }
    
    // Create user
    user := &models.User{
        Username: req.Username,
        Email:    req.Email,
        Password: hashedPassword,
        Role:   "user",
    }
    
    if err := h.userRepo.CreateUser(user); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Failed to create user",
        })
        return
    }
    
    // Generate JWT token
    token, err := h.generateJWT(user.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Failed to generate token",
        })
        return
    }
    
    user.Password = "" // Don't send password back
    
    c.JSON(http.StatusCreated, gin.H{
        "status":  "success",
        "message": "User registered successfully",
        "data": models.AuthResponse{
            Token: token,
            User:  *user,
        },
    })
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req models.UserLogin
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "Invalid request body",
        })
        return
    }
    
    // Find user
    user, err := h.userRepo.FindUserByEmail(req.Email)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "error",
            "message": "Invalid credentials",
        })
        return
    }
    
    // Verify password
    if err := h.userRepo.VerifyPassword(user.Password, req.Password); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "error",
            "message": "Invalid credentials",
        })
        return
    }
    
    // Generate JWT token
    token, err := h.generateJWT(user.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Failed to generate token",
        })
        return
    }
    
    user.Password = "" // Don't send password back
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Login successful",
        "data": models.AuthResponse{
            Token: token,
            User:  *user,
        },
    })
}

func (h *AuthHandler) Me(c *gin.Context) {
    // Get user ID from JWT middleware
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "error",
            "message": "User not authenticated",
        })
        return
    }

    // Convert to uint
    uid, ok := userID.(uint)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "error",
            "message": "Invalid user ID format",
        })
        return
    }

    // Get user from database
    user, err := h.userRepo.FindUserByID(uid)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Failed to fetch user data",
            "error":   err.Error(),
        })
        return
    }

    if user == nil {
        c.JSON(http.StatusNotFound, gin.H{
            "status":  "error",
            "message": "User not found",
        })
        return
    }

    // Clear sensitive data
    user.Password = ""
    
    // Optionally clear other sensitive fields if needed
    if user.DeletedAt.Valid {
        user.DeletedAt = gorm.DeletedAt{} // Hide soft delete info
    }

    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "User data retrieved successfully",
        "data":    user,
    })
}

func (h *AuthHandler) generateJWT(userID uint) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
        "iat":     time.Now().Unix(),
    })
    
    return token.SignedString([]byte(h.config.JWTSecret))
}