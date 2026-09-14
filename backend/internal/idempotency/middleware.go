package idempotency

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kxs888/kxs/backend/internal/crypto"
	"github.com/kxs888/kxs/backend/internal/domain"
	"github.com/kxs888/kxs/backend/internal/errcode"
	"github.com/kxs888/kxs/backend/internal/respond"
)

type Store interface {
	BeginOrGet(ctx context.Context, key, fingerprint string) (rec *domain.IdempotencyRecord, isNew bool, err error)
	Complete(ctx context.Context, key string, status int, responseBody []byte) error
}

// Middleware 仅应挂在需要幂等的写路由上，禁止全局 Use。
func Middleware(store Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
		if key == "" {
			respond.Err(c, errcode.New(errcode.IdempotencyKeyRequired, 400, "Idempotency-Key header required"))
			c.Abort()
			return
		}
		if len(key) < 8 || len(key) > 128 {
			respond.Err(c, errcode.InvalidArgument("Idempotency-Key length must be 8-128"))
			c.Abort()
			return
		}
		body, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20+1))
		if err != nil {
			respond.Err(c, errcode.InvalidArgument("read body"))
			c.Abort()
			return
		}
		_ = c.Request.Body.Close()
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		fp := crypto.SHA256Hex([]byte(c.Request.Method), []byte(c.Request.URL.Path), body)
		rec, isNew, err := store.BeginOrGet(c.Request.Context(), key, fp)
		if err != nil {
			respond.Err(c, err)
			c.Abort()
			return
		}
		if !isNew && rec.Completed {
			c.Header("Content-Type", "application/json; charset=utf-8")
			c.Header("Idempotency-Replayed", "true")
			if rec.StatusCode == 0 {
				rec.StatusCode = http.StatusOK
			}
			c.Status(rec.StatusCode)
			_, _ = c.Writer.Write(rec.ResponseBody)
			if len(rec.ResponseBody) == 0 || rec.ResponseBody[len(rec.ResponseBody)-1] != '\n' {
				_, _ = c.Writer.Write([]byte("\n"))
			}
			c.Abort()
			return
		}
		if !isNew && !rec.Completed {
			respond.Err(c, errcode.New(errcode.IdempotencyKeyConflict, 409, "idempotency key in progress"))
			c.Abort()
			return
		}

		cw := &captureWriter{ResponseWriter: c.Writer}
		c.Writer = cw
		c.Next()
		status := cw.Status()
		if status == 0 {
			status = http.StatusOK
		}
		_ = store.Complete(c.Request.Context(), key, status, cw.buf.Bytes())
	}
}

type captureWriter struct {
	gin.ResponseWriter
	buf bytes.Buffer
}

func (w *captureWriter) Write(p []byte) (int, error) {
	_, _ = w.buf.Write(p)
	return w.ResponseWriter.Write(p)
}

func (w *captureWriter) WriteString(s string) (int, error) {
	_, _ = w.buf.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
