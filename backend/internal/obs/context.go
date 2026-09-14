package obs

import (
	"context"
	"log/slog"
	"net/http"
)

type ctxKey int

const (
	keyRequestID ctxKey = iota
	keyLogger
	keyUserID
	keyTraceID
)

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyRequestID, id)
}

func RequestIDFrom(ctx context.Context) string {
	v, _ := ctx.Value(keyRequestID).(string)
	return v
}

func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, keyLogger, l)
}

func LoggerFrom(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(keyLogger).(*slog.Logger); ok && l != nil {
		return l
	}
	return App()
}

func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyUserID, id)
}

func UserIDFrom(ctx context.Context) string {
	v, _ := ctx.Value(keyUserID).(string)
	return v
}

func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyTraceID, id)
}

func TraceIDFrom(ctx context.Context) string {
	v, _ := ctx.Value(keyTraceID).(string)
	return v
}

func RequestIDFromHeader(r *http.Request) string {
	if v := r.Header.Get("X-Request-ID"); v != "" {
		return v
	}
	if v := r.Header.Get("X-Request-Id"); v != "" {
		return v
	}
	return ""
}
