package mhttp

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graingo/maltose/internal/intlog"
)

// RateLimitConfig defines the configuration for rate limiting
type RateLimitConfig struct {
	// Rate defines the number of requests allowed per second
	Rate float64
	// Burst defines the maximum number of requests that can be processed at once
	Burst int
	// SkipFunc is an optional function to determine if rate limiting should be skipped
	SkipFunc func(*Request) bool
	// ErrorHandler is an optional function to handle rate limit errors
	ErrorHandler func(*Request)
}

// DefaultRateLimitConfig returns a default rate limit configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Rate:  100, // 100 requests per second
		Burst: 10,  // Allow burst of 10 requests
	}
}

// RateLimitMiddleware creates a middleware that implements rate limiting using a token bucket algorithm
func RateLimitMiddleware(config RateLimitConfig) MiddlewareFunc {
	if config.Rate <= 0 {
		config.Rate = 100 // Default to 100 requests per second
	}
	if config.Burst <= 0 {
		config.Burst = 10 // Default burst size
	}

	// Create a token bucket
	tokens := float64(config.Burst)
	lastRefill := time.Now()
	var mu sync.Mutex

	// Refill tokens at the specified rate
	refillInterval := time.Duration(float64(time.Second) / config.Rate)
	ticker := time.NewTicker(refillInterval)
	go func() {
		for range ticker.C {
			mu.Lock()
			now := time.Now()
			elapsed := now.Sub(lastRefill).Seconds()
			tokens += elapsed * config.Rate
			if tokens > float64(config.Burst) {
				tokens = float64(config.Burst)
			}
			lastRefill = now
			mu.Unlock()
		}
	}()

	return func(r *Request) {
		// Skip rate limiting if SkipFunc returns true
		if config.SkipFunc != nil && config.SkipFunc(r) {
			return
		}

		mu.Lock()
		if tokens < 1 {
			mu.Unlock()
			if config.ErrorHandler != nil {
				config.ErrorHandler(r)
			} else {
				r.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Too Many Requests",
				})
			}
			return
		}
		tokens--
		mu.Unlock()

		// Log rate limit info if debug is enabled
		if r.Request.Context() != nil {
			intlog.Printf(r.Request.Context(), "Rate limit: %.2f tokens remaining", tokens)
		}
	}
}

// RateLimitByIP creates a middleware that implements rate limiting per IP address
func RateLimitByIP(config RateLimitConfig) MiddlewareFunc {
	// Create a map to store rate limiters for each IP
	limiters := make(map[string]*rateLimiter)
	var mu sync.RWMutex

	return func(r *Request) {
		// Skip rate limiting if SkipFunc returns true
		if config.SkipFunc != nil && config.SkipFunc(r) {
			return
		}

		// Get client IP
		ip := r.ClientIP()

		// Get or create rate limiter for this IP
		mu.RLock()
		limiter, exists := limiters[ip]
		mu.RUnlock()

		if !exists {
			mu.Lock()
			limiter = &rateLimiter{
				tokens:     float64(config.Burst),
				lastRefill: time.Now(),
			}
			limiters[ip] = limiter
			mu.Unlock()
		}

		// Check rate limit
		if !limiter.allow(config.Rate, config.Burst) {
			if config.ErrorHandler != nil {
				config.ErrorHandler(r)
			} else {
				r.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Too Many Requests",
				})
			}
			return
		}
	}
}

// rateLimiter implements a simple token bucket rate limiter
type rateLimiter struct {
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

func (l *rateLimiter) allow(rate float64, burst int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastRefill).Seconds()
	l.tokens += elapsed * rate
	if l.tokens > float64(burst) {
		l.tokens = float64(burst)
	}
	l.lastRefill = now

	if l.tokens < 1 {
		return false
	}

	l.tokens--
	return true
}
