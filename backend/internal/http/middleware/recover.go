package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/kxs888/kxs/backend/internal/errcode"
	"github.com/kxs888/kxs/backend/internal/obs"
	"github.com/kxs888/kxs/backend/internal/respond"
	"runtime/debug"
)

// Recover 是第 1 层 HTTP 包装：捕获 panic，不向客户端泄漏栈。
func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				obs.App().ErrorContext(c.Request.Context(), "panic",
					"panic", panicString(rec),
					"stack", string(debug.Stack()),
				)
				respond.Err(c, errcode.Internal("internal error"))
				c.Abort()
			}
		}()
		c.Next()
	}
}

func panicString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case error:
		return t.Error()
	default:
		return "panic"
	}
}
