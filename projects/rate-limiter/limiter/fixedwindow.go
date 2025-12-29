package limiter

import (
	"errors"
	"math"
	"sync"
	"time"
)

type FixedWindowConfig struct {
	// Limit is max allowed requests per window.
	Limit int
	// Window is the fixed window size.
	Window time.Duration
	// Clock is optional. If nil, SystemClock is used.
	Clock Clock

	// StateTTL enables lazy cleanup of per-key state when > 0.
	StateTTL time.Duration
	// CleanupInterval controls cleanup frequency when StateTTL > 0.
	// If <= 0, it defaults to StateTTL.
	CleanupInterval time.Duration
}

type FixedWindow struct {
	mu sync.Mutex

	clock           Clock
	limit           int
	window          time.Duration
	stateTTL        time.Duration
	cleanupInterval time.Duration

	lastCleanup time.Time
	states      map[string]*fixedWindowState
}

type fixedWindowState struct {
	windowStart time.Time
	count       int
	lastSeen    time.Time
}

func NewFixedWindow(cfg FixedWindowConfig) (*FixedWindow, error) {
	if cfg.Limit < 1 {
		return nil, errors.New("limit must be >= 1")
	}
	if cfg.Window <= 0 {
		return nil, errors.New("window must be > 0")
	}

	clk := cfg.Clock
	if clk == nil {
		clk = SystemClock()
	}

	fw := &FixedWindow{
		clock:  clk,
		limit:  cfg.Limit,
		window: cfg.Window,
		states: make(map[string]*fixedWindowState),
	}

	if cfg.StateTTL > 0 {
		fw.stateTTL = cfg.StateTTL
		fw.cleanupInterval = cfg.CleanupInterval
		if fw.cleanupInterval <= 0 {
			fw.cleanupInterval = fw.stateTTL
		}
		fw.lastCleanup = fw.clock.Now()
	}

	return fw, nil
}

func (fw *FixedWindow) Allow(key string) (bool, time.Duration) {
	now := fw.clock.Now()
	windowStart := now.Truncate(fw.window)

	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.stateTTL > 0 && now.Sub(fw.lastCleanup) >= fw.cleanupInterval {
		fw.cleanupLocked(now)
		fw.lastCleanup = now
	}

	st := fw.states[key]
	if st == nil {
		st = &fixedWindowState{
			windowStart: windowStart,
			count:       0,
			lastSeen:    now,
		}
		fw.states[key] = st
	} else {
		st.lastSeen = now
		if windowStart.After(st.windowStart) {
			st.windowStart = windowStart
			st.count = 0
		} else if windowStart.Before(st.windowStart) {
			// Clock went backwards. Align to current computed windowStart.
			st.windowStart = windowStart
			st.count = 0
		}
	}

	if st.count < fw.limit {
		st.count++
		return true, 0
	}

	end := st.windowStart.Add(fw.window)
	seconds := end.Sub(now).Seconds()
	retryNs := math.Ceil(seconds * float64(time.Second))
	if retryNs < 1 {
		retryNs = 1
	}
	return false, time.Duration(retryNs)
}

func (fw *FixedWindow) cleanupLocked(now time.Time) {
	if fw.stateTTL <= 0 {
		return
	}
	for k, st := range fw.states {
		if now.Sub(st.lastSeen) > fw.stateTTL {
			delete(fw.states, k)
		}
	}
}

// stateCount is for tests.
func (fw *FixedWindow) stateCount() int {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	return len(fw.states)
}
