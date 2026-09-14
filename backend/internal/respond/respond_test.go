package respond_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kxs888/kxs/backend/internal/errcode"
	"github.com/kxs888/kxs/backend/internal/obs"
	"github.com/kxs888/kxs/backend/internal/respond"
)

func testCtx(method, path string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rr := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rr)
	req := httptest.NewRequest(method, path, nil)
	c.Request = req
	return c, rr
}

func TestOKEnvelope(t *testing.T) {
	c, rr := testCtx(http.MethodGet, "/x")
	ctx := obs.WithRequestID(c.Request.Context(), "rid-1")
	ctx = obs.WithTraceID(ctx, "tid-1")
	c.Request = c.Request.WithContext(ctx)
	respond.OK(c, http.StatusOK, map[string]string{"status": "ok"})
	if rr.Code != 200 {
		t.Fatalf("status %d", rr.Code)
	}
	var env respond.Envelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if !env.OK || env.Code != errcode.OK {
		t.Fatalf("envelope %+v", env)
	}
	if env.Meta.RequestID != "rid-1" || env.Meta.TraceID != "tid-1" {
		t.Fatalf("meta %+v", env.Meta)
	}
	if env.Error != nil {
		t.Fatal("success must not include error object")
	}
	if string(env.Data) != `{"status":"ok"}` {
		t.Fatalf("data %s", env.Data)
	}
}

func TestErrEnvelope(t *testing.T) {
	c, rr := testCtx(http.MethodGet, "/x")
	ctx := obs.WithRequestID(c.Request.Context(), "rid-2")
	ctx = obs.WithTraceID(ctx, "tid-2")
	c.Request = c.Request.WithContext(ctx)
	respond.Err(c, errcode.NotImplemented("nope"))
	if rr.Code != 501 {
		t.Fatalf("status %d", rr.Code)
	}
	var env respond.Envelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if env.OK || env.Code != errcode.CommonNotImplemented {
		t.Fatalf("envelope %+v", env)
	}
	if env.Error == nil || env.Error.Code != errcode.CommonNotImplemented || env.Error.TraceID != "tid-2" {
		t.Fatalf("error %+v", env.Error)
	}
	if env.Meta.RequestID != "rid-2" || env.Meta.TraceID != "tid-2" {
		t.Fatalf("meta %+v", env.Meta)
	}
	if string(env.Data) != "null" {
		t.Fatalf("data %s", env.Data)
	}
}
