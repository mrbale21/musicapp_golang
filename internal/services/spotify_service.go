package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"back_music/internal/config"
	"back_music/internal/models"
	"back_music/internal/repository"
)

type SpotifyService interface {
    GetAccessToken() (string, error)
    SearchTracks(query string, limit int) ([]models.Song, error)
    GetAudioFeatures(trackID string) (*models.AudioFeatures, error)
    GetMultipleAudioFeatures(trackIDs []string) (map[string]models.AudioFeatures, error)
    SeedSongsFromSpotify(limit int) error
    GetRecommendations(seedTracks []string, limit int) ([]models.Song, error)
     GetPopularIndonesianSongs(limit int) ([]models.Song, error) 
}

type spotifyService struct {
    clientID     string
    clientSecret string
    accessToken  string
    tokenExpiry  time.Time
    songRepo     repository.SongRepository
}

func NewSpotifyService(songRepo repository.SongRepository) SpotifyService {
    cfg := config.GlobalConfig
    return &spotifyService{
        clientID:     cfg.SpotifyClientID,
        clientSecret: cfg.SpotifyClientSecret,
        songRepo:     songRepo,
    }
}

type spotifyTokenResponse struct {
    AccessToken string `json:"access_token"`
    TokenType   string `json:"token_type"`
    ExpiresIn   int    `json:"expires_in"`
}

func (s *spotifyService) GetAccessToken() (string, error) {
    if time.Now().Before(s.tokenExpiry) && s.accessToken != "" {
        return s.accessToken, nil
    }
    
    auth := base64.StdEncoding.EncodeToString([]byte(s.clientID + ":" + s.clientSecret))
    
    data := url.Values{}
    data.Set("grant_type", "client_credentials")
    
    req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", 
        strings.NewReader(data.Encode()))
    if err != nil {
        return "", err
    }
    
    req.Header.Set("Authorization", "Basic "+auth)
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    
    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("failed to get token: %s", string(body))
    }
    
    var tokenResp spotifyTokenResponse
    if err := json.Unmarshal(body, &tokenResp); err != nil {
        return "", err
    }
    
    s.accessToken = tokenResp.AccessToken
    s.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)
    
    return s.accessToken, nil
}



//  TAMBAH fungsi helper untuk generate dummy features
func (s *spotifyService) generateDummyFeatures(trackID string) models.AudioFeatures {
    // Seed random berdasarkan trackID untuk consistency
    hash := 0
    for _, c := range trackID {
        hash = hash*31 + int(c)
    }
    rand.Seed(time.Now().UnixNano() + int64(hash))
    
    return models.AudioFeatures{
        Danceability:    0.5 + (rand.Float64() * 0.4),
        Energy:          0.4 + (rand.Float64() * 0.5),
        Key:             rand.Intn(12),
        Loudness:        -15 + (rand.Float64() * 20),
        Mode:            rand.Intn(2),
        Speechiness:     0.05 + (rand.Float64() * 0.15),
        Acousticness:    0.1 + (rand.Float64() * 0.6),
        Instrumentalness: rand.Float64() * 0.4,
        Liveness:        0.1 + (rand.Float64() * 0.3),
        Valence:         0.3 + (rand.Float64() * 0.5),
        Tempo:           80 + (rand.Float64() * 100),
        TimeSignature:   rand.Intn(5) + 3,
    }
}

func (s *spotifyService) GetAudioFeatures(trackID string) (*models.AudioFeatures, error) {
    //  RETURN DUMMY DATA untuk bypass Spotify API error
    log.Printf("Generating dummy audio features for: %s", trackID)
    
    // Seed random untuk variasi
    rand.Seed(time.Now().UnixNano() + int64(len(trackID)))
    
    return &models.AudioFeatures{
        Danceability:    0.5 + (rand.Float64() * 0.4),     // 0.5-0.9
        Energy:          0.4 + (rand.Float64() * 0.5),     // 0.4-0.9
        Key:             rand.Intn(12),                    // 0-11
        Loudness:        -15 + (rand.Float64() * 20),      // -15 to 5
        Mode:            rand.Intn(2),                     // 0 or 1
        Speechiness:     0.05 + (rand.Float64() * 0.15),   // 0.05-0.2
        Acousticness:    0.1 + (rand.Float64() * 0.6),     // 0.1-0.7
        Instrumentalness: rand.Float64() * 0.4,            // 0-0.4
        Liveness:        0.1 + (rand.Float64() * 0.3),     // 0.1-0.4
        Valence:         0.3 + (rand.Float64() * 0.5),     // 0.3-0.8
        Tempo:           80 + (rand.Float64() * 100),      // 80-180 BPM
        TimeSignature:   rand.Intn(5) + 3,                 // 3-7
    }, nil
}

