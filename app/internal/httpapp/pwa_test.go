package httpapp

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"alam-segar/kos-management/internal/config"
)

func TestEmbeddedTemplatesParse(t *testing.T) {
	app, err := New(config.Config{}, nil)
	if err != nil {
		t.Fatalf("template aplikasi gagal diparse: %v", err)
	}
	for _, name := range []string{"public_home", "owner_dashboard", "owner_finance", "owner_smart", "owner_documents", "tenant_dashboard", "tenant_room", "tenant_smart", "tenant_documents", "document_detail", "document_verify"} {
		if app.templates.Lookup(name) == nil {
			t.Fatalf("template %s tidak ditemukan", name)
		}
	}
}

func TestPWAServiceWorkerHeadersAndPathRestriction(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "service-worker.js"), []byte("self.addEventListener('fetch',()=>{});"), 0600); err != nil {
		t.Fatal(err)
	}
	app := &App{cfg: config.Config{PWAAssetsRoot: root}}
	handler := app.pwaFile("service-worker.js", "application/javascript; charset=utf-8")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/service-worker.js", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status service worker = %d", recorder.Code)
	}
	if got := recorder.Header().Get("Service-Worker-Allowed"); got != "/" {
		t.Fatalf("Service-Worker-Allowed = %q", got)
	}
	if got := recorder.Header().Get("Cache-Control"); got != "no-cache" {
		t.Fatalf("Cache-Control = %q", got)
	}
	if !strings.Contains(recorder.Body.String(), "fetch") {
		t.Fatal("isi service worker tidak dikirim")
	}

	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/pwa/icon.svg", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("path yang tidak sesuai seharusnya 404, mendapat %d", recorder.Code)
	}
}
