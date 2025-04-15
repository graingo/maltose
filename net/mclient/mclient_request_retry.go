package mclient

import (
	"math/rand"
	"net/http"
	"time"
)

// -----------------------------------------------------------------------------
// Retry Configuration Methods
// -----------------------------------------------------------------------------

// RetryConfig is the configuration for request retry.
type RetryConfig struct {
	// Count is the maximum number of retries.
	// For example, if Count is 3, the request will be tried up to 4 times (initial attempt + 3 retries).
	Count int

	// BaseInterval is the base interval between retries.
	// This is the starting point for calculating the delay between retries.
	// For example, if BaseInterval is 1 second, the first retry will wait at least 1 second.
	BaseInterval time.Duration

	// MaxInterval is the maximum interval between retries.
	// This prevents the delay from growing too large due to exponential backoff.
	// For example, if MaxInterval is 30 seconds, even if the calculated delay is 60 seconds,
	// the actual delay will be capped at 30 seconds.
	MaxInterval time.Duration

	// BackoffFactor is the factor for exponential backoff.
	// Each retry's delay is calculated by multiplying the previous delay by this factor.
	// For example, if BaseInterval is 1 second and BackoffFactor is 2.0:
	// - First retry: 1 second
	// - Second retry: 2 seconds
	// - Third retry: 4 seconds
	// - And so on...
	BackoffFactor float64

	// JitterFactor is the factor for random jitter.
	// This adds randomness to the delay to prevent multiple clients from retrying simultaneously.
	// The actual jitter is calculated as: delay * JitterFactor * (random number between -1 and 1)
	// For example, if the calculated delay is 1 second and JitterFactor is 0.1:
	// - The actual delay will be between 0.9 and 1.1 seconds
	// A value of 0 means no jitter will be added.
	JitterFactor float64
}

// DefaultRetryConfig returns the default retry configuration.
// The default values are:
// - Count: 3 retries
// - BaseInterval: 1 second
// - MaxInterval: 30 seconds
// - BackoffFactor: 2.0 (doubles the delay each time)
// - JitterFactor: 0.1 (adds ±10% random jitter)
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		Count:         3,
		BaseInterval:  time.Second,
		MaxInterval:   30 * time.Second,
		BackoffFactor: 2.0,
		JitterFactor:  0.1,
	}
}

// SetRetry sets retry configuration.
func (r *Request) SetRetry(config RetryConfig) *Request {
	r.retryCount = config.Count
	r.retryInterval = config.BaseInterval
	r.retryConfig = config
	return r
}

// SetRetrySimple sets retry count and base interval with default backoff and jitter.
func (r *Request) SetRetrySimple(count int, baseInterval time.Duration) *Request {
	config := DefaultRetryConfig()
	config.Count = count
	config.BaseInterval = baseInterval
	return r.SetRetry(config)
}

// SetRetryCondition sets a custom retry condition function.
// The function takes the HTTP response and error as input and returns
// true if the request should be retried.
func (r *Request) SetRetryCondition(condition func(*http.Response, error) bool) *Request {
	r.retryCondition = condition
	return r
}

// shouldRetry determines if a request should be retried based on the response and error.
func (r *Request) shouldRetry(resp *http.Response, err error) bool {
	// Use custom condition if provided
	if r.retryCondition != nil {
		return r.retryCondition(resp, err)
	}

	// Default retry condition
	if err != nil {
		// Retry on network/connection errors
		return true
	}

	if resp != nil {
		// Retry on 5xx (server errors) and 429 (too many requests)
		return resp.StatusCode >= 500 || resp.StatusCode == 429
	}

	return false
}

// calculateRetryDelay calculates the delay for the next retry attempt.
func (r *Request) calculateRetryDelay(attempt int) time.Duration {
	// If no retry config, use simple interval
	if r.retryConfig == (RetryConfig{}) {
		return r.retryInterval
	}

	// Calculate exponential backoff
	delay := r.retryConfig.BaseInterval
	for i := 1; i < attempt; i++ {
		delay = time.Duration(float64(delay) * r.retryConfig.BackoffFactor)
		if delay > r.retryConfig.MaxInterval {
			delay = r.retryConfig.MaxInterval
			break
		}
	}

	// Add jitter
	if r.retryConfig.JitterFactor > 0 {
		jitter := time.Duration(float64(delay) * r.retryConfig.JitterFactor * (rand.Float64()*2 - 1))
		delay += jitter
		if delay < 0 {
			delay = 0
		}
	}

	return delay
}
