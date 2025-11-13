package pkg

import (
	"context"
	"fmt"
	"time"

	"AwsProj/internal/app/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func NewRedisSessionStore(cfg *config.RedisConfig) (sessions.Store, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	logrus.Infof("🔧 Creating Redis store:")
	logrus.Infof("   Address: %s", addr)
	logrus.Infof("   Password: '%s' (length: %d)", cfg.Password, len(cfg.Password))

	// Создайте Redis клиент напрямую
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
	})

	// Проверьте подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logrus.Errorf("❌ Redis ping failed: %v", err)
		return nil, err
	}

	logrus.Info("✅ Redis connection successful")

	// Если нужен session store, используйте cookie store
	// или ищите совместимый redis session store для go-redis
	store := cookie.NewStore([]byte(cfg.SecretKey))

	logrus.Info("✅ Session store created successfully")
	return store, nil
}
