package middleware

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

func AdmissionQueue(maxConcurrent int, maxQueue int, waitTimeout time.Duration) gin.HandlerFunc {
	if maxConcurrent <= 0 {
		maxConcurrent = 16
	}
	if maxQueue <= 0 {
		maxQueue = 128
	}
	if waitTimeout <= 0 {
		waitTimeout = 3 * time.Second
	}

	sem := make(chan struct{}, maxConcurrent)
	var queued int32

	return func(c *gin.Context) {
		q := atomic.AddInt32(&queued, 1)
		if q > int32(maxQueue) {
			atomic.AddInt32(&queued, -1)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"statusCode": http.StatusServiceUnavailable, "message": "Server busy, try again shortly"})
			return
		}

		timer := time.NewTimer(waitTimeout)
		defer timer.Stop()
		select {
		case sem <- struct{}{}:
			atomic.AddInt32(&queued, -1)
			defer func() { <-sem }()
			c.Next()
		case <-timer.C:
			atomic.AddInt32(&queued, -1)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"statusCode": http.StatusServiceUnavailable, "message": "Request queue timeout"})
			return
		}
	}
}
