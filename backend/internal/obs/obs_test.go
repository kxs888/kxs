package obs_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/kxs888/kxs/backend/internal/obs"
)

func TestMaskPII(t *testing.T) {
	type row struct {
		Name     string `json:"name"`
		Phone    string `json:"phone" pii:"mask"`
		Password string `json:"password" pii:"mask"`
	}
	got, ok := obs.Mask(row{Name: "ada", Phone: "13800000000", Password: "s3cret"}).(map[string]any)
	if !ok {
		t.Fatal("want map")
	}
	if got["name"] != "ada" {
		t.Fatalf("name %v", got["name"])
	}
	if got["phone"] != "[REDACTED]" || got["password"] != "[REDACTED]" {
		t.Fatalf("mask %+v", got)
	}
}

func TestRedactJWTInLogs(t *testing.T) {
	var buf bytes.Buffer
	obs.Configure(&buf, slog.LevelInfo)
	t.Cleanup(func() { obs.Configure(nilWriter{}, slog.LevelInfo) })
	tok := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIn0.signaturepart"
	obs.App().Info("issued", "access_token", tok, "authorization", "Bearer "+tok)
	s := buf.String()
	if strings.Contains(s, tok) || strings.Contains(s, "eyJhbGciOiJIUzI1Ni") {
		t.Fatalf("jwt leaked: %s", s)
	}
	if !strings.Contains(s, "[REDACTED]") {
		t.Fatalf("expected redacted: %s", s)
	}
}

type nilWriter struct{}

func (nilWriter) Write(p []byte) (int, error) { return len(p), nil }
