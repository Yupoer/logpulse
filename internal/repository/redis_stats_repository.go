package repository

import (
	"context"
	"strconv"

	"github.com/Yupoer/logpulse/internal/domain"
	"github.com/redis/go-redis/v9"
)

type redisStatsRepository struct {
	client *redis.Client
}

// NewStatsRepository injects Redis client dependency.
func NewStatsRepository(client *redis.Client) domain.StatsRepository {
	return &redisStatsRepository{client: client}
}

func (r *redisStatsRepository) IncrementLogCount(ctx context.Context) error {
	return r.client.Incr(ctx, "stats:log_count").Err()
}

func (r *redisStatsRepository) GetLogCount(ctx context.Context) (int64, error) {
	val, err := r.client.Get(ctx, "stats:log_count").Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(val, 10, 64)
}
