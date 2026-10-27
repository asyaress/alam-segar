package config

import "testing"

func TestValidateRejectsUnsafeLiveConfiguration(t *testing.T) {
	base := Config{Environment: "live", BaseURL: "https://kos.example", SecureCookie: true, HashKey: "0123456789abcdef0123456789abcdef", MailDriver: "log"}
	if err := validate(base); err != nil {
		t.Fatalf("konfigurasi aman ditolak: %v", err)
	}

	tests := []Config{
		{Environment: "live", BaseURL: "http://kos.example", SecureCookie: true, HashKey: base.HashKey, MailDriver: "log"},
		{Environment: "live", BaseURL: base.BaseURL, SecureCookie: false, HashKey: base.HashKey, MailDriver: "log"},
		{Environment: "live", BaseURL: base.BaseURL, SecureCookie: true, HashKey: "terlalu-pendek", MailDriver: "log"},
	}
	for i, cfg := range tests {
		if err := validate(cfg); err == nil {
			t.Fatalf("konfigurasi tidak aman #%d seharusnya ditolak", i+1)
		}
	}
}

func TestValidateAllowsLocalHTTP(t *testing.T) {
	cfg := Config{Environment: "development", BaseURL: "http://127.0.0.1:8080", HashKey: "local-development-only-change-me", MailDriver: "log"}
	if err := validate(cfg); err != nil {
		t.Fatalf("konfigurasi lokal ditolak: %v", err)
	}
}

func TestValidateTrustedProxyCIDR(t *testing.T) {
	cfg := Config{Environment: "development", BaseURL: "http://127.0.0.1:8080", HashKey: "local-development-only-change-me", MailDriver: "log", TrustedProxyCIDRs: []string{"127.0.0.1/32", "::1/128"}}
	if err := validate(cfg); err != nil {
		t.Fatalf("CIDR proxy valid ditolak: %v", err)
	}
	cfg.TrustedProxyCIDRs = []string{"semua-proxy"}
	if err := validate(cfg); err == nil {
		t.Fatal("CIDR proxy tidak valid seharusnya ditolak")
	}
}
