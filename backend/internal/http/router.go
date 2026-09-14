package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kxs888/kxs/backend/internal/config"
	"github.com/kxs888/kxs/backend/internal/handler"
	"github.com/kxs888/kxs/backend/internal/http/middleware"
	"github.com/kxs888/kxs/backend/internal/idempotency"
	"github.com/kxs888/kxs/backend/internal/obs"
)

type Deps struct {
	Cfg         *config.Config
	API         *handler.API
	Idempotency idempotency.Store
}

// NewRouter 组装最多 4 层全局包装（国密开时第 5 层仅挂 /api/v1）。
// Auth / RBAC / 幂等 / 限流 / 审计 / 团队隔离禁止在此全局 Use。
func NewRouter(d Deps) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(
		middleware.Recover(),
		middleware.RequestIDTrace(),
		middleware.Limits(d.Cfg),
		middleware.AccessLog(),
	)

	r.GET("/healthz", d.API.Healthz)
	r.GET("/readyz", d.API.Readyz)
	r.GET("/metrics", d.API.Metrics)

	// SSE 永不走国密层。ticket 查询参数代替 Authorization，便于 EventSource。
	r.GET("/api/v1/stream", d.API.Stream)

	api := r.Group("/api/v1")
	if d.Cfg.SMCryptoEnabled {
		obs.App().Info("sm_crypto.enabled", "layer", 5)
		api.Use(middleware.SMCrypto(true, d.Cfg.SM4KeyHex))
	}

	api.POST("/auth/login", d.API.Login)

	authed := api.Group("")
	authed.Use(middleware.RequireAuth(d.Cfg.JWTSecret, nil))
	authed.GET("/me", d.API.Me)
	authed.POST("/stream/tickets", d.API.IssueStreamTicket)
	authed.GET("/tasks", d.API.Tasks)
	authed.POST("/ping-writes", idempotency.Middleware(d.Idempotency), d.API.PingWrite)

	return r
}