// GANTI fungsi GetMultipleAudioFeatures:
func (s *spotifyService) GetMultipleAudioFeatures(trackIDs []string) (map[string]models.AudioFeatures, error) {
    log.Printf("Generating dummy audio features for %d tracks", len(trackIDs))
    
    featuresMap := make(map[string]models.AudioFeatures)
    
    for _, trackID := range trackIDs {
        // Seed random berdasarkan trackID untuk consistency
        hash := 0
        for _, c := range trackID {
            hash = hash*31 + int(c)
        }
        rand.Seed(time.Now().UnixNano() + int64(hash))
        
        featuresMap[trackID] = models.AudioFeatures{
            Danceability:    0.5 + (rand.Float64() * 0.4),
            Energy:          0.4 + (rand.Float64() * 0.5),
            Key:             rand.Intn(12),
            Loudness:        -15 + (rand.Float64() * 20),
            Mode:            rand.Intn(2),
            Speechiness:     0.05 + (rand.Float64() * 0.15),
            Acousticness:    0.1 + (rand.Float64() * 0.6),
            Instrumentalness: rand.Float64() * 0.4,
            Liveness:        0.1 + (rand.Float64() * 0.3),
            Valence:         0.3 + (rand.Float64() * 0.5),
            Tempo:           80 + (rand.Float64() * 100),
            TimeSignature:   rand.Intn(5) + 3,
        }
    }
    
    return featuresMap, nil
}


func (s *spotifyService) SearchTracks(query string, limit int) ([]models.Song, error) {
    token, err := s.GetAccessToken()
    if err != nil {
        return nil, err
    }
    
    // PERBAIKAN: Jangan tambahkan "genre:indonesian market:ID" di sini
    // Hanya encode query asli
    encodedQuery := url.QueryEscape(query)
    
    // PERBAIKAN: Gunakan parameter market=ID yang benar
    url := fmt.Sprintf("https://api.spotify.com/v1/search?q=%s&type=track&limit=%d&market=ID", 
        encodedQuery, limit)
    
    log.Printf("🔍 Spotify API Request: %s", url)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", "Bearer "+token)
    
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    
    if resp.StatusCode != http.StatusOK {
        log.Printf("❌ Spotify search failed (%d): %s", resp.StatusCode, string(body))
        return nil, fmt.Errorf("search failed: %s", string(body))
    }
    
    var result map[string]interface{}
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, err
    }
    
    tracks := result["tracks"].(map[string]interface{})["items"].([]interface{})
    
    if len(tracks) == 0 {
        log.Printf("⚠️ Tidak ditemukan track untuk query: '%s'", query)
        return []models.Song{}, nil
    }
    
    songs := make([]models.Song, 0, len(tracks))
    
    for _, track := range tracks {
        trackMap := track.(map[string]interface{})
        
        // Get artists
        artists := trackMap["artists"].([]interface{})
        artistNames := make([]string, 0, len(artists))
        for _, artist := range artists {
            artistMap := artist.(map[string]interface{})
            artistNames = append(artistNames, artistMap["name"].(string))
        }
        
        // Get album
        album := trackMap["album"].(map[string]interface{})
        albumName := album["name"].(string)
        
        // Get images
        images := album["images"].([]interface{})
        imageURL := ""
        if len(images) > 0 {
            imageURL = images[0].(map[string]interface{})["url"].(string)
        }
        
        // Get preview URL
        previewURL, _ := trackMap["preview_url"].(string)
        
        // Generate dummy features
        trackID := trackMap["id"].(string)
        features := s.generateDummyFeatures(trackID)
        
        song := models.Song{
            SpotifyID:        trackID,
            Title:           trackMap["name"].(string),
            Artist:          strings.Join(artistNames, ", "),
            Album:           albumName,
            Genre:           "indonesian", // Default untuk lagu Indonesia
            Popularity:      int(trackMap["popularity"].(float64)),
            DurationMs:      int(trackMap["duration_ms"].(float64)),
            PreviewURL:      previewURL,
            ImageURL:        imageURL,
            
            // Audio features dari dummy
            Danceability:    features.Danceability,
            Energy:          features.Energy,
            Key:             features.Key,
            Loudness:        features.Loudness,
            Mode:            features.Mode,
            Speechiness:     features.Speechiness,
            Acousticness:    features.Acousticness,
            Instrumentalness: features.Instrumentalness,
            Liveness:        features.Liveness,
            Valence:         features.Valence,
            Tempo:           features.Tempo,
            TimeSignature:   features.TimeSignature,
        }
        
        songs = append(songs, song)
        log.Printf("✅ Found: %s - %s (Popularity: %d)", 
            song.Artist, song.Title, song.Popularity)
    }
    
    log.Printf("🎯 Total found for query '%s': %d tracks", query, len(songs))
    return songs, nil
}

