package auth_test

import (
	"testing"

	"github.com/kxs888/kxs/backend/internal/auth"
)

func TestPlaceholderPermissions(t *testing.T) {
	want := []string{"patient.view", "task.create", "report.view"}
	got := auth.PermissionsOrPlaceholder(nil)
	if len(got) != 3 {
		t.Fatalf("got %v", got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Fatalf("perm %d: %s", i, got[i])
		}
	}
	custom := []string{"task.create"}
	if auth.PermissionsOrPlaceholder(custom)[0] != "task.create" {
		t.Fatal("keep provided")
	}
}
