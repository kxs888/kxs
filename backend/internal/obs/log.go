package obs

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

const (
	ChannelApp    = "app"
	ChannelAccess = "access"
	ChannelAudit  = "audit"
)

var (
	mu    sync.RWMutex
	appL  *slog.Logger
	accL  *slog.Logger
	audL  *slog.Logger
	level = new(slog.LevelVar)
)

func init() {
	Configure(os.Stdout, slog.LevelInfo)
}

// Configure 初始化分渠道 JSON logger。所有渠道共用脱敏 Handler，禁止输出 JWT。
func Configure(w io.Writer, lvl slog.Level) {
	level.Set(lvl)
	opts := &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: redactAttr,
	}
	base := slog.NewJSONHandler(w, opts)
	h := newRedactHandler(base)
	mu.Lock()
	defer mu.Unlock()
	appL = slog.New(h).With("channel", ChannelApp)
	accL = slog.New(h).With("channel", ChannelAccess)
	audL = slog.New(h).With("channel", ChannelAudit)
}

func App() *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()
	return appL
}

func Access() *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()
	return accL
}

func AuditLog() *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()
	return audL
}

func SetLevel(lvl slog.Level) { level.Set(lvl) }

func ParseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func LoggerWithRequest(ctx context.Context) *slog.Logger {
	l := LoggerFrom(ctx)
	if rid := RequestIDFrom(ctx); rid != "" {
		l = l.With("request_id", rid)
	}
	if tid := TraceIDFrom(ctx); tid != "" {
		l = l.With("trace_id", tid)
	}
	if uid := UserIDFrom(ctx); uid != "" {
		l = l.With("user_id", uid)
	}
	return l
}
