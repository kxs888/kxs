package middleware_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kxs888/kxs/backend/internal/http/middleware"
	"github.com/kxs888/kxs/backend/internal/sm"
)

func ginServe(t *testing.T, mw gin.HandlerFunc, method, path string, body []byte, inner gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(mw)
	r.Handle(method, path, inner)
	var rdr *bytes.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	} else {
		rdr = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, rdr)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func TestSMDisabledPassthrough(t *testing.T) {
	rr := ginServe(t, middleware.SMCrypto(false, ""), http.MethodPost, "/api/v1/auth/login",
		[]byte(`{"username":"demo"}`),
		func(c *gin.Context) {
			b, _ := io.ReadAll(c.Request.Body)
			c.Data(http.StatusOK, "application/json", []byte(`{"ok":true,"echo":`+string(b)+`}`))
		})
	if rr.Header().Get("X-SM-Crypto") != "" {
		t.Fatal("disabled must not set sm header")
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte(`"username":"demo"`)) {
		t.Fatalf("passthrough %s", rr.Body.String())
	}
}

func TestSMEnabledRoundTrip(t *testing.T) {
	keyHex := "0123456789abcdeffedcba9876543210"
	key, err := sm.ParseKeyHex(keyHex)
	if err != nil {
		t.Fatal(err)
	}
	env, err := sm.Encrypt(key, []byte(`{"x":1}`))
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := sm.MarshalEnvelope(env)
	rr := ginServe(t, middleware.SMCrypto(true, keyHex), http.MethodPost, "/api/v1/ping", raw,
		func(c *gin.Context) {
			b, _ := io.ReadAll(c.Request.Body)
			if string(b) != `{"x":1}` {
				c.String(http.StatusBadRequest, "bad plain")
				return
			}
			c.Data(http.StatusOK, "application/json", []byte(`{"ok":true}`))
		})
	if rr.Code != 200 || rr.Header().Get("X-SM-Crypto") != "sm4-cbc" {
		t.Fatalf("status %d header %s body %s", rr.Code, rr.Header().Get("X-SM-Crypto"), rr.Body.String())
	}
	out, err := sm.UnmarshalEnvelope(rr.Body.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	pt, err := sm.Decrypt(key, out)
	if err != nil {
		t.Fatal(err)
	}
	if string(bytes.TrimSpace(pt)) != `{"ok":true}` {
		t.Fatalf("plain %s", pt)
	}
}

func TestSMSkipsProbes(t *testing.T) {
	for _, p := range []string{"/healthz", "/readyz", "/metrics", "/api/v1/stream", "/stream"} {
		if !middleware.SkipSMPath(p) {
			t.Fatalf("must skip %s", p)
		}
	}
	if middleware.SkipSMPath("/api/v1/auth/login") {
		t.Fatal("login is encrypted when layer 5 on")
	}
}
