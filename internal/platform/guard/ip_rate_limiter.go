package guard

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type counter struct {
	count      int
	windowEnds time.Time
	lastSeen   time.Time
}

type IPRateLimiter struct {
	mu      sync.Mutex
	windows map[string]counter
}

func NewIPRateLimiter() *IPRateLimiter {
	return &IPRateLimiter{windows: make(map[string]counter)}
}

// Middleware enforces two fixed-window limits per client IP:
//   - globalLimit: all requests in a window
//   - routeLimit: per method+route requests in a window
//
// For write-heavy methods (POST/PATCH/DELETE), route limit is halved.
func (l *IPRateLimiter) Middleware(globalLimit int, routeLimit int, window time.Duration) gin.HandlerFunc {
	if globalLimit <= 0 {
		globalLimit = 1200
	}
	if routeLimit <= 0 {
		routeLimit = 300
	}
	if window <= 0 {
		window = time.Minute
	}

	return func(c *gin.Context) {
		now := time.Now()
		ip := clientIP(c)
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		okGlobal, retryGlobal := l.allow("g:"+ip, globalLimit, window, now)
		if !okGlobal {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"statusCode":        http.StatusTooManyRequests,
				"message":           "Too many requests (global ip)",
				"retryAfterSeconds": retryGlobal,
			})
			return
		}

		limit := routeLimit
		switch c.Request.Method {
		case http.MethodPost, http.MethodPatch, http.MethodDelete:
			limit = routeLimit / 2
			if limit < 50 {
				limit = 50
			}
		}

		okRoute, retryRoute := l.allow("r:"+ip+":"+c.Request.Method+":"+route, limit, window, now)
		if !okRoute {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"statusCode":        http.StatusTooManyRequests,
				"message":           "Too many requests (route ip)",
				"retryAfterSeconds": retryRoute,
			})
			return
		}

		if now.Unix()%30 == 0 {
			l.prune(now, 10*window)
		}

		c.Next()
	}
}

func (l *IPRateLimiter) allow(key string, limit int, window time.Duration, now time.Time) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	item, ok := l.windows[key]
	if !ok || now.After(item.windowEnds) {
		item = counter{count: 0, windowEnds: now.Add(window)}
	}
	if item.count >= limit {
		retry := int(time.Until(item.windowEnds).Seconds())
		if retry < 1 {
			retry = 1
		}
		item.lastSeen = now
		l.windows[key] = item
		return false, retry
	}
	item.count++
	item.lastSeen = now
	l.windows[key] = item
	return true, 0
}

func (l *IPRateLimiter) prune(now time.Time, ttl time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for k, v := range l.windows {
		if now.Sub(v.lastSeen) > ttl {
			delete(l.windows, k)
		}
	}
}

func clientIP(c *gin.Context) string {
	if ip := strings.TrimSpace(c.ClientIP()); ip != "" {
		return ip
	}
	host := c.Request.RemoteAddr
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return host
}
