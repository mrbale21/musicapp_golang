package database

import (
	"log"
)

func RunMigrations() {
    if err := AutoMigrate(); err != nil {
        log.Fatalf("Migration failed: %v", err)
    }
}

// RunMigrationsWithSeed runs migrations and then seeds the database
// Useful for production initialization or manual seeding
func RunMigrationsWithSeed() {
	if err := AutoMigrate(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	if err := SeedInitialData(); err != nil {
		log.Printf("Seeding failed: %v", err)
	}
}