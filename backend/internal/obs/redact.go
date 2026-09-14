package obs

import (
	"context"
	"log/slog"
	"strings"
)

var secretKeyNames = map[string]struct{}{
	"authorization": {},
	"token":         {},
	"access_token":  {},
	"refresh_token": {},
	"jwt":           {},
	"id_token":      {},
	"password":      {},
	"passwd":        {},
	"secret":        {},
	"sm4_key":       {},
	"sm2_key":       {},
	"private_key":   {},
	"jwt_secret":    {},
}

const redacted = "[REDACTED]"

type redactHandler struct {
	inner slog.Handler
}

func newRedactHandler(inner slog.Handler) slog.Handler {
	return &redactHandler{inner: inner}
}

func (h *redactHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.inner.Enabled(ctx, l)
}

func (h *redactHandler) Handle(ctx context.Context, rec slog.Record) error {
	if LooksLikeJWT(rec.Message) {
		rec.Message = redactJWTIn(rec.Message)
	}
	return h.inner.Handle(ctx, rec)
}

func (h *redactHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	out := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		out[i] = redactAttrVal(a)
	}
	return &redactHandler{inner: h.inner.WithAttrs(out)}
}

func (h *redactHandler) WithGroup(name string) slog.Handler {
	return &redactHandler{inner: h.inner.WithGroup(name)}
}

func redactAttr(groups []string, a slog.Attr) slog.Attr {
	_ = groups
	return redactAttrVal(a)
}

func redactAttrVal(a slog.Attr) slog.Attr {
	key := strings.ToLower(a.Key)
	if _, ok := secretKeyNames[key]; ok {
		return slog.String(a.Key, redacted)
	}
	if a.Value.Kind() == slog.KindString {
		s := a.Value.String()
		if LooksLikeJWT(s) || looksLikeBearer(s) {
			return slog.String(a.Key, redactJWTIn(s))
		}
	}
	return a
}

func looksLikeBearer(s string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(s)), "bearer ")
}

// LooksLikeJWT 识别三点分段的 JWT，供测试与日志脱敏共用。
func LooksLikeJWT(s string) bool {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(strings.ToLower(s), "bearer ") {
		s = strings.TrimSpace(s[7:])
	}
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
	}
	// JWT header 通常以 eyJ 开头（{"alg"... 的 base64）。
	return strings.HasPrefix(parts[0], "eyJ") || strings.HasPrefix(parts[0], "eyJ")
}

func redactJWTIn(s string) string {
	if looksLikeBearer(s) {
		return "Bearer " + redacted
	}
	if LooksLikeJWT(s) {
		return redacted
	}
	return s
}
