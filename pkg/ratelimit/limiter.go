package ratelimit

import (
	"net/http"
	"sync"
	"time"
)

// TokenBucket represents a thread-safe rate limiting bucket for an entity.
type TokenBucket struct {
	mu           sync.Mutex
	maxTokens    float64
	refillRate   float64
	tokens       float64
	lastRefilled time.Time
}

// NewTokenBucket creates a pre-filled token bucket.
func NewTokenBucket(maxTokens float64, refillRate float64) *TokenBucket {
	return &TokenBucket{
		maxTokens:    maxTokens,
		refillRate:   refillRate,
		tokens:       maxTokens,
		lastRefilled: time.Now(),
	}
}

// Allow checks if a request can consume a token. Return true if allowed.
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefilled).Seconds()
	tb.lastRefilled = now

	// Calculate token accumulated over the elapsed time.
	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}

	// Check if we have at least 1 token to allow the incoming request.
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}

	return false
}

// RateLimiter Middleware wraps an http.Handler to intercept excess requests.
type RateLimiter struct {
	bucket *TokenBucket
}

func NewRateLimiter(maxTokens float64, refillRate float64) *RateLimiter {
	return &RateLimiter{
		bucket: NewTokenBucket(maxTokens, refillRate),
	}
}

// Middleware executes the intercepting logic.
func (rl *RateLimiter) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !rl.bucket.Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "Too Many Requests. Rate limit exceeded."}`))
			return
		}
		next(w, r)
	}
}

// // Bucket returns the underlying TokenBucket instance safely
// func (rl *RateLimiter) Bucket() *TokenBucket {
// 	return rl.bucket
// }
