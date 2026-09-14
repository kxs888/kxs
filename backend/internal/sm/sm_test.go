package sm_test

import (
	"bytes"
	"testing"

	"github.com/kxs888/kxs/backend/internal/sm"
)

func TestSM4RoundTrip(t *testing.T) {
	key, err := sm.ParseKeyHex("0123456789abcdeffedcba9876543210")
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte(`{"username":"demo","password":"x"}`)
	env, err := sm.Encrypt(key, plain)
	if err != nil {
		t.Fatal(err)
	}
	got, err := sm.Decrypt(key, env)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatalf("got %s", got)
	}
}

func TestSM4MACMismatch(t *testing.T) {
	key, _ := sm.ParseKeyHex("0123456789abcdeffedcba9876543210")
	env, err := sm.Encrypt(key, []byte("hi"))
	if err != nil {
		t.Fatal(err)
	}
	env.MAC = "00"
	if _, err := sm.Decrypt(key, env); err == nil {
		t.Fatal("want mac error")
	}
}
