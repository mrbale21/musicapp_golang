package handlers

import (
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"back_music/internal/database"
	"back_music/internal/models"
	"back_music/internal/repository"
	"back_music/internal/services"
)

type SongHandler struct {
    songRepo  repository.SongRepository
    userRepo  repository.UserRepository
    spotifyService services.SpotifyService
    youtubeService services.YouTubeService
}



func NewSongHandler(songRepo repository.SongRepository, userRepo repository.UserRepository, spotifyService services.SpotifyService, youtubeService services.YouTubeService, ) *SongHandler {
    return &SongHandler{
        songRepo:       songRepo,
        userRepo:       userRepo,
        spotifyService: spotifyService,
        // uploadService:   uploadService,  
        youtubeService: youtubeService,
    }
}

// handlers/songHandler.go
func (h *SongHandler) GetAllSongs(c *gin.Context) {
    userID := c.GetUint("user_id")
    
    var songs []models.Song
    var err error
    
    // LOG REQUEST
    log.Printf("[Handler GetAllSongs] Request from UserID: %d", userID)
    
    if userID > 0 {
        songs, err = h.songRepo.GetAllSongsWithLikeStatus(userID)
        log.Printf("[Handler] Used GetAllSongsWithLikeStatus, got %d songs", len(songs))
    } else {
        songs, err = h.songRepo.GetAllSongs()
        log.Printf("[Handler] Used GetAllSongs, got %d songs", len(songs))
    }
    
    if err != nil {
        log.Printf("[Handler GetAllSongs] ERROR: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Failed to fetch songs",
        })
        return
    }
    
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Songs fetched successfully",
        "data": gin.H{
            "songs": songs,
            "metadata": gin.H{
                "total": len(songs),
            },
        },
    })
}

func (h *SongHandler) SearchSongs(c *gin.Context) {
    query := c.Query("q")
    limitStr := c.DefaultQuery("limit", "20")
    
    limit, err := strconv.Atoi(limitStr)
    if err != nil || limit <= 0 {
        limit = 20
    }
    
    if query == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "Search query is required",
        })
        return
    }
    
    userID := c.GetUint("user_id")
    var songs []models.Song
    
    if userID > 0 {
        songs, err = h.songRepo.SearchSongsWithLikeStatus(query, limit, userID)
    } else {
        songs, err = h.songRepo.SearchSongs(query, limit)
    }
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Failed to search songs",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Search completed",
        "data":    songs,
    })
}

func (h *SongHandler) GetSongByID(c *gin.Context) {
    songID := c.Param("id")
    
    // Validate UUID format to prevent invalid UUID errors in DB
    if _, err := uuid.Parse(songID); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "Invalid song ID format",
        })
        return
    }
    userID := c.GetUint("user_id")
    
    song, err := h.songRepo.GetSongByID(songID)
    if err != nil {
        if errors.Is(err, repository.ErrSongNotFound) {
            c.JSON(http.StatusNotFound, gin.H{
                "status":  "error",
                "message": "Song not found",
            })
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Failed to fetch song",
        })
        return
    }

    if userID > 0 {
        isLiked, _ := h.songRepo.IsSongLikedByUser(songID, userID)
        song.IsLiked = isLiked
    }

    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Song fetched successfully",
        "data":    song,
    })
}

func (h *SongHandler) SeedSongs(c *gin.Context) {
    limitStr := c.DefaultQuery("limit", "100")
    
    limit, err := strconv.Atoi(limitStr)
    if err != nil || limit <= 0 {
        limit = 100
    }
    
    if limit > 200 {
        limit = 200 // Safety limit
    }
    
    go func() {
        if err := h.spotifyService.SeedSongsFromSpotify(limit); err != nil {
            // Log error but don't return to client since this is async
        }
    }()
    
    c.JSON(http.StatusAccepted, gin.H{
        "status":  "success",
        "message": "Seeding process started in background",
        "data": gin.H{
            "limit": limit,
        },
    })
}

