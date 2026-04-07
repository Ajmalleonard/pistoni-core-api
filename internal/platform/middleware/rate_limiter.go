package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type fixedWindowCounter struct {
	count      int
	windowEnds time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	windows map[string]fixedWindowCounter
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{windows: map[string]fixedWindowCounter{}}
}

func (l *RateLimiter) Middleware(globalLimit int, routeLimit int, window time.Duration) gin.HandlerFunc {
	if globalLimit <= 0 {
		globalLimit = 240
	}
	if routeLimit <= 0 {
		routeLimit = 120
	}
	if window <= 0 {
		window = time.Minute
	}
	return func(c *gin.Context) {
		now := time.Now()
		ip := c.ClientIP()
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		allowedGlobal, retryGlobal := l.allow("global:"+ip, globalLimit, window, now)
		if !allowedGlobal {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"statusCode": http.StatusTooManyRequests, "message": "Too many requests (global)", "retryAfterSeconds": retryGlobal})
			return
		}

		limit := routeLimit
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPatch || c.Request.Method == http.MethodDelete {
			limit = routeLimit / 2
			if limit < 20 {
				limit = 20
			}
		}
		allowedRoute, retryRoute := l.allow("route:"+ip+":"+c.Request.Method+":"+route, limit, window, now)
		if !allowedRoute {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"statusCode": http.StatusTooManyRequests, "message": "Too many requests (route)", "retryAfterSeconds": retryRoute})
			return
		}
		c.Next()
	}
}

func (l *RateLimiter) allow(key string, limit int, window time.Duration, now time.Time) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	item, ok := l.windows[key]
	if !ok || now.After(item.windowEnds) {
		item = fixedWindowCounter{count: 0, windowEnds: now.Add(window)}
	}
	if item.count >= limit {
		retry := int(time.Until(item.windowEnds).Seconds())
		if retry < 1 {
			retry = 1
		}
		l.windows[key] = item
		return false, retry
	}
	item.count++
	l.windows[key] = item
	return true, 0
}
