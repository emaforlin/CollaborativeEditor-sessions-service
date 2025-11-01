package repository

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisSessionRepo struct {
	redis *redis.Client
}

func (r *RedisSessionRepo) GetActiveSessions(ctx context.Context, docID string) ([]string, error) {
	key := fmt.Sprintf("doc_sessions:%s", docID)

	sessionMembers, err := r.redis.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}
	return sessionMembers, nil
}

func (r *RedisSessionRepo) AddSession(ctx context.Context, docID, userID string) error {
	key := fmt.Sprintf("doc_sessions:%s", docID)
	_, err := r.redis.SAdd(ctx, key, userID).Result()

	if err != nil {
		return fmt.Errorf("failed to add a new session: %w", err)
	}
	return nil
}

func (r *RedisSessionRepo) RemoveSession(ctx context.Context, docID, userID string) error {
	key := fmt.Sprintf("doc_sessions:%s", docID)
	_, err := r.redis.SRem(ctx, key, userID).Result()

	if err != nil {
		return fmt.Errorf("failed to remove a session: %w", err)
	}
	return nil
}

func NewRedisSessionsRepo(redisClient *redis.Client) (*RedisSessionRepo, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("redisClient cannot be nil")
	}

	return &RedisSessionRepo{
		redis: redisClient,
	}, nil
}