func (h *SongHandler) LikeSong(c *gin.Context) {
    userID := c.GetUint("user_id")
    songID := c.Param("song_id")
    
    // Validate UUID format before using in queries
    if _, err := uuid.Parse(songID); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "Invalid song ID format",
        })
        return
    }
    
    _, err := h.songRepo.GetSongByID(songID)
    if err != nil {
        if errors.Is(err, repository.ErrSongNotFound) {
            c.JSON(http.StatusNotFound, gin.H{
                "status":  "error",
                "message": "Song not found",
            })
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch song"})
        return
    }

    // Check if already liked
    var existingLike models.UserLike
    err = database.DB.Where("user_id = ? AND song_id = ?", userID, songID).First(&existingLike).Error
    if err == nil {
        c.JSON(http.StatusConflict, gin.H{
            "status":  "error",
            "message": "Song already liked",
        })
        return
    }
    
    // Create like
    like := models.UserLike{
        UserID: userID,
        SongID: songID,
    }
    
    if err := database.DB.Create(&like).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Failed to like song",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Song liked successfully",
        "data":    like,
    })
}

func (h *SongHandler) UnlikeSong(c *gin.Context) {
    userID := c.GetUint("user_id")
    songID := c.Param("song_id")
    
    // Validate UUID format before using in queries
    if _, err := uuid.Parse(songID); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "Invalid song ID format",
        })
        return
    }
    
    result := database.DB.Where("user_id = ? AND song_id = ?", userID, songID).Delete(&models.UserLike{})
    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Failed to unlike song",
        })
        return
    }
    
    if result.RowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{
            "status":  "error",
            "message": "Like not found",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Song unliked successfully",
    })
}

func (h *SongHandler) PlaySong(c *gin.Context) {
    userID := c.GetUint("user_id")
    songID := c.Param("song_id")
    
    // Validate UUID format before using in queries
    if _, err := uuid.Parse(songID); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "Invalid song ID format",
        })
        return
    }
    
    _, err := h.songRepo.GetSongByID(songID)
    if err != nil {
        if errors.Is(err, repository.ErrSongNotFound) {
            c.JSON(http.StatusNotFound, gin.H{
                "status":  "error",
                "message": "Song not found",
            })
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch song"})
        return
    }

    // Find existing play record
    var play models.UserPlay
    err = database.DB.Where("user_id = ? AND song_id = ?", userID, songID).First(&play).Error
    
    if err != nil {
        // Create new play record
        play = models.UserPlay{
            UserID:     userID,
            SongID:     songID,
            PlayCount:  1,
            LastPlayed: time.Now(),
        }
        
        if err := database.DB.Create(&play).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "status":  "error",
                "message": "Failed to record play",
            })
            return
        }
    } else {
        // Update existing play record
        play.PlayCount++
        play.LastPlayed = time.Now()
        
        if err := database.DB.Save(&play).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "status":  "error",
                "message": "Failed to update play count",
            })
            return
        }
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Play recorded successfully",
        "data":    play,
    })
}

func (h *SongHandler) GetUserLikes(c *gin.Context) {
    userID := c.GetUint("user_id")
    
    var likes []models.UserLike
    if err := database.DB.Preload("Song").Where("user_id = ?", userID).Find(&likes).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Failed to fetch likes",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Likes fetched successfully",
        "data":    likes,
    })
}

func (h *SongHandler) GetUserPlays(c *gin.Context) {
    userID := c.GetUint("user_id")
    
    var plays []models.UserPlay
    if err := database.DB.Preload("Song").Where("user_id = ?", userID).Find(&plays).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Failed to fetch plays",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Plays fetched successfully",
        "data":    plays,
    })
}

// GetPopularSongs handler di songHandler.go
func (h *SongHandler) GetPopularSongs(c *gin.Context) {
    limitStr := c.DefaultQuery("limit", "20")
    userID := c.GetUint("user_id")
    
    limit, err := strconv.Atoi(limitStr)
    if err != nil || limit <= 0 {
        limit = 20
    }
    
    songs, err := h.songRepo.GetPopularSongs(limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Failed to fetch popular songs",
            "error":   err.Error(),
        })
        return
    }
    
    // Check like status jika user logged in
    if userID > 0 && len(songs) > 0 {
        // Get all song IDs
        songIDs := make([]string, len(songs))
        for i, song := range songs {
            songIDs[i] = song.ID
        }
        
        // Query liked songs
        var likedSongIDs []string
        database.DB.Model(&models.UserLike{}).
            Where("user_id = ? AND song_id IN ?", userID, songIDs).
            Pluck("song_id", &likedSongIDs)
        
        // Create lookup map
        likedMap := make(map[string]bool)
        for _, id := range likedSongIDs {
            likedMap[id] = true
        }
        
        // Set IsLiked field
        for i := range songs {
            songs[i].IsLiked = likedMap[songs[i].ID]
        }
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Popular songs fetched successfully",
        "data": gin.H{
            "songs": songs,
            "limit": limit,
            "total": len(songs),
            "order_by": "popularity_desc",
        },
    })
}

