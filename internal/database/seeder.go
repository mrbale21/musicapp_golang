package database

import (
	"back_music/internal/models"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedInitialData creates default users and data for fresh database
func SeedInitialData() error {
	log.Println("🌱 Seeding initial data...")

	// Check if admin already exists
	var adminCount int64
	if err := DB.Model(&models.User{}).Where("role = ?", "admin").Count(&adminCount).Error; err != nil {
		return err
	}

	// If admin already exists, skip seeding
	if adminCount > 0 {
		log.Println("✅ Admin user already exists, skipping seed")
		return nil
	}

	// Hash password untuk admin
	adminPassword := "admin123" // CHANGE THIS IN PRODUCTION!
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Create admin user
	adminUser := models.User{
		Username: "admin",
		Email:    "admin@musicapp.local",
		Password: string(hashedPassword),
		Role:     "admin",
	}

	if err := DB.Create(&adminUser).Error; err != nil {
		log.Printf("⚠️ Failed to create admin user: %v", err)
		return err
	}
	log.Println("✅ Admin user created successfully")
	log.Println("   Email: admin@musicapp.local")
	log.Println("   Password: admin123")
	log.Println("   ⚠️  CHANGE THIS PASSWORD IN PRODUCTION!")

	// Create some demo users for testing
	demoUsers := []models.User{
		{
			Username: "demo_user1",
			Email:    "demo1@musicapp.local",
			Password: hashPassword("demo123"),
			Role:     "user",
		},
		{
			Username: "demo_user2",
			Email:    "demo2@musicapp.local",
			Password: hashPassword("demo123"),
			Role:     "user",
		},
	}

	for _, user := range demoUsers {
		if err := DB.Create(&user).Error; err != nil {
			log.Printf("⚠️ Failed to create demo user %s: %v", user.Username, err)
		}
	}
	log.Println("✅ Demo users created successfully")

	log.Println("🌱 Initial data seeding completed!")
	return nil
}

// Helper function to hash password
func hashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return ""
	}
	return string(hashedPassword)
}
