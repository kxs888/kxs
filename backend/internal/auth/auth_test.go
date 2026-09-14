package auth_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kxs888/kxs/backend/internal/auth"
	"github.com/kxs888/kxs/backend/internal/errcode"
)

func TestArgon2idRoundTrip(t *testing.T) {
	h, err := auth.HashPassword("correct horse")
	if err != nil {
		t.Fatal(err)
	}
	if !auth.VerifyPassword("correct horse", h) {
		t.Fatal("verify true")
	}
	if auth.VerifyPassword("wrong", h) {
		t.Fatal("verify false")
	}
	if !stringsHasPrefix(h, "$argon2id$") {
		t.Fatalf("hash format %s", h)
	}
}

func stringsHasPrefix(s, p string) bool {
	return len(s) >= len(p) && s[:len(p)] == p
}

func TestJWTShortTTL(t *testing.T) {
	secret := "0123456789abcdef0123456789abcdef"
	id := uuid.New()
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	tok, exp, err := auth.IssueAccess(secret, 15*time.Minute, id, "demo", now)
	if err != nil {
		t.Fatal(err)
	}
	if exp != 900 {
		t.Fatalf("expires_in %d", exp)
	}
	c, err := auth.ParseAccess(secret, tok, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if c.Username != "demo" || c.Subject != id.String() {
		t.Fatalf("claims %+v", c)
	}
	_, err = auth.ParseAccess(secret, tok, now.Add(16*time.Minute))
	if err == nil {
		t.Fatal("want expired")
	}
	var e *errcode.Error
	if !errorAs(err, &e) || e.Code != errcode.AuthTokenExpired {
		t.Fatalf("err %v", err)
	}
}

func errorAs(err error, target **errcode.Error) bool {
	if err == nil {
		return false
	}
	e, ok := err.(*errcode.Error)
	if !ok {
		return false
	}
	*target = e
	return true
}