//  TAMBAHAN: Get Popular Indonesian Songs
func (h *SongHandler) GetPopularIndonesianSongs(c *gin.Context) {
    limitStr := c.DefaultQuery("limit", "50")
    limit, _ := strconv.Atoi(limitStr)
    
    if limit > 100 {
        limit = 100 // Max limit
    }
    
    songs, err := h.spotifyService.GetPopularIndonesianSongs(limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Failed to fetch popular Indonesian songs",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Popular Indonesian songs fetched successfully",
        "data": gin.H{
            "total":   len(songs),
            "songs":   songs,
            "filters": map[string]string{
                "country": "Indonesia",
                "sort":    "popularity",
            },
        },
    })
}



// func (h *SongHandler) UploadCustomMP3(c *gin.Context) {
//     // 1. Auth check
//     val, exists := c.Get("user_id")
//     if !exists {
//         c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
//         return
//     }
//     userID := val.(uint)

//     // 2. Get song_id
//     songID := c.Param("song_id")
    
//     // 3. Ambil file dari form (Key: mp3_file)
//     fileHeader, err := c.FormFile("mp3_file")
//     if err != nil {
//         log.Printf("[Upload Error] Get file dari form gagal: %v", err)
//         c.JSON(http.StatusBadRequest, gin.H{"error": "File field 'mp3_file' is required"})
//         return
//     }

//     log.Printf("[Upload Debug] File Name: %s, Size: %d bytes", fileHeader.Filename, fileHeader.Size)

//     // Validasi Ekstensi
//     if !strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".mp3") {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "Only MP3 files are allowed"})
//         return
//     }

//     // 4. Buka file sebagai Stream (io.Reader)
//     // HAPUS io.ReadAll(file) karena ini yang bikin error tipe data
//     file, err := fileHeader.Open()
//     if err != nil {
//         c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
//         return
//     }
//     defer file.Close()

//     // 5. Upload ke Cloudinary lewat Service (Langsung oper 'file')
//     // 'file' di sini bertipe multipart.File yang sudah mengimplementasikan io.Reader
//     mp3URL, err := h.uploadService.UploadMP3ToCloudinary(file, fileHeader.Filename, songID)
//     if err != nil {
//         log.Printf("Upload error: %v", err)
//         c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//         return
//     }

//     // 6. Update DB
//     err = h.uploadService.UpdateSongWithCustomMP3(songID, mp3URL)
//     if err != nil {
//         c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update database"})
//         return
//     }

//     updatedSong, err := h.songRepo.GetSongByID(songID)
//     if err != nil {
//         c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated song"})
//         return
//     }

//     c.JSON(http.StatusOK, gin.H{
//         "status":  "success",
//         "message": "MP3 uploaded successfully",
//         "data": gin.H{
//             "song_id":     songID,
//             "mp3_url":     mp3URL,
//             "uploaded_by": userID,
//             "song":        updatedSong,
//         },
//     })
// }

func getFieldNames(files map[string][]*multipart.FileHeader) []string {
	var names []string
	for name := range files {
		names = append(names, name)
	}
	return names
}

func (h *SongHandler) GetAudioSource(c *gin.Context) {
    songID := c.Param("id")

    // Validate UUID format before fetching from repository
    if _, err := uuid.Parse(songID); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "Invalid song ID format",
        })
        return
    }

    song, err := h.songRepo.GetSongByID(songID)
    if err != nil {
        if errors.Is(err, repository.ErrSongNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Lagu tidak ditemukan"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch song"})
        return
    }

    if song.YoutubeID != "" {
        c.JSON(http.StatusOK, gin.H{
            "status": "success",
            "data": gin.H{
                "video_id": song.YoutubeID,
                "source":   "database_cache",
            },
        })
        return
    }

    searchQuery := fmt.Sprintf("%s - %s", song.Artist, song.Title)
    videoID, err := h.youtubeService.SearchAudio(searchQuery)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Gagal mencari audio otomatis"})
        return
    }

    song.YoutubeID = videoID
    if err := h.songRepo.UpdateSong(song); err != nil {
        log.Printf("[GetAudioSource] failed to cache YoutubeID for song %s: %v", songID, err)
    }

    c.JSON(http.StatusOK, gin.H{
        "status": "success",
        "data": gin.H{
            "video_id": videoID,
            "source":   "youtube_api_fresh",
        },
    })
}