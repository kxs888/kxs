package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kxs888/kxs/backend/internal/config"
)

// Limits 是第 3 层：Timeout + BodyLimit + 安全头 / CORS。合并为单一包装，避免拆成多层全局中间件。
func Limits(cfg *config.Config) gin.HandlerFunc {
	timeout := cfg.HTTPTimeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	limit := cfg.HTTPBodyLimit
	if limit <= 0 {
		limit = 1 << 20
	}
	origins := cfg.CORSOrigins
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Cache-Control", "no-store")
		applyCORS(c, origins)
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func applyCORS(c *gin.Context, origins []string) {
	origin := c.GetHeader("Origin")
	if origin == "" {
		return
	}
	allow := ""
	for _, o := range origins {
		if o == "*" || strings.EqualFold(o, origin) {
			allow = origin
			break
		}
	}
	if allow == "" {
		return
	}
	c.Header("Access-Control-Allow-Origin", allow)
	c.Header("Vary", "Origin")
	c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, X-Request-ID")
	c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	c.Header("Access-Control-Max-Age", strconv.Itoa(600))
}
