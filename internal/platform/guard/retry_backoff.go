package guard

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type backoffEntry struct {
	lockedUntil time.Time
	level       int
	lastSeen    time.Time
}

type RetryBackoffLimiter struct {
	mu      sync.Mutex
	entries map[string]backoffEntry
}

func NewRetryBackoffLimiter() *RetryBackoffLimiter {
	return &RetryBackoffLimiter{entries: map[string]backoffEntry{}}
}

// Middleware applies adaptive lockout per IP+route+method.
// If same endpoint is retried during lockout, wait escalates: base, 2x, 4x... up to maxWait.
func (l *RetryBackoffLimiter) Middleware(baseWait, maxWait time.Duration) gin.HandlerFunc {
	if baseWait <= 0 {
		baseWait = 60 * time.Second
	}
	if maxWait <= 0 || maxWait < baseWait {
		maxWait = 15 * time.Minute
	}

	return func(c *gin.Context) {
		now := time.Now()
		ip := clientIP(c)
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}
		key := ip + ":" + c.Request.Method + ":" + route

		allowed, retry, level := l.allow(key, now, baseWait, maxWait)
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"statusCode":        http.StatusTooManyRequests,
				"message":           "Too many retries on this endpoint",
				"retryAfterSeconds": retry,
				"penaltyLevel":      level,
			})
			return
		}

		if now.Unix()%45 == 0 {
			l.prune(now, 2*maxWait)
		}
		c.Next()
	}
}

func (l *RetryBackoffLimiter) allow(key string, now time.Time, baseWait, maxWait time.Duration) (bool, int, int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := l.entries[key]
	if now.Before(entry.lockedUntil) {
		entry.level++
		next := baseWait * time.Duration(1<<(entry.level-1))
		if next > maxWait {
			next = maxWait
		}
		entry.lockedUntil = now.Add(next)
		entry.lastSeen = now
		l.entries[key] = entry
		retry := int(next.Seconds())
		if retry < 1 {
			retry = 1
		}
		return false, retry, entry.level
	}

	entry.level = 0
	entry.lockedUntil = now.Add(baseWait)
	entry.lastSeen = now
	l.entries[key] = entry
	return true, 0, entry.level
}

func (l *RetryBackoffLimiter) prune(now time.Time, ttl time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for k, v := range l.entries {
		if now.Sub(v.lastSeen) > ttl {
			delete(l.entries, k)
		}
	}
}
