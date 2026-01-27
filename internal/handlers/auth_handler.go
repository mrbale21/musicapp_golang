package handlers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

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

    user, err := h.userRepo.FindUserByEmail(req.Email)
    if err != nil || user == nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "error",
            "message": "Invalid credentials",
        })
        return
    }

    // Verify password (DEBUG TIMING)
    start := time.Now()
    if err := h.userRepo.VerifyPassword(user.Password, req.Password); err != nil {
        log.Println("bcrypt time (failed):", time.Since(start))
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "error",
            "message": "Invalid credentials",
        })
        return
    }
    log.Println("bcrypt time (success):", time.Since(start))

    token, err := h.generateJWT(user.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Failed to generate token",
        })
        return
    }

    user.Password = ""

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
    userIDRaw, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "error",
            "message": "User not authenticated",
        })
        return
    }

    var uid uint

    switch v := userIDRaw.(type) {
    case float64:
        uid = uint(v)
    case int:
        uid = uint(v)
    case int64:
        uid = uint(v)
    case uint:
        uid = v
    case uint64:
        uid = uint(v)
    default:
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "error",
            "message": "Invalid user ID type",
        })
        return
    }

    user, err := h.userRepo.FindUserByID(uid)
    if err != nil {
        if errors.Is(err, repository.ErrUserNotFound) {
            c.JSON(http.StatusUnauthorized, gin.H{
                "status":  "error",
                "message": "User not found",
            })
            return
        }
        log.Printf("[Me] FindUserByID error: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Failed to fetch user data",
        })
        return
    }

    if user == nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "error",
            "message": "User not found",
        })
        return
    }

    user.Password = ""

    c.JSON(http.StatusOK, gin.H{
        "status": "success",
        "data":   user,
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