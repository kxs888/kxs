package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kxs888/kxs/backend/internal/http/middleware"
	"github.com/kxs888/kxs/backend/internal/obs"
)

func TestRequestIDTraceGeneratesAndPropagates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestIDTrace())
	r.GET("/healthz", func(c *gin.Context) {
		if obs.RequestIDFrom(c.Request.Context()) == "" || obs.TraceIDFrom(c.Request.Context()) == "" {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("status %d", rr.Code)
	}
	if rr.Header().Get("X-Request-ID") == "" || rr.Header().Get("X-Trace-Id") == "" {
		t.Fatalf("headers %+v", rr.Header())
	}
	tp := rr.Header().Get("Traceparent")
	if !strings.HasPrefix(tp, "00-") {
		t.Fatalf("traceparent %s", tp)
	}

	upTrace := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	upSpan := "bbbbbbbbbbbbbbbb"
	req2 := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req2.Header.Set("Traceparent", "00-"+upTrace+"-"+upSpan+"-01")
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)
	if got := rr2.Header().Get("X-Trace-Id"); got != upTrace {
		t.Fatalf("upstream trace got %s", got)
	}
}
