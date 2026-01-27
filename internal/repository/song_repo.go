package repository

import (
	"errors"
	"log"

	"back_music/internal/database"
	"back_music/internal/models"
	"gorm.io/gorm"
)

var ErrSongNotFound = errors.New("song not found")

// repository/songRepo.go - PERBAIKAN:

type SongRepository interface {
    CreateSong(song *models.Song) error
    GetSongByID(id string) (*models.Song, error)
    GetSongBySpotifyID(spotifyID string) (*models.Song, error)
    GetAllSongs() ([]models.Song, error)
    GetSongsByIDs(ids []string) ([]models.Song, error)
    GetRandomSongs(limit int) ([]models.Song, error)
    SearchSongs(query string, limit int) ([]models.Song, error)
    GetSongsByGenre(genre string, limit int) ([]models.Song, error)
    GetPopularSongs(limit int) ([]models.Song, error)
    UpdateSong(song *models.Song) error
     IsSongLikedByUser(songID string, userID uint) (bool, error)
 GetAllSongsWithLikeStatus(userID uint) ([]models.Song, error)
  SearchSongsWithLikeStatus(query string, limit int, userID uint) ([]models.Song, error)
}

type songRepo struct {
    db *gorm.DB
}

func NewSongRepository() SongRepository {
    return &songRepo{db: database.DB}
}

// ================ PERBAIKAN UTAMA ================

func (r *songRepo) GetAllSongs() ([]models.Song, error) {
    var songs []models.Song
    err := r.db.Unscoped().Order("created_at DESC").Find(&songs).Error
    if err != nil {
        return nil, err
    }
    if songs == nil {
        songs = []models.Song{}
    }
    log.Printf("[Repository GetAllSongs] Total songs fetched: %d", len(songs))
    return songs, nil
}

func (r *songRepo) GetAllSongsWithLikeStatus(userID uint) ([]models.Song, error) {
    var songs []models.Song
    err := r.db.Unscoped().Order("created_at DESC").Find(&songs).Error
    if err != nil {
        return nil, err
    }
    if songs == nil {
        songs = []models.Song{}
    }
    log.Printf("[GetAllSongsWithLikeStatus] Fetched %d songs", len(songs))

    var likedSongIDs []string
    _ = r.db.Model(&models.UserLike{}).
        Where("user_id = ?", userID).
        Pluck("song_id", &likedSongIDs).Error

    likedMap := make(map[string]bool)
    for _, id := range likedSongIDs {
        likedMap[id] = true
    }
    for i := range songs {
        songs[i].IsLiked = likedMap[songs[i].ID]
    }
    return songs, nil
}

// ================ METHOD LAINNYA DENGAN LOGGING ================

func (r *songRepo) CreateSong(song *models.Song) error {
    return r.db.Create(song).Error
}

func (r *songRepo) GetSongByID(id string) (*models.Song, error) {
    var song models.Song
    err := r.db.First(&song, "id = ?", id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrSongNotFound
        }
        return nil, err
    }
    log.Printf("[GetSongByID] Found song: %s - %s", song.Title, song.Artist)
    return &song, nil
}

func (r *songRepo) GetSongBySpotifyID(spotifyID string) (*models.Song, error) {
    var song models.Song
    err := r.db.First(&song, "spotify_id = ?", spotifyID).Error
    return &song, err
}

func (r *songRepo) GetSongsByIDs(ids []string) ([]models.Song, error) {
    var songs []models.Song
    err := r.db.Where("id IN ?", ids).Find(&songs).Error
    return songs, err
}

func (r *songRepo) GetRandomSongs(limit int) ([]models.Song, error) {
    var songs []models.Song
    err := r.db.Order("RANDOM()").Limit(limit).Find(&songs).Error
    return songs, err
}

func (r *songRepo) SearchSongs(query string, limit int) ([]models.Song, error) {
    var songs []models.Song
    err := r.db.Where("title ILIKE ? OR artist ILIKE ? OR genre ILIKE ?", 
        "%"+query+"%", "%"+query+"%", "%"+query+"%").
        Limit(limit).
        Find(&songs).Error
    return songs, err
}

func (r *songRepo) GetSongsByGenre(genre string, limit int) ([]models.Song, error) {
    var songs []models.Song
    err := r.db.Where("genre ILIKE ?", "%"+genre+"%").
        Limit(limit).
        Find(&songs).Error
    return songs, err
}

func (r *songRepo) GetPopularSongs(limit int) ([]models.Song, error) {
    var songs []models.Song
    err := r.db.Order("popularity DESC").Limit(limit).Find(&songs).Error
    if err != nil {
        return nil, err
    }
    if songs == nil {
        songs = []models.Song{}
    }
    return songs, nil
}



func (r *songRepo) UpdateSong(song *models.Song) error {
    log.Printf("[UpdateSong] Updating song: %s - %s", song.Title, song.Artist)
    return r.db.Save(song).Error
}

func (r *songRepo) IsSongLikedByUser(songID string, userID uint) (bool, error) {
    var count int64
    err := r.db.Model(&models.UserLike{}).
        Where("song_id = ? AND user_id = ?", songID, userID).
        Count(&count).Error
    
    return count > 0, err
}

func (r *songRepo) SearchSongsWithLikeStatus(query string, limit int, userID uint) ([]models.Song, error) {
    var songs []models.Song
    
    // Search songs
    if err := r.db.Where("title ILIKE ? OR artist ILIKE ?", 
        "%"+query+"%", "%"+query+"%").Limit(limit).Find(&songs).Error; err != nil {
        return nil, err
    }
    
    if len(songs) == 0 {
        return songs, nil
    }
    
    // Get song IDs
    songIDs := make([]string, len(songs))
    for i, song := range songs {
        songIDs[i] = song.ID
    }
    
    // Get which of these songs are liked by user
    var likedSongIDs []string
    err := r.db.Model(&models.UserLike{}).
        Where("user_id = ? AND song_id IN ?", userID, songIDs).
        Pluck("song_id", &likedSongIDs).Error
    
    if err != nil {
        return songs, nil
    }
    
    // Create map for faster lookup
    likedMap := make(map[string]bool)
    for _, id := range likedSongIDs {
        likedMap[id] = true
    }
    
    // Set IsLiked field
    for i := range songs {
        songs[i].IsLiked = likedMap[songs[i].ID]
    }
    
    return songs, nil
}
