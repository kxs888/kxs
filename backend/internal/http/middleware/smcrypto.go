package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kxs888/kxs/backend/internal/errcode"
	"github.com/kxs888/kxs/backend/internal/obs"
	"github.com/kxs888/kxs/backend/internal/respond"
	"github.com/kxs888/kxs/backend/internal/sm"
)

// SMCrypto 是可选第 5 层，仅当 SM_CRYPTO_ENABLED=true 时挂到 /api/v1。
// healthz / readyz / metrics / stream 永不加密（由路由排除 + 路径跳过双重保证）。
func SMCrypto(enabled bool, keyHex string) gin.HandlerFunc {
	if !enabled {
		return func(c *gin.Context) { c.Next() }
	}
	key, err := sm.ParseKeyHex(keyHex)
	if err != nil {
		obs.App().Error("sm4.key_invalid", "err", err.Error())
		return func(c *gin.Context) {
			respond.Err(c, errcode.Internal("sm crypto misconfigured"))
			c.Abort()
		}
	}
	return func(c *gin.Context) {
		if SkipSMPath(c.Request.URL.Path) {
			c.Next()
			return
		}
		if c.Request.Body != nil && c.Request.ContentLength != 0 &&
			(c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut || c.Request.Method == http.MethodPatch) {
			raw, err := io.ReadAll(c.Request.Body)
			_ = c.Request.Body.Close()
			if err != nil {
				respond.Err(c, errcode.InvalidArgument("read body"))
				c.Abort()
				return
			}
			env, err := sm.UnmarshalEnvelope(raw)
			if err != nil {
				respond.Err(c, err)
				c.Abort()
				return
			}
			pt, err := sm.Decrypt(key, env)
			if err != nil {
				respond.Err(c, err)
				c.Abort()
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(pt))
			c.Request.ContentLength = int64(len(pt))
		}
		cw := &smWriter{ResponseWriter: c.Writer, status: 0}
		c.Writer = cw
		c.Next()
		if cw.hijacked {
			return
		}
		env, err := sm.Encrypt(key, cw.body.Bytes())
		if err != nil {
			respond.Err(c, errcode.Internal("sm encrypt"))
			c.Abort()
			return
		}
		out, err := sm.MarshalEnvelope(env)
		if err != nil {
			respond.Err(c, errcode.Internal("sm encode"))
			c.Abort()
			return
		}
		orig := cw.ResponseWriter
		orig.Header().Set("Content-Type", "application/json; charset=utf-8")
		orig.Header().Set("X-SM-Crypto", "sm4-cbc")
		orig.Header().Del("Content-Length")
		status := cw.status
		if status == 0 {
			status = http.StatusOK
		}
		orig.WriteHeader(status)
		_, _ = orig.Write(out)
		_, _ = orig.Write([]byte("\n"))
	}
}

func SkipSMPath(path string) bool {
	switch path {
	case "/healthz", "/readyz", "/metrics", "/stream":
		return true
	}
	if strings.HasPrefix(path, "/stream/") {
		return true
	}
	if path == "/api/v1/stream" || strings.HasPrefix(path, "/api/v1/stream/") {
		return true
	}
	return false
}

type smWriter struct {
	gin.ResponseWriter
	body     bytes.Buffer
	status   int
	hijacked bool
}

func (w *smWriter) Write(p []byte) (int, error) {
	return w.body.Write(p)
}

func (w *smWriter) WriteString(s string) (int, error) {
	return w.body.WriteString(s)
}

func (w *smWriter) WriteHeader(code int) {
	if w.status == 0 {
		w.status = code
	}
}

func (w *smWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *smWriter) Size() int { return w.body.Len() }

func (w *smWriter) Written() bool { return w.body.Len() > 0 || w.status != 0 }

func (w *smWriter) Flush() {
	w.hijacked = true
	w.ResponseWriter.Flush()
}
