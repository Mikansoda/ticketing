package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var requests = make(map[string]int)
var mu sync.Mutex

func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}

		mu.Lock()
		defer mu.Unlock()
        // kalau user baru (!ok), inisialisasi count = 0
        // buat goroutine yang ngehapus data user setelah durasi window lewat, counter reset otomatis
		if _, ok := requests[userID]; !ok {
			requests[userID] = 0
			go func(uid string) {
				time.Sleep(window)
				mu.Lock()
				delete(requests, uid)
				mu.Unlock()
			}(userID)
		}

		if requests[userID] >= limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		requests[userID]++
		c.Next()
	}
}
