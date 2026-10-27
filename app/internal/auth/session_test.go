package auth

import (
	"net/http/httptest"
	"testing"
)

func TestClientIPOnlyTrustsForwardedHeaderFromConfiguredProxy(t *testing.T) {
	store := NewSessionStore(nil, false, "0123456789abcdef0123456789abcdef", []string{"127.0.0.1/32"})
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "127.0.0.1:54321"
	r.Header.Set("X-Forwarded-For", "203.0.113.8, 127.0.0.1")
	if got := store.clientIP(r); got != "203.0.113.8" {
		t.Fatalf("client IP = %q", got)
	}

	r.RemoteAddr = "198.51.100.7:54321"
	r.Header.Set("X-Forwarded-For", "203.0.113.8")
	if got := store.clientIP(r); got != "198.51.100.7" {
		t.Fatalf("header dari peer tak terpercaya dipakai: %q", got)
	}
}

func TestClientIPRejectsMalformedForwardedChain(t *testing.T) {
	store := NewSessionStore(nil, false, "0123456789abcdef0123456789abcdef", []string{"127.0.0.1/32"})
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "127.0.0.1:54321"
	r.Header.Set("X-Forwarded-For", "203.0.113.8, bukan-ip")
	if got := store.clientIP(r); got != "127.0.0.1" {
		t.Fatalf("rantai rusak tidak fail-closed: %q", got)
	}
}
