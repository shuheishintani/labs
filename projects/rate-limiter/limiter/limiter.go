package limiter

import "time"

// Limiter は key 単位のレートリミッタです。
// Allow は「いま許可するか」と「次に許可されるまでの目安」を返します。
type Limiter interface {
	Allow(key string) (allowed bool, retryAfter time.Duration)
}