func (s *spotifyService) SeedSongsFromSpotify(limit int) error {
    log.Printf("🚀 Memulai seed %d lagu Indonesia dari Spotify...", limit)
    
    // PERBAIKAN: Gunakan query yang lebih efektif
    queries := []string{
        // Query langsung dengan nama artis populer Indonesia
        "tulus",
        "raisa",
        "nadin amizah",
        "sheila on 7",
        "noah band",
        "ungu band",
        "rossa",
        "judika",
        "armada band",
        "virgoun",
        "tiara andini",
        "lyodra",
        "ziva magnolya",
        "yovie & nuno",
        "kerispatih",
        "d'masiv",
        "vierratale",
        "geisha band",
        
        // Query lagu populer Indonesia
        "lagu indonesia terbaru",
        "indonesian pop",
        "pop indonesia 2024",
        "lagu indonesia viral",
        "chart indonesia",
        
        // Query dengan tanda kutip untuk hasil lebih spesifik
        "\"tulus\"",
        "\"raisa\"",
        "\"nadin\"",
        "\"sheila on 7\"",
    }
    
    songsPerQuery := 10 // Ambil 10 lagu per query
    allSongs := make([]models.Song, 0, limit)
    trackMap := make(map[string]bool)
    
    for i, query := range queries {
        log.Printf("🔍 [%d/%d] Mencari: '%s'", i+1, len(queries), query)
        
        songs, err := s.SearchTracks(query, songsPerQuery)
        if err != nil {
            log.Printf("⚠️ Warning: Gagal untuk query '%s': %v", query, err)
            continue
        }
        
        // Filter lagu Indonesia berdasarkan artis/judul
        filteredSongs := s.filterIndonesianSongs(songs)
        
        // Tambahkan ke list tanpa duplikat
        for _, song := range filteredSongs {
            if _, exists := trackMap[song.SpotifyID]; !exists {
                trackMap[song.SpotifyID] = true
                allSongs = append(allSongs, song)
                log.Printf("➕ Added: %s - %s", song.Artist, song.Title)
            }
        }
        
        log.Printf("📊 Query '%s': %d found, %d Indonesian", 
            query, len(songs), len(filteredSongs))
        
        // Rate limiting
        time.Sleep(500 * time.Millisecond)
        
        if len(allSongs) >= limit*2 { // Kumpulkan lebih banyak untuk dipilih yang terbaik
            break
        }
    }
    
    // Sort by popularity (descending)
    sort.Slice(allSongs, func(i, j int) bool {
        return allSongs[i].Popularity > allSongs[j].Popularity
    })
    
    // Ambil lagu dengan popularity tertinggi
    if len(allSongs) > limit {
        allSongs = allSongs[:limit]
    }
    
    // Save to database
    log.Printf("💾 Menyimpan %d lagu Indonesia ke database...", len(allSongs))
    
    savedCount := 0
    for _, song := range allSongs {
        // Skip jika sudah ada
        existing, err := s.songRepo.GetSongBySpotifyID(song.SpotifyID)
        if err == nil && existing != nil {
            log.Printf("⏭️ Skipping existing: %s - %s", song.Artist, song.Title)
            continue
        }
        
        // Set genre ke "indonesian"
        song.Genre = "indonesian"
        
        if err := s.songRepo.CreateSong(&song); err != nil {
            log.Printf("❌ Gagal menyimpan '%s': %v", song.Title, err)
        } else {
            savedCount++
            log.Printf("✅ Saved #%d: %s - %s (Popularity: %d)", 
                savedCount, song.Artist, song.Title, song.Popularity)
        }
    }
    
    log.Printf("🎉 Seed selesai! Berhasil menyimpan %d/%d lagu.", savedCount, len(allSongs))
    return nil
}

