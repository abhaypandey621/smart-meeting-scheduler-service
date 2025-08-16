package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/meeting-scheduler/pkg/repository"
)

func main() {
	// Get database configuration from environment variables
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "meeting_scheduler")

	// Build database connection string
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	// Initialize repository with database connection
	repo, err := repository.NewMySQLRepository(dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Seed test data if environment variable is set
	if os.Getenv("SEED_DATA") == "true" {
		log.Println("Seeding test data...")
		if err := repo.SeedTestData(context.Background()); err != nil {
			log.Fatal("Failed to seed test data:", err)
		}
		log.Println("Test data seeded successfully")
	}

	log.Println("Migration completed successfully")
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
