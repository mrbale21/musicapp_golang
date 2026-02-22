package database

import (
	"back_music/internal/config"
	"back_music/internal/models"
	"fmt"
	"log"
	"os"
	"time"

	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() error {
	cfg := config.GlobalConfig

	// Tambahkan log ini untuk debug di Railway Logs
	log.Printf("Attempting DB Connect: Host=%s, User=%s, DB=%s, Port=%s, SSLMode=%s", 
		cfg.DBHost, cfg.DBUser, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)

	if cfg.DBHost == "" || cfg.DBUser == "" || cfg.DBName == "" {
		// Berikan info variabel mana yang hilang
		return fmt.Errorf("database config incomplete: please check environment variables")
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=10 statement_timeout=30000",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	var err error

	// IMPORTANT:
	// Railway / Supabase biasanya memakai PgBouncer (transaction mode) yang
	// TIDAK cocok dengan prepared statements + statement cache bawaan pgx.
	// Ini yang menyebabkan error:
	//   "prepared statement \"stmtcache_xxx\" already exists (SQLSTATE 42P05)"
	//
	// Solusi: pakai PreferSimpleProtocol=true agar pgx TIDAK memakai prepared
	// statements secara implisit. Ini aman untuk workload kita dan
	// menghilangkan error 500 di endpoint rekomendasi content-based.
	DB, err = gorm.Open(
		gormpostgres.New(gormpostgres.Config{
			DSN:                  dsn,
			PreferSimpleProtocol: true, // ⬅ matikan implicit prepared statements
		}),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	// Pool config optimized for Railway/Supabase
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)

	// 🔥 TEST REAL CONNECTION
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	log.Println("✅ Database connected successfully (Supabase PostgreSQL)")

	// OPTIONAL tapi recommended
	if os.Getenv("ENV") != "production" {
	if err := AutoMigrate(); err != nil {
		return err
	}
}

	return nil
}


func AutoMigrate() error {
	models := []interface{}{
		&models.User{},
		&models.Song{},
		&models.UserLike{},
		&models.UserPlay{},
	}

	for _, model := range models {
		if err := DB.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
	}

	// ⭐ Create performance indexes
	log.Println("📍 Creating performance indexes...")
	
	// Songs indexes for faster filtering
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_songs_genre ON songs(genre)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_songs_popularity ON songs(popularity DESC)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_songs_artist ON songs(artist)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_songs_spotify_id ON songs(spotify_id)")
	
	// UserLike indexes for faster user preference queries
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_user_likes_user_id ON user_likes(user_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_user_likes_song_id ON user_likes(song_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_user_likes_created_at ON user_likes(user_id, created_at DESC)")
	
	// UserPlay indexes for faster play history queries
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_user_plays_user_id ON user_plays(user_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_user_plays_song_id ON user_plays(song_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_user_plays_play_count ON user_plays(user_id, play_count DESC)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_user_plays_last_played ON user_plays(user_id, last_played DESC)")
	
	log.Println("✅ Database migration & indexes completed")
	return nil
}
