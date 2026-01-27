package database

import (
	"back_music/internal/config"
	"back_music/internal/models"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
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
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
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

	log.Println("✅ Database migration completed")
	return nil
}
