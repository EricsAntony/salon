package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// RateLimiter represents a configurable in-memory rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// RateLimitConfig configures rate limiting behavior
type RateLimitConfig struct {
	Limit  int
	Window time.Duration
}

// NewRateLimiter creates a new rate limiter with the given configuration
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	
	// Start cleanup goroutine
	go rl.cleanup()
	
	return rl
}

// Allow checks if a request should be allowed for the given key
func (rl *RateLimiter) Allow(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now()
	cutoff := now.Add(-rl.window)
	
	// Get existing requests for this key
	requests := rl.requests[key]
	
	// Remove old requests
	validRequests := make([]time.Time, 0, len(requests))
	for _, req := range requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}
	
	// Check if we're under the limit
	if len(validRequests) >= rl.limit {
		return false
	}
	
	// Add this request
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests
	
	return true
}

// GetCurrentCount returns the current number of requests for a key
func (rl *RateLimiter) GetCurrentCount(key string) int {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()
	
	now := time.Now()
	cutoff := now.Add(-rl.window)
	
	requests := rl.requests[key]
	count := 0
	for _, req := range requests {
		if req.After(cutoff) {
			count++
		}
	}
	
	return count
}

// Reset clears all requests for a specific key
func (rl *RateLimiter) Reset(key string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	delete(rl.requests, key)
}

// cleanup removes old entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()
	
	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window)
		
		for key, requests := range rl.requests {
			validRequests := make([]time.Time, 0, len(requests))
			for _, req := range requests {
				if req.After(cutoff) {
					validRequests = append(validRequests, req)
				}
			}
			
			if len(validRequests) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = validRequests
			}
		}
		rl.mutex.Unlock()
	}
}

// OTPRateLimitMiddleware creates a rate limiting middleware for OTP requests
func OTPRateLimitMiddleware(rateLimiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP for rate limiting
			clientIP := GetClientIP(r)
			
			if !rateLimiter.Allow(clientIP) {
				log.Warn().
					Str("client_ip", clientIP).
					Int("current_count", rateLimiter.GetCurrentCount(clientIP)).
					Msg("OTP rate limit exceeded")
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// GenericRateLimitMiddleware creates a generic rate limiting middleware
func GenericRateLimitMiddleware(rateLimiter *RateLimiter, keyExtractor func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyExtractor(r)
			
			if !rateLimiter.Allow(key) {
				log.Warn().
					Str("rate_limit_key", key).
					Int("current_count", rateLimiter.GetCurrentCount(key)).
					Msg("rate limit exceeded")
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// IPBasedKeyExtractor extracts client IP as the rate limiting key
func IPBasedKeyExtractor(r *http.Request) string {
	return GetClientIP(r)
}

// UserBasedKeyExtractor extracts user ID as the rate limiting key
func UserBasedKeyExtractor(r *http.Request) string {
	if userID := r.Context().Value("user_id"); userID != nil {
		return userID.(string)
	}
	return GetClientIP(r) // Fallback to IP
}
