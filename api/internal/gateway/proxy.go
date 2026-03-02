// Package gateway implements the LLM reverse proxy core that routes API calls
// through AgentTrace for automatic tracing, cost-optimized routing, fallback
// chains, and rate limiting.
package gateway

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RateLimiter implements a token-bucket rate limiter for gateway requests
type RateLimiter struct {
	mu        sync.Mutex
	rpm       int
	tokens    int
	lastReset time.Time
}

// NewRateLimiter creates a rate limiter with the given requests-per-minute limit
func NewRateLimiter(rpm int) *RateLimiter {
	return &RateLimiter{
		rpm:       rpm,
		tokens:    rpm,
		lastReset: time.Now(),
	}
}

// Allow checks if a request is allowed under the rate limit
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	if now.Sub(rl.lastReset) >= time.Minute {
		rl.tokens = rl.rpm
		rl.lastReset = now
	}

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	return false
}

// CircuitBreaker implements a circuit breaker for provider health checking
type CircuitBreaker struct {
	mu            sync.Mutex
	failures      int
	threshold     int
	state         CircuitState
	lastFailure   time.Time
	resetTimeout  time.Duration
	logger        *zap.Logger
}

// CircuitState represents the state of a circuit breaker
type CircuitState string

const (
	CircuitClosed   CircuitState = "closed"
	CircuitOpen     CircuitState = "open"
	CircuitHalfOpen CircuitState = "half_open"
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(threshold int, resetTimeout time.Duration, logger *zap.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		threshold:    threshold,
		state:        CircuitClosed,
		resetTimeout: resetTimeout,
		logger:       logger,
	}
}

// Allow checks if a request should be allowed through the circuit breaker
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
			cb.logger.Info("circuit breaker half-open")
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return true
	}
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == CircuitHalfOpen {
		cb.state = CircuitClosed
		cb.failures = 0
		cb.logger.Info("circuit breaker closed")
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	if cb.failures >= cb.threshold {
		cb.state = CircuitOpen
		cb.logger.Warn("circuit breaker opened",
			zap.Int("failures", cb.failures),
			zap.Int("threshold", cb.threshold),
		)
	}
}

// State returns the current circuit breaker state
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// String returns a human-readable description
func (cb *CircuitBreaker) String() string {
	return fmt.Sprintf("CircuitBreaker(state=%s, failures=%d/%d)", cb.State(), cb.failures, cb.threshold)
}
