package mclient

import (
	"context"
	"fmt"
	"net/http"
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

// RateLimitOption represents options for rate limiting middleware.
type RateLimitOption struct {
	// Limiter is the rate limiter implementation to use.
	Limiter RateLimiter
	// SkipCondition determines if rate limiting should be skipped for a request.
	SkipCondition func(*http.Request) bool
	// ErrorHandler handles rate limit errors.
	ErrorHandler func(context.Context, error) (*http.Response, error)
}

// WithRateLimit returns a middleware that limits the rate of requests.
func WithRateLimit(options RateLimitOption) MiddlewareFunc {
	// Create a default token bucket limiter if none is provided
	if options.Limiter == nil {
		options.Limiter = NewTokenBucketLimiter(10, 10) // Default: 10 requests/second
	}

	return func(next HandlerFunc) HandlerFunc {
		return func(req *http.Request) (*http.Response, error) {
			ctx := req.Context()

			// Skip rate limiting if condition is met
			if options.SkipCondition != nil && options.SkipCondition(req) {
				return next(req)
			}

			// Attempt to acquire a token
			err := options.Limiter.Wait(ctx)
			if err != nil {
				if options.ErrorHandler != nil {
					return options.ErrorHandler(ctx, err)
				}
				return nil, fmt.Errorf("rate limit error: %w", err)
			}

			// Log rate limiting info in debug mode
			intlog.Printf(ctx, "Rate limiter allowed request to %s", req.URL.String())

			// Proceed with the request
			return next(req)
		}
	}
}

// WithGlobalRateLimit applies rate limiting to all requests from this client.
func (c *Client) WithGlobalRateLimit(rps float64, burst int) *Client {
	limiter := NewTokenBucketLimiter(rps, burst)
	c.Use(WithRateLimit(RateLimitOption{
		Limiter: limiter,
	}))
	return c
}
