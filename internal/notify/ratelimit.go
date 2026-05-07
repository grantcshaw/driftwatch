package notify

import (
	"sync"
	"time"
)

// RateLimiter tracks per-environment send counts within a rolling window
// to prevent alert storms when many keys drift simultaneously.
type RateLimiter struct {
	mu       sync.Mutex
	window   time.Duration
	maxSends int
	records  map[string][]time.Time
}

// NewRateLimiter creates a RateLimiter that allows at most maxSends notifications
// per environment within the given window duration.
func NewRateLimiter(window time.Duration, maxSends int) (*RateLimiter, error) {
	if window <= 0 {
		return nil, fmt.Errorf("notify: rate limiter window must be positive")
	}
	if maxSends <= 0 {
		return nil, fmt.Errorf("notify: rate limiter maxSends must be positive")
	}
	return &RateLimiter{
		window:   window,
		maxSends: maxSends,
		records:  make(map[string][]time.Time),
	}, nil
}

// Allow returns true if a notification for the given environment is permitted
// under the current rate limit, and records the attempt if so.
func (r *RateLimiter) Allow(env string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	times := r.records[env]
	filtered := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) >= r.maxSends {
		r.records[env] = filtered
		return false
	}

	r.records[env] = append(filtered, now)
	return true
}

// Reset clears the rate limit history for the given environment.
func (r *RateLimiter) Reset(env string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.records, env)
}

// Remaining returns how many sends are still allowed for env within the current window.
func (r *RateLimiter) Remaining(env string) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)
	count := 0
	for _, t := range r.records[env] {
		if t.After(cutoff) {
			count++
		}
	}
	remaining := r.maxSends - count
	if remaining < 0 {
		return 0
	}
	return remaining
}
