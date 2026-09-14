package errcode_test

import (
	"strings"
	"testing"

	"github.com/kxs888/kxs/backend/internal/errcode"
)

func TestPrefixes(t *testing.T) {
	common := []errcode.Code{
		errcode.CommonInvalidArgument,
		errcode.CommonNotFound,
		errcode.CommonInternal,
		errcode.CommonNotImplemented,
		errcode.CommonUnavailable,
	}
	for _, c := range common {
		if !strings.HasPrefix(string(c), "COMMON_") {
			t.Fatalf("%s: want COMMON_ prefix", c)
		}
	}
	auth := []errcode.Code{errcode.AuthUnauthorized, errcode.AuthInvalidCredentials, errcode.AuthTokenExpired}
	for _, c := range auth {
		if !strings.HasPrefix(string(c), "AUTH_") {
			t.Fatalf("%s: want AUTH_ prefix", c)
		}
	}
	if !strings.HasPrefix(string(errcode.IdempotencyKeyRequired), "IDEMPOTENCY_") {
		t.Fatal("C4 code")
	}
	if !strings.HasPrefix(string(errcode.IdempotencyKeyConflict), "IDEMPOTENCY_") {
		t.Fatal("C3 code")
	}
}
