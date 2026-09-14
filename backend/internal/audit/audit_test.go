package audit_test

import (
	"context"
	"testing"

	"github.com/kxs888/kxs/backend/internal/audit"
)

func TestSanitizeDropsPHIAndJWT(t *testing.T) {
	got := audit.SanitizeDetail(map[string]any{
		"username":       "demo",
		"password":       "secret",
		"access_token":   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.aaa.bbb",
		"jwt":            "nope",
		"medical_record": "病历全文",
		"note_full":      "评估笔记",
		"id":             "abc",
		"token":          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.ccc.ddd",
	})
	if _, ok := got["username"]; !ok {
		t.Fatal("keep username")
	}
	if _, ok := got["id"]; !ok {
		t.Fatal("keep id")
	}
	for _, k := range []string{"password", "access_token", "jwt", "medical_record", "note_full", "token"} {
		if _, ok := got[k]; ok {
			t.Fatalf("should drop %s", k)
		}
	}
}

func TestRecordLoginAndMemoryStore(t *testing.T) {
	mem := audit.NewMemory()
	svc := audit.New(mem)
	if err := svc.Record(context.Background(), audit.Record{
		Action:       "auth.login_failed",
		ResourceType: "user",
		ResourceID:   "demo",
		Detail:       map[string]any{"username": "demo", "password": "x", "jwt": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.a.b"},
	}); err != nil {
		t.Fatal(err)
	}
	rows := mem.Snapshot()
	if len(rows) != 1 || rows[0].Action != "auth.login_failed" {
		t.Fatalf("%+v", rows)
	}
	if _, ok := rows[0].Detail["password"]; ok {
		t.Fatal("password leaked")
	}
	if _, ok := rows[0].Detail["jwt"]; ok {
		t.Fatal("jwt leaked")
	}
	if rows[0].Detail["username"] != "demo" {
		t.Fatalf("username %v", rows[0].Detail["username"])
	}
}
