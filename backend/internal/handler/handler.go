package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kxs888/kxs/backend/internal/audit"
	"github.com/kxs888/kxs/backend/internal/auth"
	"github.com/kxs888/kxs/backend/internal/errcode"
	"github.com/kxs888/kxs/backend/internal/obs"
	"github.com/kxs888/kxs/backend/internal/respond"
	"github.com/kxs888/kxs/backend/internal/service"
	"github.com/kxs888/kxs/backend/internal/stream"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type API struct {
	PingDB   Pinger
	Auth     *service.AuthService
	Ping     *service.PingService
	Streams  *stream.Service
	Validate *validator.Validate
}

func New(ping Pinger, authsvc *service.AuthService, pingsvc *service.PingService, streams *stream.Service) *API {
	return &API{
		PingDB:   ping,
		Auth:     authsvc,
		Ping:     pingsvc,
		Streams:  streams,
		Validate: validator.New(),
	}
}

func (a *API) Healthz(c *gin.Context) {
	respond.OK(c, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) Readyz(c *gin.Context) {
	if a.PingDB == nil {
		respond.Err(c, errcode.Unavailable("database unavailable"))
		return
	}
	if err := a.PingDB.Ping(c.Request.Context()); err != nil {
		obs.App().ErrorContext(c.Request.Context(), "readyz.db", "err", err.Error())
		respond.Err(c, errcode.Unavailable("database unreachable"))
		return
	}
	respond.OK(c, http.StatusOK, map[string]string{"status": "ready", "db": "ok"})
}

func (a *API) Metrics(c *gin.Context) {
	c.Header("Content-Type", "text/plain; version=0.0.4")
	c.String(http.StatusOK, "# HELP cga_up 1 if process is up\n# TYPE cga_up gauge\ncga_up 1\n")
}

type loginReq struct {
	Username string `json:"username" validate:"required,min=1,max=64"`
	Password string `json:"password" validate:"required,min=1,max=128"`
}

func (a *API) Login(c *gin.Context) {
	var req loginReq
	if err := decodeJSON(c, &req); err != nil {
		respond.Err(c, err)
		return
	}
	if err := a.Validate.Struct(req); err != nil {
		respond.Err(c, errcode.InvalidArgument("invalid login payload"))
		return
	}
	res, err := a.Auth.Login(c.Request.Context(), req.Username, req.Password, audit.ClientIP(c.Request))
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, http.StatusOK, res)
}

func (a *API) Me(c *gin.Context) {
	p, ok := auth.PrincipalFrom(c.Request.Context())
	if !ok {
		respond.Err(c, errcode.Unauthorized("unauthorized"))
		return
	}
	u, err := a.Auth.Me(c.Request.Context(), p.UserID)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, http.StatusOK, map[string]any{
		"id":           u.ID,
		"username":     u.Username,
		"display_name": u.DisplayName,
	})
}

type pingWriteReq struct {
	Message string `json:"message" validate:"required,min=1,max=256"`
}

func (a *API) PingWrite(c *gin.Context) {
	var req pingWriteReq
	if err := decodeJSON(c, &req); err != nil {
		respond.Err(c, err)
		return
	}
	if err := a.Validate.Struct(req); err != nil {
		respond.Err(c, errcode.InvalidArgument("invalid payload"))
		return
	}
	row, err := a.Ping.Create(c.Request.Context(), req.Message, audit.ClientIP(c.Request))
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, http.StatusCreated, row)
}

func (a *API) Tasks(c *gin.Context) {
	respond.OK(c, http.StatusOK, map[string]any{"items": []any{}})
}

func (a *API) NotImplemented(c *gin.Context) {
	respond.Err(c, errcode.NotImplemented("not implemented in slice 0"))
}

func (a *API) IssueStreamTicket(c *gin.Context) {
	p, ok := auth.PrincipalFrom(c.Request.Context())
	if !ok {
		respond.Err(c, errcode.Unauthorized("unauthorized"))
		return
	}
	t, err := a.Streams.Issue(c.Request.Context(), p.UserID)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, http.StatusOK, map[string]any{
		"ticket":     t.ID.String(),
		"expires_at": t.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		"expires_in": stream.DefaultTicketTTL,
		"note":       "use GET /api/v1/stream?ticket=... ; events never include PHI",
	})
}

func (a *API) Stream(c *gin.Context) {
	raw := c.Query("ticket")
	t, err := a.Streams.Validate(c.Request.Context(), raw)
	if err != nil {
		respond.Err(c, err)
		return
	}
	stream.WritePlaceholder(c.Writer, c.Request, t)
}

func decodeJSON(c *gin.Context, dst any) error {
	dec := json.NewDecoder(io.LimitReader(c.Request.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return errcode.InvalidArgument("invalid json")
	}
	return nil
}

func UserIDParam(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, errcode.InvalidArgument("invalid id")
	}
	return id, nil
}
