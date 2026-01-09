package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig configures the rate limiter
type RateLimitConfig struct {
	// Max requests per window
	Max int
	// Window duration
	Window time.Duration
	// Key generator function
	KeyGenerator func(*fiber.Ctx) string
	// Skip function
	Skip func(*fiber.Ctx) bool
	// Custom limit exceeded handler
	LimitReached fiber.Handler
}

// DefaultRateLimitConfig returns default rate limit config
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Max:    100,
		Window: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		Skip: nil,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Too Many Requests",
				"message": "Rate limit exceeded. Please try again later.",
			})
		},
	}
}

// RateLimitMiddleware creates a rate limiter using Redis
type RateLimitMiddleware struct {
	redis  *redis.Client
	config RateLimitConfig
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(redisClient *redis.Client, config ...RateLimitConfig) *RateLimitMiddleware {
	cfg := DefaultRateLimitConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return &RateLimitMiddleware{
		redis:  redisClient,
		config: cfg,
	}
}

// Handler returns the rate limit handler
func (m *RateLimitMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip if skip function returns true
		if m.config.Skip != nil && m.config.Skip(c) {
			return c.Next()
		}

		// Generate key
		key := fmt.Sprintf("ratelimit:%s", m.config.KeyGenerator(c))

		// Use sliding window counter algorithm
		now := time.Now().Unix()
		windowStart := now - int64(m.config.Window.Seconds())

		ctx := context.Background()

		// Remove old entries
		m.redis.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(windowStart, 10))

		// Count requests in current window
		count, err := m.redis.ZCard(ctx, key).Result()
		if err != nil {
			// If Redis fails, allow request but log error
			return c.Next()
		}

		// Check if limit exceeded
		if count >= int64(m.config.Max) {
			// Set rate limit headers
			c.Set("X-RateLimit-Limit", strconv.Itoa(m.config.Max))
			c.Set("X-RateLimit-Remaining", "0")
			c.Set("X-RateLimit-Reset", strconv.FormatInt(now+int64(m.config.Window.Seconds()), 10))
			c.Set("Retry-After", strconv.FormatInt(int64(m.config.Window.Seconds()), 10))

			return m.config.LimitReached(c)
		}

		// Add current request
		m.redis.ZAdd(ctx, key, redis.Z{
			Score:  float64(now),
			Member: fmt.Sprintf("%d:%s", now, c.Get("X-Request-ID")),
		})

		// Set expiry on key
		m.redis.Expire(ctx, key, m.config.Window*2)

		// Set rate limit headers
		remaining := m.config.Max - int(count) - 1
		c.Set("X-RateLimit-Limit", strconv.Itoa(m.config.Max))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(now+int64(m.config.Window.Seconds()), 10))

		return c.Next()
	}
}

// ProjectRateLimit creates a rate limiter per project
func (m *RateLimitMiddleware) ProjectRateLimit(maxPerMinute int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		projectID, ok := GetProjectID(c)
		if !ok {
			return c.Next()
		}

		key := fmt.Sprintf("ratelimit:project:%s", projectID.String())
		now := time.Now().Unix()
		windowStart := now - 60 // 1 minute window

		ctx := context.Background()

		// Remove old entries
		m.redis.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(windowStart, 10))

		// Count requests in current window
		count, err := m.redis.ZCard(ctx, key).Result()
		if err != nil {
			return c.Next()
		}

		// Check if limit exceeded
		if count >= int64(maxPerMinute) {
			c.Set("X-RateLimit-Limit", strconv.Itoa(maxPerMinute))
			c.Set("X-RateLimit-Remaining", "0")
			c.Set("Retry-After", "60")

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Too Many Requests",
				"message": "Project rate limit exceeded",
			})
		}

		// Add current request
		m.redis.ZAdd(ctx, key, redis.Z{
			Score:  float64(now),
			Member: fmt.Sprintf("%d:%s", now, c.Get("X-Request-ID")),
		})
		m.redis.Expire(ctx, key, 2*time.Minute)

		remaining := maxPerMinute - int(count) - 1
		c.Set("X-RateLimit-Limit", strconv.Itoa(maxPerMinute))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

		return c.Next()
	}
}

// APIKeyRateLimit creates a rate limiter per API key
func (m *RateLimitMiddleware) APIKeyRateLimit(maxPerMinute int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKeyID, ok := GetAPIKeyID(c)
		if !ok {
			return c.Next()
		}

		key := fmt.Sprintf("ratelimit:apikey:%s", apiKeyID.String())
		now := time.Now().Unix()
		windowStart := now - 60

		ctx := context.Background()

		m.redis.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(windowStart, 10))

		count, err := m.redis.ZCard(ctx, key).Result()
		if err != nil {
			return c.Next()
		}

		if count >= int64(maxPerMinute) {
			c.Set("X-RateLimit-Limit", strconv.Itoa(maxPerMinute))
			c.Set("X-RateLimit-Remaining", "0")
			c.Set("Retry-After", "60")

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Too Many Requests",
				"message": "API key rate limit exceeded",
			})
		}

		m.redis.ZAdd(ctx, key, redis.Z{
			Score:  float64(now),
			Member: fmt.Sprintf("%d:%s", now, c.Get("X-Request-ID")),
		})
		m.redis.Expire(ctx, key, 2*time.Minute)

		remaining := maxPerMinute - int(count) - 1
		c.Set("X-RateLimit-Limit", strconv.Itoa(maxPerMinute))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

		return c.Next()
	}
}

// BurstRateLimit allows bursting with a token bucket algorithm
func (m *RateLimitMiddleware) BurstRateLimit(maxTokens int, refillRate float64) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := fmt.Sprintf("ratelimit:burst:%s", c.IP())
		ctx := context.Background()

		// Get current tokens and last update time
		pipe := m.redis.Pipeline()
		tokensCmd := pipe.Get(ctx, key+":tokens")
		lastCmd := pipe.Get(ctx, key+":last")
		pipe.Exec(ctx)

		tokens := float64(maxTokens)
		if t, err := tokensCmd.Float64(); err == nil {
			tokens = t
		}

		lastUpdate := time.Now()
		if l, err := lastCmd.Int64(); err == nil {
			lastUpdate = time.Unix(l, 0)
		}

		// Calculate token refill
		elapsed := time.Since(lastUpdate).Seconds()
		tokens = min(float64(maxTokens), tokens+elapsed*refillRate)

		// Check if we have tokens
		if tokens < 1 {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Too Many Requests",
				"message": "Rate limit exceeded. Please slow down.",
			})
		}

		// Consume a token
		tokens--

		// Update Redis
		now := time.Now().Unix()
		pipe = m.redis.Pipeline()
		pipe.Set(ctx, key+":tokens", tokens, time.Hour)
		pipe.Set(ctx, key+":last", now, time.Hour)
		pipe.Exec(ctx)

		return c.Next()
	}
}
