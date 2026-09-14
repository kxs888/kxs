package idempotency_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kxs888/kxs/backend/internal/errcode"
	"github.com/kxs888/kxs/backend/internal/idempotency"
	"github.com/kxs888/kxs/backend/internal/obs"
	"github.com/kxs888/kxs/backend/internal/respond"
)

func TestDefaultTTL24h(t *testing.T) {
	if idempotency.DefaultTTL != 24*time.Hour {
		t.Fatalf("C6 DefaultTTL %s", idempotency.DefaultTTL)
	}
}

func TestMemoryConflictDoesNotCompleteOverwrite(t *testing.T) {
	m := idempotency.NewMemory()
	ctx := context.Background()
	rec, isNew, err := m.BeginOrGet(ctx, "idem-key-abc", "fp-a")
	if err != nil || !isNew {
		t.Fatalf("begin %+v %v %v", rec, isNew, err)
	}
	if rec.ExpiresAt.Sub(time.Now()) < 23*time.Hour {
		t.Fatalf("expires_at %s", rec.ExpiresAt)
	}
	if err := m.Complete(ctx, "idem-key-abc", 201, []byte(`{"ok":true}`)); err != nil {
		t.Fatal(err)
	}
	_, _, err = m.BeginOrGet(ctx, "idem-key-abc", "fp-b")
	if err == nil {
		t.Fatal("want conflict")
	}
	var e *errcode.Error
	if !asErr(err, &e) || e.Code != errcode.IdempotencyKeyConflict || e.HTTP != 409 {
		t.Fatalf("err %v", err)
	}
	got, isNew, err := m.BeginOrGet(ctx, "idem-key-abc", "fp-a")
	if err != nil || isNew || !got.Completed || string(got.ResponseBody) != `{"ok":true}` {
		t.Fatalf("must keep original %+v %v %v", got, isNew, err)
	}
	_ = m.Complete(ctx, "idem-key-abc", 409, []byte(`{"ok":false}`))
	got, _, err = m.BeginOrGet(ctx, "idem-key-abc", "fp-a")
	if err != nil || string(got.ResponseBody) != `{"ok":true}` {
		t.Fatalf("complete must not overwrite %+v", got)
	}
}

func asErr(err error, target **errcode.Error) bool {
	e, ok := err.(*errcode.Error)
	if !ok {
		return false
	}
	*target = e
	return true
}

func TestMiddlewareMissingKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		ctx := obs.WithRequestID(c.Request.Context(), "rid")
		ctx = obs.WithTraceID(ctx, "tid")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	r.POST("/ping-writes", idempotency.Middleware(idempotency.NewMemory()), func(c *gin.Context) {
		respond.OK(c, http.StatusCreated, map[string]string{"ok": "n"})
	})
	req := httptest.NewRequest(http.MethodPost, "/ping-writes", strings.NewReader(`{"message":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != 400 {
		t.Fatalf("status %d %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "IDEMPOTENCY_KEY_REQUIRED") {
		t.Fatalf("body %s", rr.Body.String())
	}
}
