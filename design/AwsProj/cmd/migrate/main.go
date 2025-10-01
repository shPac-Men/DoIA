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

	log.Printf("Loaded DB_HOST: '%s'", os.Getenv("DB_HOST"))
	log.Printf("Loaded DB_USER: '%s'", os.Getenv("DB_USER"))
	log.Printf("Loaded DB_NAME: '%s'", os.Getenv("DB_NAME"))
	log.Printf("Loaded DB_PASSWORD: '%s'", os.Getenv("DB_PASSWORD"))
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
		&ds.Elements{},
		&ds.Mixed{},
		&ds.Elements{},
		&ds.Users{},
	)
	if err != nil {
		log.Fatalf("cant migrate db: %v", err)
	}

	log.Println("Database migrated successfully!")
}
