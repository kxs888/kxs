package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kxs888/kxs/backend/internal/auth"
	"github.com/kxs888/kxs/backend/internal/errcode"
	"github.com/kxs888/kxs/backend/internal/respond"
)

// RequireAuth 仅按路由局部挂载，禁止 engine.Use 全局套在整棵树上。
func RequireAuth(secret string, now func() time.Time) gin.HandlerFunc {
	if now == nil {
		now = time.Now
	}
	return func(c *gin.Context) {
		raw, err := auth.BearerToken(c.GetHeader("Authorization"))
		if err != nil {
			respond.Err(c, err)
			c.Abort()
			return
		}
		claims, err := auth.ParseAccess(secret, raw, now())
		if err != nil {
			respond.Err(c, err)
			c.Abort()
			return
		}
		id, err := uuid.Parse(claims.Subject)
		if err != nil {
			respond.Err(c, errcode.TokenInvalid())
			c.Abort()
			return
		}
		p := auth.Principal{UserID: id, Username: claims.Username}
		c.Request = c.Request.WithContext(auth.WithPrincipal(c.Request.Context(), p))
		c.Next()
	}
}
