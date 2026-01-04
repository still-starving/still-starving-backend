package services

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	client *redis.Client
}

func NewRedisService(client *redis.Client) *RedisService {
	return &RedisService{client: client}
}

// StoreRefreshToken stores refresh token with TTL
func (s *RedisService) StoreRefreshToken(ctx context.Context, userID, token string, expiration time.Duration) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	return s.client.Set(ctx, key, userID, expiration).Err()
}

// GetUserIDByToken retrieves user ID from refresh token
func (s *RedisService) GetUserIDByToken(ctx context.Context, token string) (string, error) {
	key := fmt.Sprintf("refresh_token:%s", token)
	return s.client.Get(ctx, key).Result()
}

// RevokeRefreshToken deletes refresh token
func (s *RedisService) RevokeRefreshToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	return s.client.Del(ctx, key).Err()
}

// RevokeAllUserTokens revokes all tokens for a user (for logout all sessions)
func (s *RedisService) RevokeAllUserTokens(ctx context.Context, userID string) error {
	pattern := "refresh_token:*"
	iter := s.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()
		val, err := s.client.Get(ctx, key).Result()
		if err == nil && val == userID {
			s.client.Del(ctx, key)
		}
	}

	return iter.Err()
}
