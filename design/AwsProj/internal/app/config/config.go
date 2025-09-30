package config

import (
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceHost string
	ServicePort int
}

func NewConfig() (*Config, error) {
	var err error

	configName := "config"
	_ = godotenv.Load()
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")

	// Добавляем правильные пути поиска конфига
	viper.AddConfigPath("config")       // текущая директория/config
	viper.AddConfigPath("../config")    // на уровень выше/config
	viper.AddConfigPath("../../config") // на два уровня выше/config (из cmd/awsProject)
	viper.AddConfigPath(".")            // текущая директория
	viper.AddConfigPath("../")          // на уровень выше
	viper.AddConfigPath("../../")       // на два уровня выше (корень проекта)

	viper.WatchConfig()

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	log.Info("config parsed")

	return cfg, nil
}
