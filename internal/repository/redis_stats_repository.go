package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Yupoer/logpulse/internal/domain"
	"github.com/redis/go-redis/v9"
)

type redisCacheRepository struct {
	client *redis.Client
}

// Updated Constructor name
func NewLogCacheRepository(client *redis.Client) domain.LogCacheRepository {
	return &redisCacheRepository{client: client}
}

// --- Original Stats Methods ---

func (r *redisCacheRepository) IncrementLogCount(ctx context.Context) error {
	return r.client.Incr(ctx, "stats:log_count").Err()
}

func (r *redisCacheRepository) GetLogCount(ctx context.Context) (int64, error) {
	val, err := r.client.Get(ctx, "stats:log_count").Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(val, 10, 64)
}

// --- Caching Methods (Cache-Aside) ---

func (r *redisCacheRepository) SetLog(ctx context.Context, entry *domain.LogEntry) error {
	key := fmt.Sprintf("log:%d", entry.ID)

	bytes, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// Set with 1 hour TTL
	// This prevents the cache from growing indefinitely
	return r.client.Set(ctx, key, bytes, 1*time.Hour).Err()
}

func (r *redisCacheRepository) GetLog(ctx context.Context, id uint) (*domain.LogEntry, error) {
	key := fmt.Sprintf("log:%d", id)

	val, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache Miss (Not error)
		}
		return nil, err // System Error
	}

	var entry domain.LogEntry
	if err := json.Unmarshal(val, &entry); err != nil {
		return nil, err
	}
	return &entry, nil // Cache Hit
}
