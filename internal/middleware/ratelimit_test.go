package middleware

import (
	"context"
	"testing"

	"github.com/Yupoer/logpulse/internal/config"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_Allow(t *testing.T) {
	// Setup miniredis (in-memory Redis for testing)
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	cfg := config.RateLimitConfig{
		Enabled:  true,
		Capacity: 5,  // Allow 5 burst requests
		Rate:     10, // 10 tokens/sec refill
	}

	rl := NewRateLimiter(client, cfg)
	ctx := context.Background()
	key := "ratelimit:test-ip"

	// Test: First 5 requests should be allowed (burst)
	for i := 0; i < 5; i++ {
		allowed, err := rl.Allow(ctx, key)
		assert.NoError(t, err)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// Test: 6th request should be denied (bucket empty)
	allowed, err := rl.Allow(ctx, key)
	assert.NoError(t, err)
	assert.False(t, allowed, "6th request should be denied")
}

func TestRateLimiter_Disabled(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	cfg := config.RateLimitConfig{
		Enabled:  false, // Disabled
		Capacity: 1,
		Rate:     1,
	}

	rl := NewRateLimiter(client, cfg)
	ctx := context.Background()

	// When disabled, all requests should be allowed
	for i := 0; i < 100; i++ {
		allowed, err := rl.Allow(ctx, "ratelimit:any")
		assert.NoError(t, err)
		assert.True(t, allowed)
	}
}
