package database

import (
	"back_music/internal/config"
	"back_music/internal/models"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() error {
	cfg := config.GlobalConfig

	if cfg.DBHost == "" || cfg.DBUser == "" || cfg.DBName == "" {
		return fmt.Errorf("database config incomplete")
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
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

	// Pool config (Railway/Supabase safe)
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	// 🔥 TEST REAL CONNECTION
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	log.Println("✅ Database connected successfully (Supabase PostgreSQL)")

	// OPTIONAL tapi recommended
	if err := AutoMigrate(); err != nil {
		return err
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
