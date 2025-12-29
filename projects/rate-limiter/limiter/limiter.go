package limiter

import "time"

// Limiter is a per-key rate limiter.
//
// Allow returns whether the request is allowed now.
// If not allowed, retryAfter is an estimate until the next allowed time.
type Limiter interface {
	Allow(key string) (allowed bool, retryAfter time.Duration)
}
