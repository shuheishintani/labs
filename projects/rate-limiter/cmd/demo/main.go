package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/shuhei/rate-limiter/limiter"
)

func main() {
	var (
		key         = flag.String("key", "user:123", "rate limit key")
		rate        = flag.Float64("rate", 5, "tokens per second")
		burst       = flag.Int("burst", 10, "max burst tokens")
		interval    = flag.Duration("interval", 100*time.Millisecond, "request interval")
		count       = flag.Int("count", 50, "number of requests")
		sleepOnDeny = flag.Bool("sleep-on-deny", false, "sleep retryAfter when denied")
	)
	flag.Parse()

	tb, err := limiter.NewTokenBucket(limiter.TokenBucketConfig{
		Rate:  *rate,
		Burst: *burst,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("token-bucket demo: key=%s rate=%.4g burst=%d interval=%s count=%d\n", *key, *rate, *burst, interval.String(), *count)

	start := time.Now()
	for i := 0; i < *count; i++ {
		now := time.Since(start)
		allowed, retryAfter := tb.Allow(*key)
		if allowed {
			fmt.Printf("%9s  allowed   retryAfter=%s\n", now.Truncate(time.Millisecond), retryAfter)
		} else {
			fmt.Printf("%9s  denied    retryAfter=%s\n", now.Truncate(time.Millisecond), retryAfter)
			if *sleepOnDeny && retryAfter > 0 {
				time.Sleep(retryAfter)
			}
		}
		time.Sleep(*interval)
	}
}
