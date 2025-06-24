package main

import (
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/config"
)

func main() {
	// Load .env file for database URL
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading .env:", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	databaseURL := cfg.DatabaseURL

	m, err := migrate.New(
		"file://migrations",
		databaseURL,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	// Get the command line argument (e.g., "up", "down", "force")
	cmd := "up" // Default to "up"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "up":
		log.Println("Running database migrations UP...")
		err = m.Up() // Apply all available migrations
		if err == migrate.ErrNoChange {
			log.Println("No new migrations to apply.")
		} else if err != nil {
			log.Fatalf("Failed to apply migrations: %v", err)
		} else {
			log.Println("Database migrations applied successfully!")
		}
	case "down":
		log.Println("Running database migrations DOWN (one step)...")
		err = m.Down() // Rollback one migration
		if err == migrate.ErrNoChange {
			log.Println("No migrations to rollback.")
		} else if err != nil {
			log.Fatalf("Failed to rollback migration: %v", err)
		} else {
			log.Println("One migration rolled back successfully.")
		}
	case "force":
		if len(os.Args) < 3 {
			log.Fatal("Usage: go run ./cmd/orderservice/migrate force <version>")
		}
		version := os.Args[2]
		log.Printf("Forcing migration version %s...\n", version)
		v, parseErr := time.Parse("20060102150405", version) // Parse timestamp from filename
		if parseErr != nil {
			log.Fatalf("Invalid version format. Use YYYYMMDDHHmmss: %v", parseErr)
		}
		err = m.Force(int(v.Unix())) // Force set version (use with caution!)
		if err != nil {
			log.Fatalf("Failed to force version: %v", err)
		}
		log.Printf("Successfully forced version to %s.\n", version)
	default:
		log.Fatalf("Unknown command: %s. Use 'up' or 'down'.", cmd)
	}
}
