package audit_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kxs888/kxs/backend/internal/audit"
	"github.com/kxs888/kxs/backend/internal/repo"
)

func TestOpenEmptyURLUsesMemory(t *testing.T) {
	st, closer, err := audit.Open(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer closer()
	if _, ok := st.(*audit.Memory); !ok {
		t.Fatalf("want Memory fallback, got %T", st)
	}
	if _, ok := audit.NewStore(nil).(*audit.Memory); !ok {
		t.Fatal("nil pool must Memory fallback")
	}
}

func TestMigrationSQLHasAuditLogsUpDown(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	dir := filepath.Join(filepath.Dir(file), "..", "..", "migrations")
	up, err := os.ReadFile(filepath.Join(dir, "000004_audit_logs.up.sql"))
	if err != nil {
		t.Fatal(err)
	}
	down, err := os.ReadFile(filepath.Join(dir, "000001_framework.down.sql"))
	if err != nil {
		t.Fatal(err)
	}
	for _, needle := range []string{"CREATE TABLE IF NOT EXISTS audit_logs", "detail_json", "request_id", "trace_id", "outcome"} {
		if !strings.Contains(string(up), needle) {
			t.Fatalf("up missing %s", needle)
		}
	}
	if !strings.Contains(string(down), "DROP TABLE IF EXISTS audit_logs") {
		t.Fatal("000001 down must drop audit_logs")
	}
	down4, err := os.ReadFile(filepath.Join(dir, "000004_audit_logs.down.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(down4), "DROP COLUMN IF EXISTS outcome") {
		t.Fatal("000004 down")
	}
}

func TestPostgresPersistsAndSanitizes(t *testing.T) {
	url := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if url == "" {
		url = "postgres://cga:cga@127.0.0.1:5432/cga?sslmode=disable"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Skip("no postgres:", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Skip("no postgres:", err)
	}
	if err := repo.Migrate(url); err != nil {
		t.Fatal(err)
	}
	st, closer, err := audit.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer closer()
	if _, ok := st.(*audit.Postgres); !ok {
		t.Fatalf("want Postgres, got %T", st)
	}
	svc := audit.New(st)
	if err := svc.Record(ctx, audit.Record{
		Action:       "auth.login",
		ResourceType: "user",
		ResourceID:   "demo",
		Outcome:      "success",
		Detail: map[string]any{
			"username":       "demo",
			"password":       "demo-pass-change-me",
			"access_token":   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.aaa.bbb",
			"medical_record": "病历全文",
		},
		RequestID: "rid-audit-pg",
		TraceID:   "tid-audit-pg",
	}); err != nil {
		t.Fatal(err)
	}
	var detail []byte
	var outcome, action string
	err = pool.QueryRow(ctx, `
		SELECT action, outcome, detail_json
		FROM audit_logs
		WHERE request_id = $1
		ORDER BY created_at DESC
		LIMIT 1`, "rid-audit-pg").Scan(&action, &outcome, &detail)
	if err != nil {
		t.Fatal(err)
	}
	if action != "auth.login" || outcome != "success" {
		t.Fatalf("row %s %s", action, outcome)
	}
	raw := string(detail)
	if strings.Contains(raw, "demo-pass") || strings.Contains(raw, "eyJ") || strings.Contains(raw, "病历") {
		t.Fatalf("forbidden detail %s", raw)
	}
	var m map[string]any
	if err := json.Unmarshal(detail, &m); err != nil {
		t.Fatal(err)
	}
	if m["username"] != "demo" {
		t.Fatalf("username %v", m["username"])
	}
}
