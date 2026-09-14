package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/kxs888/kxs/backend/internal/crypto"
	"github.com/kxs888/kxs/backend/internal/obs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// RequestIDTrace 是第 2 层：RequestID + Trace。无上游则生成，并写入 Traceparent / X-Trace-Id。
func RequestIDTrace() gin.HandlerFunc {
	tr := otel.Tracer("cga-api/http")
	prop := otel.GetTextMapPropagator()
	return func(c *gin.Context) {
		rid := obs.RequestIDFromHeader(c.Request)
		if rid == "" {
			rid, _ = crypto.RandomHex(16)
		}

		upstreamTID := obs.TraceIDFromHeader(c.Request)
		spanID, _ := crypto.RandomHex(8)

		ctx := obs.WithRequestID(c.Request.Context(), rid)
		ctx = prop.Extract(ctx, propagation.HeaderCarrier(c.Request.Header))
		ctx, span := tr.Start(ctx, c.Request.Method+" "+c.Request.URL.Path, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		tid := upstreamTID
		if sc := span.SpanContext(); sc.IsValid() {
			tid = sc.TraceID().String()
			spanID = sc.SpanID().String()
		}
		if tid == "" {
			tid, _ = crypto.RandomHex(16)
		}
		ctx = obs.WithTraceID(ctx, tid)
		c.Request = c.Request.WithContext(ctx)

		c.Header("X-Request-ID", rid)
		c.Header("X-Trace-Id", tid)
		c.Header("Traceparent", obs.FormatTraceparent(normalizeTraceID(tid), normalizeSpanID(spanID)))
		c.Next()
	}
}

func normalizeTraceID(id string) string {
	if len(id) == 32 {
		return id
	}
	if len(id) == 16 {
		return id + id
	}
	padded := id + "00000000000000000000000000000000"
	if len(padded) < 32 {
		return padded
	}
	return padded[:32]
}

func normalizeSpanID(id string) string {
	if len(id) >= 16 {
		return id[:16]
	}
	return (id + "0000000000000000")[:16]
}
