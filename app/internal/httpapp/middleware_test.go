package httpapp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublicRateRulesOnlyProtectSensitivePOST(t *testing.T) {
	for path, expected := range map[string]string{
		"/masuk": "login", "/lupa-password": "forgot_password", "/daftar": "registration",
		"/booking": "booking", "/booking/pembayaran": "booking_payment",
	} {
		rule, ok := publicRateRule(httptest.NewRequest(http.MethodPost, path, nil))
		if !ok || rule.action != expected || rule.limit < 1 || rule.window <= 0 {
			t.Fatalf("rule %s tidak valid: %#v, %v", path, rule, ok)
		}
	}
	if _, ok := publicRateRule(httptest.NewRequest(http.MethodGet, "/masuk", nil)); ok {
		t.Fatal("GET login tidak boleh mengonsumsi rate limit")
	}
	if _, ok := publicRateRule(httptest.NewRequest(http.MethodPost, "/pemilik", nil)); ok {
		t.Fatal("route internal tidak boleh memakai limiter publik")
	}
}

func TestCompressionAndSessionFreePaths(t *testing.T) {
	if !acceptsGzip("br, gzip") || acceptsGzip("gzip;q=0") {
		t.Fatal("negosiasi gzip tidak benar")
	}
	for _, path := range []string{"/assets/admin/css/style.css", "/service-worker.js", "/pemilik"} {
		if !compressiblePath(path) {
			t.Fatalf("%s seharusnya dapat dikompresi", path)
		}
	}
	for _, path := range []string{"/uploads/rooms/a.jpg", "/penghuni/file/payment/1", "/assets/admin/images/a.png"} {
		if compressiblePath(path) {
			t.Fatalf("%s tidak boleh dikompresi ulang", path)
		}
	}
	for _, path := range []string{"/assets/admin/js/app.js", "/uploads/rooms/a.jpg", "/pwa/offline.html", "/kesehatan"} {
		if !sessionFreePath(path) {
			t.Fatalf("%s seharusnya bebas query session", path)
		}
	}
	if sessionFreePath("/pemilik") {
		t.Fatal("dashboard wajib memuat session")
	}
}

func TestSecurityHeadersUseNonceInsteadOfUnsafeInlineScript(t *testing.T) {
	app := &App{}
	handler := app.securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secured, ok := w.(*securityResponseWriter)
		if !ok || secured.nonce == "" {
			t.Fatal("nonce tidak diteruskan ke renderer")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	csp := recorder.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "script-src 'self' 'nonce-") || strings.Contains(csp, "script-src 'self' 'unsafe-inline'") {
		t.Fatalf("CSP script tidak ketat: %s", csp)
	}
	if recorder.Header().Get("X-Frame-Options") != "DENY" || recorder.Header().Get("Cross-Origin-Opener-Policy") != "same-origin" {
		t.Fatal("header isolasi tidak lengkap")
	}
}
