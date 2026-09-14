package config_test

import (
	"os"
	"testing"

	"github.com/kxs888/kxs/backend/internal/config"
)

func TestSMDefaultOff(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://cga:cga@localhost:5432/cga?sslmode=disable")
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("SM_CRYPTO_ENABLED", "")
	os.Unsetenv("SM_CRYPTO_ENABLED")
	c, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.SMCryptoEnabled {
		t.Fatal("SM_CRYPTO_ENABLED must default false")
	}
}

func TestSMEnabledRequiresKey(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://cga:cga@localhost:5432/cga?sslmode=disable")
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("SM_CRYPTO_ENABLED", "true")
	t.Setenv("SM4_KEY_HEX", "")
	if _, err := config.Load(); err == nil {
		t.Fatal("want error")
	}
}
