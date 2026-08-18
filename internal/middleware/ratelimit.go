package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// fixedWindowRateLimiter is a simple in-memory, per-IP limiter. It resets on
// restart, which is acceptable at this scale (see docs/roadmap.md).
type fixedWindowRateLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	requests map[string]*windowEntry
}

type windowEntry struct {
	count      int
	windowStart time.Time
}

func newFixedWindowRateLimiter(limit int, window time.Duration) *fixedWindowRateLimiter {
	return &fixedWindowRateLimiter{
		limit:    limit,
		window:   window,
		requests: make(map[string]*windowEntry),
	}
}

// allow reports whether the given key may proceed within the current window.
func (l *fixedWindowRateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	e, ok := l.requests[key]
	if !ok || now.Sub(e.windowStart) >= l.window {
		l.requests[key] = &windowEntry{count: 1, windowStart: now}
		return true
	}

	e.count++
	return e.count <= l.limit
}

// RateLimit returns a middleware that limits each client IP to `limit`
// requests per `window` for the routes it is applied to.
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	l := newFixedWindowRateLimiter(limit, window)

	return func(c *gin.Context) {
		key := c.ClientIP()
		if !l.allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please wait and try again.",
			})
			return
		}
		c.Next()
	}
}