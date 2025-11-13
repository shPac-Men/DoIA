package config

import (
	"os"
)

type RedisConfig struct {
	Host      string
	Port      string
	Password  string
	SecretKey string
}

func NewRedisConfig() *RedisConfig {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	password := os.Getenv("REDIS_PASSWORD")
	// password может быть пустым или содержать значение

	secretKey := os.Getenv("REDIS_SECRET_KEY")
	if secretKey == "" {
		secretKey = "secret-key"
	}

	return &RedisConfig{
		Host:      host,
		Port:      port,
		Password:  password,
		SecretKey: secretKey,
	}
}
