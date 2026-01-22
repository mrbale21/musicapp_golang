package services

// import (
// 	"back_music/internal/models"
// 	"back_music/internal/repository"
// 	"context"
// 	"fmt"
// 	"io"
// 	"log"
// 	"os"
// 	"strings"
// 	"time"

// 	"github.com/cloudinary/cloudinary-go/v2"
// 	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
// )

// type UploadService interface {
// 	UploadMP3ToCloudinary(file io.Reader, filename string, songID string) (string, error)
// 	UpdateSongWithCustomMP3(songID string, mp3URL string) error
// 	GetCustomSongs() ([]models.Song, error)
// }

// func (u UploadService) SearchYouTube(searchQuery string) (any, error) {
// 	panic("unimplemented")
// }

// type uploadService struct {
// 	songRepo        repository.SongRepository
// 	cloudinaryCloud *cloudinary.Cloudinary
// }

// func NewUploadService(songRepo repository.SongRepository) (UploadService, error) {
// 	cloudinaryURL := fmt.Sprintf("cloudinary://%s:%s@%s",
// 		os.Getenv("CLOUDINARY_API_KEY"),
// 		os.Getenv("CLOUDINARY_API_SECRET"),
// 		os.Getenv("CLOUDINARY_CLOUD_NAME"),
// 	)

// 	cld, err := cloudinary.NewFromURL(cloudinaryURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to initialize Cloudinary: %v", err)
// 	}

// 	return &uploadService{
// 		songRepo:        songRepo,
// 		cloudinaryCloud: cld,
// 	}, nil
// }

// func (us *uploadService) UploadMP3ToCloudinary(file io.Reader, filename string, songID string) (string, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
// 	defer cancel()

// 	log.Printf("[Cloudinary] Memulai upload: %s", filename)

// 	uploadResult, err := us.cloudinaryCloud.Upload.Upload(ctx, file, uploader.UploadParams{
// 		// Paksa sebagai video agar muncul di tab Video/Audio Cloudinary
// 		ResourceType: "video",
// 		Folder:       "music_app/songs",
// 		// Hapus karakter aneh dari nama file untuk PublicID
// 		PublicID: fmt.Sprintf("song_%s", songID),
// 	})

// 	if err != nil {
// 		return "", err
// 	}

// 	// Ambil URL yang tersedia (SecureURL lebih baik, tapi URL biasa juga OK)
// 	finalURL := uploadResult.SecureURL
// 	if finalURL == "" {
// 		finalURL = uploadResult.URL
// 	}

// 	// Jika keduanya kosong, kita buat manual menggunakan PublicID sebagai failover
// 	if finalURL == "" && uploadResult.PublicID != "" {
// 		log.Printf("[Cloudinary] URL kosong, mencoba generate manual dari PublicID...")
// 		finalURL = fmt.Sprintf("https://res.cloudinary.com/%s/video/upload/%s",
// 			us.cloudinaryCloud.Config.Cloud.CloudName,
// 			uploadResult.PublicID)
// 	}

// 	if finalURL == "" {
// 		return "", fmt.Errorf("semua field URL dari cloudinary kosong")
// 	}

// 	log.Printf("[Cloudinary] BERHASIL! URL: %s", finalURL)
// 	return finalURL, nil
// }
// func (us *uploadService) UpdateSongWithCustomMP3(songID string, mp3URL string) error {
// 	song, err := us.songRepo.GetSongByID(songID)
// 	if err != nil {
// 		return err
// 	}

// 	now := time.Now()
// 	song.CustomMP3URL = mp3URL
// 	song.HasCustomMP3 = true
// 	song.CustomUploadDate = &now

// 	return us.songRepo.UpdateSong(song)
// }

// func (us *uploadService) GetCustomSongs() ([]models.Song, error) {
// 	return us.songRepo.GetSongsWithCustomMP3()
// }

// func sanitizeFilename(filename string) string {
// 	filename = strings.TrimSuffix(filename, ".mp3")
// 	var result []rune
// 	for _, r := range filename {
// 		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
// 			result = append(result, r)
// 		} else {
// 			result = append(result, '_')
// 		}
// 	}
// 	return string(result)
// }
