package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kxs888/kxs/backend/internal/obs"
)

// AccessLog 是第 4 层。禁止记录 Authorization / JWT / 请求体。
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		obs.Access().InfoContext(c.Request.Context(), "http_access",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"bytes", c.Writer.Size(),
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", obs.RequestIDFrom(c.Request.Context()),
			"user_id", obs.UserIDFrom(c.Request.Context()),
			"remote", c.Request.RemoteAddr,
		)
	}
}
