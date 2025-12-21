package repository

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisSessionRepo struct {
	redis *redis.Client
}

func (r *RedisSessionRepo) GetSessionParticipants(ctx context.Context, docID string) ([]string, error) {
	key := fmt.Sprintf("sessions:doc:%s:users", docID)

	sessionMembers, err := r.redis.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}
	return sessionMembers, nil
}

func (r *RedisSessionRepo) AddSessionParticipant(ctx context.Context, docID, userID string) error {
	userKey := fmt.Sprintf("sessions:doc:%s:users", docID)

	_, err := r.redis.SAdd(ctx, userKey, userID).Result()
	if err != nil {
		return fmt.Errorf("failed to add a new session: %w", err)
	}
	return nil
}

func (r *RedisSessionRepo) RemoveSessionParticipant(ctx context.Context, docID, userID string) (int64, error) {
	userKey := fmt.Sprintf("sessions:doc:%s:users", docID)
	_, err := r.redis.SRem(ctx, userKey, userID).Result()
	count, _ := r.redis.SCard(ctx, userKey).Result()

	return count, err
}

func NewRedisSessionsRepo(redisClient *redis.Client) (*RedisSessionRepo, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("redisClient cannot be nil")
	}

	return &RedisSessionRepo{
		redis: redisClient,
	}, nil
}
