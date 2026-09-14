package obs

import (
	"net/http"
	"strings"
)

func TraceIDFromHeader(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-Trace-Id")); v != "" {
		return v
	}
	if v := strings.TrimSpace(r.Header.Get("X-Trace-ID")); v != "" {
		return v
	}
	if tid, _, ok := ParseTraceparent(r.Header.Get("Traceparent")); ok {
		return tid
	}
	if tid, _, ok := ParseTraceparent(r.Header.Get("traceparent")); ok {
		return tid
	}
	return ""
}

// ParseTraceparent 解析 W3C traceparent：00-{32hex}-{16hex}-{flags}。
func ParseTraceparent(h string) (traceID, spanID string, ok bool) {
	h = strings.TrimSpace(h)
	if h == "" {
		return "", "", false
	}
	parts := strings.Split(h, "-")
	if len(parts) != 4 {
		return "", "", false
	}
	if len(parts[1]) != 32 || len(parts[2]) != 16 {
		return "", "", false
	}
	return strings.ToLower(parts[1]), strings.ToLower(parts[2]), true
}

func FormatTraceparent(traceID, spanID string) string {
	return "00-" + traceID + "-" + spanID + "-01"
}
