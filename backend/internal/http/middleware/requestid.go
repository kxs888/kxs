package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/kxs888/kxs/backend/internal/crypto"
	"github.com/kxs888/kxs/backend/internal/obs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// RequestIDTrace 是第 2 层：RequestID + OTel span。endpoint 为空时 tracer 为 noop。
func RequestIDTrace() gin.HandlerFunc {
	tr := otel.Tracer("cga-api/http")
	prop := otel.GetTextMapPropagator()
	return func(c *gin.Context) {
		rid := obs.RequestIDFromHeader(c.Request)
		if rid == "" {
			rid, _ = crypto.RandomHex(16)
		}
		ctx := obs.WithRequestID(c.Request.Context(), rid)
		ctx = prop.Extract(ctx, propagation.HeaderCarrier(c.Request.Header))
		ctx, span := tr.Start(ctx, c.Request.Method+" "+c.Request.URL.Path, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()
		if sc := span.SpanContext(); sc.IsValid() {
			ctx = obs.WithTraceID(ctx, sc.TraceID().String())
		}
		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Request-ID", rid)
		c.Next()
	}
}
