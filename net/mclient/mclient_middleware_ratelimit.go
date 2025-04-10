package mclient

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/graingo/maltose/internal/intlog"
)

// RateLimiter defines the interface for rate limiters.
type RateLimiter interface {
	// Wait blocks until a request can be allowed or context is cancelled.
	Wait(ctx context.Context) error
	// TryAcquire tries to acquire a token without blocking.
	TryAcquire() bool
}

// -----------------------------------------------------------------------------
// Token Bucket Rate Limiter Implementation
// -----------------------------------------------------------------------------

// TokenBucketLimiter implements a token bucket rate limiter.
type TokenBucketLimiter struct {
	rate       float64    // tokens per second
	bucketSize int        // maximum burst size
	tokens     float64    // current number of tokens
	lastTime   time.Time  // last time tokens were added
	mu         sync.Mutex // mutex for thread safety
}

// NewTokenBucketLimiter creates a new token bucket rate limiter.
// The rate is specified in requests per second, and bucketSize
// determines the maximum burst size.
func NewTokenBucketLimiter(rate float64, bucketSize int) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		rate:       rate,
		bucketSize: bucketSize,
		tokens:     float64(bucketSize),
		lastTime:   time.Now(),
	}
}

// refill adds tokens to the bucket based on elapsed time.
func (l *TokenBucketLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(l.lastTime).Seconds()
	l.lastTime = now

	// Calculate new tokens to add based on rate and elapsed time
	newTokens := elapsed * l.rate
	l.tokens += newTokens
	if l.tokens > float64(l.bucketSize) {
		l.tokens = float64(l.bucketSize)
	}
}

// TryAcquire attempts to take a token from the bucket without blocking.
// Returns true if a token was successfully taken, false otherwise.
func (l *TokenBucketLimiter) TryAcquire() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.refill()

	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}

// Wait blocks until a token is available or the context is cancelled.
func (l *TokenBucketLimiter) Wait(ctx context.Context) error {
	for {
		waitTime, allow := l.reserveToken()
		if allow {
			return nil
		}

		select {
		case <-time.After(waitTime):
			// Continue and try again
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// reserveToken calculates the wait time for the next token.
// Returns the wait time and whether a token was immediately available.
func (l *TokenBucketLimiter) reserveToken() (time.Duration, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.refill()

	if l.tokens >= 1 {
		l.tokens--
		return 0, true
	}

	// Calculate time to wait for next token
	waitTime := time.Duration((1 - l.tokens) / l.rate * float64(time.Second))
	return waitTime, false
}

// -----------------------------------------------------------------------------
// Rate Limiting Middleware
// -----------------------------------------------------------------------------

// RateLimitConfig represents options for rate limiting middleware.
type RateLimitConfig struct {
	// RequestsPerSecond is the number of requests allowed per second
	RequestsPerSecond float64
	// Burst is the maximum number of requests allowed to happen at once
	Burst int
	// Skip determines if rate limiting should be skipped for a request
	Skip func(*Request) bool
	// ErrorHandler handles rate limit errors
	ErrorHandler func(context.Context, error) (*Response, error)
}

// MiddlewareRateLimit returns a middleware that limits the rate of requests.
func MiddlewareRateLimit(config RateLimitConfig) MiddlewareFunc {
	// Default values
	rps := config.RequestsPerSecond
	if rps <= 0 {
		rps = 100 // Default: 100 requests/second
	}

	burst := config.Burst
	if burst <= 0 {
		burst = 10 // Default: 10 burst
	}

	// Create a token bucket limiter
	limiter := NewTokenBucketLimiter(rps, burst)

	return func(next HandlerFunc) HandlerFunc {
		return func(req *Request) (*Response, error) {
			ctx := req.Context()

			// Skip rate limiting if condition is met
			if config.Skip != nil && config.Skip(req) {
				return next(req)
			}

			// Attempt to acquire a token
			err := limiter.Wait(ctx)
			if err != nil {
				if config.ErrorHandler != nil {
					return config.ErrorHandler(ctx, err)
				}
				return nil, fmt.Errorf("rate limit error: %w", err)
			}

			// Log rate limiting info in debug mode
			var urlStr string
			if req.Request != nil && req.Request.URL != nil {
				urlStr = req.Request.URL.String()
			} else {
				urlStr = "<no url>"
			}
			intlog.Printf(ctx, "Rate limiter allowed request to %s", urlStr)

			// Proceed with the request
			return next(req)
		}
	}
}

// WithGlobalRateLimit applies rate limiting to all requests from this client.
func (c *Client) WithGlobalRateLimit(rps float64, burst int) *Client {
	c.Use(MiddlewareRateLimit(RateLimitConfig{
		RequestsPerSecond: rps,
		Burst:             burst,
	}))
	return c
}
