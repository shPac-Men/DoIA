package main

import (
	"AwsProj/internal/app/ds"
	"AwsProj/internal/app/dsn"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// .env находится на один уровень выше от cmd/migrate
	envPath := filepath.Join("..", "..", ".env") // от cmd/migrate до корня

	log.Printf("Looking for .env at: %s", envPath)

	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatalf("Error loading .env from %s: %v", envPath, err)
	}

	// Проверяем что переменные загрузились
	if os.Getenv("DB_HOST") == "" {
		log.Fatal("DB_HOST is empty - .env not loaded correctly")
	}

	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Migrate the schema
	err = db.AutoMigrate(
		&ds.Chat{},
		&ds.Message{},
		&ds.MessageChat{},
		&ds.Users{},
	)
	if err != nil {
		log.Fatalf("cant migrate db: %v", err)
	}

	log.Println("Database migrated successfully!")
}
