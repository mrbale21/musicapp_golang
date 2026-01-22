package database

import (
	"log"
)

func RunMigrations() {
    if err := AutoMigrate(); err != nil {
        log.Fatalf("Migration failed: %v", err)
    }
}