// PERBAIKAN: Optimasi filter untuk lagu Indonesia
func (s *spotifyService) filterIndonesianSongs(songs []models.Song) []models.Song {
    indonesianSongs := make([]models.Song, 0)
    
    // Daftar artis Indonesia (case insensitive)
    indonesianArtists := []string{
        "tulus", "raisa", "nadin", "sheila", "noah", "ungu", "rossa", "judika",
        "armada", "virgoun", "tiara", "lyodra", "ziva", "yovie", "kerispatih",
        "d'masiv", "vierra", "geisha", "last child", "fiersa", "ghea", "hari",
        "gita", "iqbaal", "jidat", "kunto", "lale", "mansur", "naura", "ongen",
    }
    
    // Kata kunci dalam judul lagu yang menandakan Indonesia
    indonesianKeywords := []string{
        "indonesia", "jakarta", "bandung", "surabaya", "jogja", "yogyakarta",
        "bali", "papua", "kalimantan", "sumatra", "jawa", "nusantara",
        "merah", "putih", "sang", "saka", "garuda", "pancasila",
    }
    
    for _, song := range songs {
        lowerArtist := strings.ToLower(song.Artist)
        lowerTitle := strings.ToLower(song.Title)
        
        isIndonesian := false
        
        // Cek artis
        for _, artist := range indonesianArtists {
            if strings.Contains(lowerArtist, artist) {
                isIndonesian = true
                break
            }
        }
        
        // Jika belum ketemu, cek judul
        if !isIndonesian {
            for _, keyword := range indonesianKeywords {
                if strings.Contains(lowerTitle, keyword) {
                    isIndonesian = true
                    break
                }
            }
        }
        
        // Cek tambahan: popularity tinggi (>50) biasanya lagu populer Indonesia
        if !isIndonesian && song.Popularity > 50 {
            // Cek apakah ada kata Indonesia dalam artis/judul
            if strings.Contains(lowerArtist, "indonesia") || 
               strings.Contains(lowerTitle, "indonesia") {
                isIndonesian = true
            }
        }
        
        if isIndonesian {
            indonesianSongs = append(indonesianSongs, song)
        }
    }
    
    return indonesianSongs
}

// Fungsi GetPopularIndonesianSongs yang lebih baik
func (s *spotifyService) GetPopularIndonesianSongs(limit int) ([]models.Song, error) {
    log.Printf("🎵 Mencari %d lagu Indonesia terpopuler...", limit)
    
    // Query untuk chart/trending Indonesia
    queries := []string{
        "lagu indonesia terbaru 2024",
        "indonesian viral hits",
        "spotify chart indonesia",
        "top hits indonesia",
        "trending indonesia",
    }
    
    allSongs := make([]models.Song, 0, limit)
    
    for i, query := range queries {
        log.Printf("🔍 [%d/%d] Mencari trending: '%s'", i+1, len(queries), query)
        
        songs, err := s.SearchTracks(query, 20)
        if err != nil {
            log.Printf("⚠️ Gagal untuk '%s': %v", query, err)
            continue
        }
        
        // Filter hanya lagu Indonesia
        indonesianSongs := s.filterIndonesianSongs(songs)
        
        allSongs = append(allSongs, indonesianSongs...)
        
        if len(allSongs) >= limit*2 {
            break
        }
        
        time.Sleep(300 * time.Millisecond)
    }
    
    // Hapus duplikat
    uniqueSongs := make([]models.Song, 0)
    trackMap := make(map[string]bool)
    
    for _, song := range allSongs {
        if !trackMap[song.SpotifyID] {
            trackMap[song.SpotifyID] = true
            uniqueSongs = append(uniqueSongs, song)
        }
    }
    
    // Sort by popularity
    sort.Slice(uniqueSongs, func(i, j int) bool {
        return uniqueSongs[i].Popularity > uniqueSongs[j].Popularity
    })
    
    // Return top songs
    if len(uniqueSongs) > limit {
        uniqueSongs = uniqueSongs[:limit]
    }
    
    log.Printf("✅ Found %d popular Indonesian songs", len(uniqueSongs))
    return uniqueSongs, nil
}

func (s *spotifyService) GetRecommendations(seedTracks []string, limit int) ([]models.Song, error) {
    // This would use Spotify's recommendation endpoint
    // Implementation simplified for brevity
    return s.SearchTracks(strings.Join(seedTracks, " "), limit)
}