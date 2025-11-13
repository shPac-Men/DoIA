package pkg

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type RedisTokenService struct {
	client *redis.Client
}

func NewRedisTokenService(client *redis.Client) *RedisTokenService {
	return &RedisTokenService{client: client}
}

// SaveToken сохраняет JWT в Redis с TTL
func (s *RedisTokenService) SaveToken(ctx context.Context, userID uint, token string, ttl time.Duration) error {
	key := fmt.Sprintf("token:%d", userID)
	logrus.Debugf("💾 Saving token to Redis: key=%s, ttl=%v", key, ttl)
	return s.client.Set(ctx, key, token, ttl).Err()
}

// GetToken получает JWT из Redis
func (s *RedisTokenService) GetToken(ctx context.Context, userID uint) (string, error) {
	key := fmt.Sprintf("token:%d", userID)
	token, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		logrus.Debugf("🔑 Token not found in Redis: key=%s", key)
		return "", fmt.Errorf("token not found or expired")
	}
	if err != nil {
		logrus.Errorf("❌ Redis error: %v", err)
		return "", err
	}
	logrus.Debugf("✅ Token found in Redis: key=%s", key)
	return token, nil
}

// RevokeToken удаляет JWT из Redis (логаут)
func (s *RedisTokenService) RevokeToken(ctx context.Context, userID uint) error {
	key := fmt.Sprintf("token:%d", userID)
	logrus.Debugf("🗑️ Revoking token: key=%s", key)
	return s.client.Del(ctx, key).Err()
}

// ValidateTokenExists проверяет, существует ли токен в Redis
func (s *RedisTokenService) ValidateTokenExists(ctx context.Context, userID uint) bool {
	key := fmt.Sprintf("token:%d", userID)
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		logrus.Errorf("❌ Error checking token existence: %v", err)
		return false
	}
	return exists > 0
}

// GetTokenTTL получает оставшееся время жизни токена
func (s *RedisTokenService) GetTokenTTL(ctx context.Context, userID uint) (time.Duration, error) {
	key := fmt.Sprintf("token:%d", userID)
	ttl, err := s.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return ttl, nil
}

// RevokeAllUserTokens удаляет все токены пользователя
func (s *RedisTokenService) RevokeAllUserTokens(ctx context.Context, userID uint) error {
	pattern := fmt.Sprintf("token:%d*", userID)
	iter := s.client.Scan(ctx, 0, pattern, 0).Iterator()

	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		logrus.Errorf("❌ Error scanning keys: %v", err)
		return err
	}

	if len(keys) > 0 {
		return s.client.Del(ctx, keys...).Err()
	}

	return nil
}
