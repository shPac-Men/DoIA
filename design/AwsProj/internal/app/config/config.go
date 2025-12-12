package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/golang-jwt/jwt"
)

type Config struct {
	ServiceHost string
	ServicePort int
	TLSCertFile string `mapstructure:"tls_cert_file"`
	TLSKeyFile  string `mapstructure:"tls_key_file"`
	JWT         JWTConfig
}

type JWTConfig struct {
	Token               string        `mapstructure:"token"`
	ExpiresIn           time.Duration `mapstructure:"expires_in"`
	SigningMethodString string        `mapstructure:"signing_method"`
	SigningMethod       jwt.SigningMethod
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

func (c *Config) initJWT() error {
	// Конвертируем строку в метод подписи JWT
	switch c.JWT.SigningMethodString {
	case "HS256":
		c.JWT.SigningMethod = jwt.SigningMethodHS256
	case "HS384":
		c.JWT.SigningMethod = jwt.SigningMethodHS384
	case "HS512":
		c.JWT.SigningMethod = jwt.SigningMethodHS512
	default:
		c.JWT.SigningMethod = jwt.SigningMethodHS256
		c.JWT.SigningMethodString = "HS256"
	}

	return nil
}
