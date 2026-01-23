package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/Yupoer/logpulse/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Token Bucket Lua Script
// This script atomically checks and consumes tokens from the bucket.
// Atomicity is guaranteed by Redis executing Lua scripts in a single thread.
const tokenBucketScript = `
-- KEYS[1] = rate limit key (e.g., "ratelimit:192.168.1.1")
-- ARGV[1] = capacity (bucket capacity)
-- ARGV[2] = rate (tokens per second refill rate)
-- ARGV[3] = now (current timestamp in milliseconds)
-- ARGV[4] = requested (number of tokens to consume, usually 1)

local tokens_key = KEYS[1]
local capacity = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])

-- Get current bucket state
local data = redis.call("HMGET", tokens_key, "tokens", "last_time")
local tokens = tonumber(data[1]) or capacity  -- Default to full bucket on first access
local last_time = tonumber(data[2]) or now

-- Calculate tokens to add based on elapsed time
local elapsed = (now - last_time) / 1000  -- Convert to seconds
local new_tokens = math.min(capacity, tokens + elapsed * rate)

-- Try to consume tokens
if new_tokens >= requested then
    new_tokens = new_tokens - requested
    redis.call("HMSET", tokens_key, "tokens", new_tokens, "last_time", now)
    redis.call("EXPIRE", tokens_key, 60)  -- TTL to prevent memory leak
    return 1  -- Allow
else
    redis.call("HMSET", tokens_key, "tokens", new_tokens, "last_time", now)
    redis.call("EXPIRE", tokens_key, 60)
    return 0  -- Deny
end
`

// RateLimiter implements Token Bucket rate limiting using Redis
type RateLimiter struct {
	client   *redis.Client
	script   *redis.Script
	capacity int64
	rate     float64
	enabled  bool
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(client *redis.Client, cfg config.RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		client:   client,
		script:   redis.NewScript(tokenBucketScript),
		capacity: cfg.Capacity,
		rate:     cfg.Rate,
		enabled:  cfg.Enabled,
	}
}

// Allow checks if a request is allowed under the rate limit
func (rl *RateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	if !rl.enabled {
		return true, nil
	}

	now := time.Now().UnixMilli()
	result, err := rl.script.Run(
		ctx,
		rl.client,
		[]string{key},
		rl.capacity,
		rl.rate,
		now,
		1, // Request 1 token
	).Int()

	if err != nil {
		return false, err
	}
	return result == 1, nil
}

// Middleware returns a Gin middleware that applies rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if rate limiting is disabled
		if !rl.enabled {
			c.Next()
			return
		}

		// Use client IP as the rate limit key
		key := "ratelimit:" + c.ClientIP()

		allowed, err := rl.Allow(c.Request.Context(), key)
		if err != nil {
			// On Redis error, fail open (allow request) but log the error
			c.Next()
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Try again later.",
			})
			return
		}

		c.Next()
	}
}
