package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/shuhei/rate-limiter/limiter"
)

func main() {
	var (
		algo        = flag.String("algo", "tokenbucket", "limiter algorithm: tokenbucket|fixedwindow")
		key         = flag.String("key", "user:123", "rate limit key")
		rate        = flag.Float64("rate", 5, "tokens per second")
		burst       = flag.Int("burst", 10, "max burst tokens")
		limit       = flag.Int("limit", 10, "max requests per window (fixedwindow)")
		window      = flag.Duration("window", 1*time.Second, "fixed window size (fixedwindow)")
		interval    = flag.Duration("interval", 100*time.Millisecond, "request interval")
		count       = flag.Int("count", 50, "number of requests")
		sleepOnDeny = flag.Bool("sleep-on-deny", false, "sleep retryAfter when denied")
	)
	flag.Parse()

	var (
		l   limiter.Limiter
		err error
	)

	switch *algo {
	case "tokenbucket":
		l, err = limiter.NewTokenBucket(limiter.TokenBucketConfig{
			Rate:  *rate,
			Burst: *burst,
		})
		if err != nil {
			panic(err)
		}
		fmt.Printf("demo: algo=tokenbucket key=%s rate=%.4g burst=%d interval=%s count=%d\n", *key, *rate, *burst, interval.String(), *count)
	case "fixedwindow":
		l, err = limiter.NewFixedWindow(limiter.FixedWindowConfig{
			Limit:  *limit,
			Window: *window,
		})
		if err != nil {
			panic(err)
		}
		fmt.Printf("demo: algo=fixedwindow key=%s limit=%d window=%s interval=%s count=%d\n", *key, *limit, window.String(), interval.String(), *count)
	default:
		fmt.Fprintf(os.Stderr, "invalid algo: %q (expected tokenbucket|fixedwindow)\n", *algo)
		os.Exit(2)
	}

	start := time.Now()
	for i := 0; i < *count; i++ {
		now := time.Since(start)
		allowed, retryAfter := l.Allow(*key)
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
