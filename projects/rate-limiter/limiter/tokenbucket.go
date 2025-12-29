package limiter

import (
	"errors"
	"math"
	"sync"
	"time"
)

type TokenBucketConfig struct {
	// Rate is tokens per second.
	Rate float64
	// Burst is the maximum token capacity (and initial tokens).
	Burst int
	// Clock is optional. If nil, SystemClock is used.
	Clock Clock

	// StateTTL enables lazy cleanup of per-key state when > 0.
	StateTTL time.Duration
	// CleanupInterval controls cleanup frequency when StateTTL > 0.
	// If <= 0, it defaults to StateTTL.
	CleanupInterval time.Duration
}

type TokenBucket struct {
	mu sync.Mutex

	clock           Clock
	rate            float64
	burst           float64
	stateTTL        time.Duration
	cleanupInterval time.Duration

	lastCleanup time.Time
	states      map[string]*tokenBucketState
}

type tokenBucketState struct {
	tokens   float64
	last     time.Time
	lastSeen time.Time
}

func NewTokenBucket(cfg TokenBucketConfig) (*TokenBucket, error) {
	if cfg.Rate <= 0 {
		return nil, errors.New("rate must be > 0")
	}
	if cfg.Burst < 1 {
		return nil, errors.New("burst must be >= 1")
	}

	clk := cfg.Clock
	if clk == nil {
		clk = SystemClock()
	}

	tb := &TokenBucket{
		clock:  clk,
		rate:   cfg.Rate,
		burst:  float64(cfg.Burst),
		states: make(map[string]*tokenBucketState),
	}

	if cfg.StateTTL > 0 {
		tb.stateTTL = cfg.StateTTL
		tb.cleanupInterval = cfg.CleanupInterval
		if tb.cleanupInterval <= 0 {
			tb.cleanupInterval = tb.stateTTL
		}
		tb.lastCleanup = tb.clock.Now()
	}

	return tb, nil
}

func (tb *TokenBucket) Allow(key string) (bool, time.Duration) {
	now := tb.clock.Now()

	tb.mu.Lock()
	defer tb.mu.Unlock()

	if tb.stateTTL > 0 && now.Sub(tb.lastCleanup) >= tb.cleanupInterval {
		tb.cleanupLocked(now)
		tb.lastCleanup = now
	}

	st := tb.states[key]
	if st == nil {
		st = &tokenBucketState{
			tokens:   tb.burst,
			last:     now,
			lastSeen: now,
		}
		tb.states[key] = st
	} else {
		tb.refillLocked(st, now)
		st.lastSeen = now
	}

	if st.tokens >= 1 {
		st.tokens -= 1
		return true, 0
	}

	need := 1 - st.tokens
	seconds := need / tb.rate
	retryNs := math.Ceil(seconds * float64(time.Second))
	if retryNs < 1 {
		retryNs = 1
	}
	return false, time.Duration(retryNs)
}

func (tb *TokenBucket) refillLocked(st *tokenBucketState, now time.Time) {
	if now.Before(st.last) {
		st.last = now
		return
	}

	elapsed := now.Sub(st.last)
	add := (float64(elapsed) / float64(time.Second)) * tb.rate
	if add <= 0 {
		st.last = now
		return
	}

	st.tokens = math.Min(tb.burst, st.tokens+add)
	st.last = now
}

func (tb *TokenBucket) cleanupLocked(now time.Time) {
	if tb.stateTTL <= 0 {
		return
	}
	for k, st := range tb.states {
		if now.Sub(st.lastSeen) > tb.stateTTL {
			delete(tb.states, k)
		}
	}
}

// stateCount is for tests.
func (tb *TokenBucket) stateCount() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return len(tb.states)
}
