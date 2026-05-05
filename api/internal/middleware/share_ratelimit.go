package middleware

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type shareRateWindow struct {
	count     int
	expiresAt time.Time
}

// ShareRateLimiter provides a local fallback limiter for unauthenticated share views.
type ShareRateLimiter struct {
	mu         sync.Mutex
	max        int
	maxEntries int
	window     time.Duration
	entries    map[string]shareRateWindow
	clock      func() time.Time
}

const defaultShareRateLimitEntries = 4096

// NewShareRateLimiter creates a bounded fixed-window share limiter.
func NewShareRateLimiter(maxRequests int, window time.Duration) *ShareRateLimiter {
	return &ShareRateLimiter{
		max:        maxRequests,
		maxEntries: defaultShareRateLimitEntries,
		window:     window,
		entries:    make(map[string]shareRateWindow),
		clock:      time.Now,
	}
}

// Handler limits by IP and periodically removes expired entries.
func (l *ShareRateLimiter) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		now := l.clock()
		key := strings.Clone(c.IP())

		l.mu.Lock()
		if _, exists := l.entries[key]; !exists && len(l.entries) >= l.maxEntries {
			l.removeExpired(now)
			if len(l.entries) >= l.maxEntries {
				l.mu.Unlock()
				c.Set("Retry-After", strconv.Itoa(max(1, int(l.window.Seconds()))))
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error":   "Too Many Requests",
					"message": "Share link rate limiter is at capacity",
				})
			}
		}
		entry := l.entries[key]
		if !now.Before(entry.expiresAt) {
			entry = shareRateWindow{expiresAt: now.Add(l.window)}
		}
		if entry.count >= l.max {
			retryAfter := int(entry.expiresAt.Sub(now).Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			l.mu.Unlock()
			c.Set("Retry-After", strconv.Itoa(retryAfter))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Too Many Requests",
				"message": "Share link rate limit exceeded",
			})
		}
		entry.count++
		l.entries[key] = entry
		remaining := l.max - entry.count
		l.mu.Unlock()

		c.Set("X-RateLimit-Limit", strconv.Itoa(l.max))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		return c.Next()
	}
}

func (l *ShareRateLimiter) removeExpired(now time.Time) {
	for candidate, entry := range l.entries {
		if !now.Before(entry.expiresAt) {
			delete(l.entries, candidate)
		}
	}
}
