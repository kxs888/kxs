// Package respond 提供统一 JSON 响应信封。所有 JSON API 均走 OK / Err。
package respond

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kxs888/kxs/backend/internal/crypto"
	"github.com/kxs888/kxs/backend/internal/errcode"
	"github.com/kxs888/kxs/backend/internal/obs"
)

// Envelope 是对外唯一 JSON 形状，与 api/openapi.yaml 保持一致。
type Envelope struct {
	OK      bool            `json:"ok"`
	Code    errcode.Code    `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Error   *ErrorBody      `json:"error,omitempty"`
	Meta    Meta            `json:"meta"`
}

// Meta 出现在成功与失败信封中，含 request_id 与 trace_id。
type Meta struct {
	RequestID string `json:"request_id"`
	TraceID   string `json:"trace_id"`
}

// ErrorBody 失败时的 error 对象，必须带 trace_id。
type ErrorBody struct {
	Code    errcode.Code `json:"code"`
	Message string       `json:"message"`
	TraceID string       `json:"trace_id"`
}

// OK 写入成功信封。data 为 nil 时输出 JSON null。
func OK(c *gin.Context, status int, data any) {
	raw, err := marshalData(data)
	if err != nil {
		Err(c, errcode.Internal("encode response"))
		return
	}
	meta := metaFrom(c)
	write(c, status, Envelope{
		OK:      true,
		Code:    errcode.OK,
		Message: "ok",
		Data:    raw,
		Meta:    meta,
	})
}

// Err 写入失败信封。未知错误一律映射为 COMMON_INTERNAL，避免泄漏内部细节。
func Err(c *gin.Context, err error) {
	var e *errcode.Error
	if !errors.As(err, &e) || e == nil {
		obs.App().ErrorContext(c.Request.Context(), "respond.untyped_error", "err", err.Error())
		e = errcode.Internal("internal error")
	}
	meta := metaFrom(c)
	write(c, e.HTTP, Envelope{
		OK:      false,
		Code:    e.Code,
		Message: e.Message,
		Data:    json.RawMessage("null"),
		Error:   &ErrorBody{Code: e.Code, Message: e.Message, TraceID: meta.TraceID},
		Meta:    meta,
	})
}

func metaFrom(c *gin.Context) Meta {
	ctx := c.Request.Context()
	rid := obs.RequestIDFrom(ctx)
	tid := obs.TraceIDFrom(ctx)
	if rid == "" {
		rid, _ = crypto.RandomHex(16)
	}
	if tid == "" {
		tid = rid
	}
	return Meta{RequestID: rid, TraceID: tid}
}

func marshalData(data any) (json.RawMessage, error) {
	if data == nil {
		return json.RawMessage("null"), nil
	}
	if raw, ok := data.(json.RawMessage); ok {
		if len(raw) == 0 {
			return json.RawMessage("null"), nil
		}
		return raw, nil
	}
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func write(c *gin.Context, status int, env Envelope) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	if env.Meta.RequestID != "" {
		c.Header("X-Request-ID", env.Meta.RequestID)
	}
	if env.Meta.TraceID != "" {
		c.Header("X-Trace-Id", env.Meta.TraceID)
	}
	c.Status(status)
	enc := json.NewEncoder(c.Writer)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(env); err != nil && !errors.Is(err, http.ErrBodyNotAllowed) {
		obs.App().ErrorContext(c.Request.Context(), "respond.encode", "err", err.Error())
	}
}
