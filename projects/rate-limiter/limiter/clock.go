package limiter

import (
	"sync"
	"time"
)

// Clock はテストで時間を制御するためのインターフェースです。
type Clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

// SystemClock は実時間の Clock を返します。
func SystemClock() Clock { return systemClock{} }

// ManualClock はテスト用の手動 Clock です。
type ManualClock struct {
	mu sync.Mutex
	t  time.Time
}

func NewManualClock(start time.Time) *ManualClock {
	return &ManualClock{t: start}
}

func (c *ManualClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *ManualClock) Add(d time.Duration) {
	c.mu.Lock()
	c.t = c.t.Add(d)
	c.mu.Unlock()
}

func (c *ManualClock) Set(t time.Time) {
	c.mu.Lock()
	c.t = t
	c.mu.Unlock()
}
