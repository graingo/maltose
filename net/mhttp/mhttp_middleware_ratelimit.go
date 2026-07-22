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

func normalizeRateLimitConfig(config RateLimitConfig) RateLimitConfig {
	if config.Rate <= 0 {
		config.Rate = 100
	}
	if config.Burst <= 0 {
		config.Burst = 10
	}
	return config
}

// DefaultRateLimitConfig returns a default rate limit configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Rate:  100, // 100 requests per second
		Burst: 10,  // Allow burst of 10 requests
	}
}

// MiddlewareRateLimit creates a middleware that implements rate limiting using a token bucket algorithm
func MiddlewareRateLimit(config RateLimitConfig) MiddlewareFunc {
	config = normalizeRateLimitConfig(config)
	limiter := &rateLimiter{tokens: float64(config.Burst), lastRefill: time.Now(), lastSeen: time.Now()}

	return func(r *Request) {
		// Skip rate limiting if SkipFunc returns true
		if config.SkipFunc != nil && config.SkipFunc(r) {
			return
		}

		if !limiter.allow(config.Rate, config.Burst) {
			if config.ErrorHandler != nil {
				config.ErrorHandler(r)
			} else {
				r.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Too Many Requests",
				})
			}
			r.Abort()
			return
		}

		// Log rate limit info if debug is enabled
		if r.Request.Context() != nil {
			intlog.Printf(r.Request.Context(), "Rate limiter allowed request")
		}
	}
}

// MiddlewareRateLimitByIP creates a middleware that implements rate limiting per IP address
func MiddlewareRateLimitByIP(config RateLimitConfig) MiddlewareFunc {
	config = normalizeRateLimitConfig(config)
	// Create a map to store rate limiters for each IP
	limiters := make(map[string]*rateLimiter)
	var mu sync.RWMutex
	lastCleanup := time.Now()

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
			if existing, ok := limiters[ip]; ok {
				limiter = existing
			} else {
				limiter = &rateLimiter{
					tokens:     float64(config.Burst),
					lastRefill: time.Now(),
					lastSeen:   time.Now(),
				}
				limiters[ip] = limiter
			}
			mu.Unlock()
		}

		mu.Lock()
		if time.Since(lastCleanup) >= time.Minute {
			cutoff := time.Now().Add(-10 * time.Minute)
			for key, candidate := range limiters {
				if candidate.lastAccessBefore(cutoff) {
					delete(limiters, key)
				}
			}
			lastCleanup = time.Now()
		}
		mu.Unlock()

		// Check rate limit
		if !limiter.allow(config.Rate, config.Burst) {
			if config.ErrorHandler != nil {
				config.ErrorHandler(r)
			} else {
				r.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Too Many Requests",
				})
			}
			r.Abort()
			return
		}
	}
}

// rateLimiter implements a simple token bucket rate limiter
type rateLimiter struct {
	tokens     float64
	lastRefill time.Time
	lastSeen   time.Time
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
	l.lastSeen = now

	if l.tokens < 1 {
		return false
	}

	l.tokens--
	return true
}

func (l *rateLimiter) lastAccessBefore(cutoff time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.lastSeen.Before(cutoff)
}
