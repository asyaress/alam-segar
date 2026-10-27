package httpapp

import (
	"compress/gzip"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	maxRequestBytes = 32 << 20
	requestTimeout  = 25 * time.Second
)

type securityResponseWriter struct {
	http.ResponseWriter
	nonce string
}

func sessionFreePath(path string) bool {
	return strings.HasPrefix(path, "/assets/") ||
		strings.HasPrefix(path, "/uploads/") ||
		strings.HasPrefix(path, "/pwa/") ||
		path == "/manifest.webmanifest" ||
		path == "/service-worker.js" ||
		path == "/kesehatan"
}

func (a *App) recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("panic %s %s: %v", r.Method, r.URL.Path, recovered)
				http.Error(w, "Terjadi kesalahan pada server.", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (a *App) requestDeadline(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *App) requestLimits(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil && r.Method != http.MethodGet && r.Method != http.MethodHead {
			r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)
		}
		next.ServeHTTP(w, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer   *gzip.Writer
	disabled bool
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	if status == http.StatusNoContent || status == http.StatusNotModified || status < 200 {
		w.disabled = true
		w.Header().Del("Content-Encoding")
	} else {
		w.Header().Del("Content-Length")
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *gzipResponseWriter) Write(body []byte) (int, error) {
	if w.disabled {
		return w.ResponseWriter.Write(body)
	}
	w.Header().Del("Content-Length")
	return w.writer.Write(body)
}

func (w *gzipResponseWriter) Flush() {
	_ = w.writer.Flush()
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func acceptsGzip(value string) bool {
	for _, part := range strings.Split(value, ",") {
		fields := strings.Split(strings.TrimSpace(part), ";")
		if strings.EqualFold(strings.TrimSpace(fields[0]), "gzip") {
			for _, parameter := range fields[1:] {
				if strings.TrimSpace(parameter) == "q=0" {
					return false
				}
			}
			return true
		}
	}
	return false
}

func compressiblePath(path string) bool {
	if strings.HasPrefix(path, "/uploads/") || strings.HasPrefix(path, "/pemilik/file/") || strings.HasPrefix(path, "/penghuni/file/") {
		return false
	}
	if strings.HasPrefix(path, "/assets/") {
		switch strings.ToLower(filepath.Ext(path)) {
		case ".css", ".js", ".svg", ".txt", ".xml":
			return true
		default:
			return false
		}
	}
	return true
}

func (a *App) gzipResponses(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead || r.Header.Get("Range") != "" || !acceptsGzip(r.Header.Get("Accept-Encoding")) || !compressiblePath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Add("Vary", "Accept-Encoding")
		w.Header().Set("Content-Encoding", "gzip")
		writer := gzip.NewWriter(w)
		defer writer.Close()
		next.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, writer: writer}, r)
	})
}

type rateLimitRule struct {
	action string
	limit  int
	window time.Duration
}

var publicRateRules = map[string]rateLimitRule{
	"/masuk":              {action: "login", limit: 30, window: 15 * time.Minute},
	"/lupa-password":      {action: "forgot_password", limit: 5, window: time.Hour},
	"/reset-password":     {action: "reset_password", limit: 10, window: time.Hour},
	"/daftar":             {action: "registration", limit: 5, window: time.Hour},
	"/booking":            {action: "booking", limit: 20, window: time.Hour},
	"/booking/pembayaran": {action: "booking_payment", limit: 20, window: time.Hour},
}

func publicRateRule(r *http.Request) (rateLimitRule, bool) {
	if r.Method != http.MethodPost {
		return rateLimitRule{}, false
	}
	rule, ok := publicRateRules[r.URL.Path]
	return rule, ok
}

func (a *App) publicRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rule, protected := publicRateRule(r)
		if !protected {
			next.ServeHTTP(w, r)
			return
		}
		allowed, retryAfter, err := a.consumeRateLimit(r.Context(), rule, a.sessions.HashIP(r))
		if err != nil {
			a.serverError(w, r, err)
			return
		}
		if !allowed {
			w.Header().Set("Retry-After", strconv.FormatInt(maxInt64(1, int64(retryAfter.Seconds())), 10))
			http.Error(w, "Terlalu banyak percobaan. Tunggu sebentar lalu coba lagi.", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *App) consumeRateLimit(ctx context.Context, rule rateLimitRule, ipHash string) (bool, time.Duration, error) {
	tx, err := a.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return false, 0, err
	}
	defer tx.Rollback()
	now := time.Now().UTC()
	if _, err = tx.ExecContext(ctx, `INSERT IGNORE INTO request_rate_limits (action,ip_hash,window_started_at,attempts,updated_at) VALUES (?,?,?,0,?)`, rule.action, ipHash, now, now); err != nil {
		return false, 0, err
	}
	var started time.Time
	var attempts int
	if err = tx.QueryRowContext(ctx, `SELECT window_started_at,attempts FROM request_rate_limits WHERE action=? AND ip_hash=? FOR UPDATE`, rule.action, ipHash).Scan(&started, &attempts); err != nil {
		return false, 0, err
	}
	if elapsed := now.Sub(started); elapsed >= rule.window || elapsed < 0 {
		started = now
		attempts = 0
	}
	if attempts >= rule.limit {
		if err = tx.Commit(); err != nil {
			return false, 0, err
		}
		return false, rule.window - now.Sub(started), nil
	}
	result, err := tx.ExecContext(ctx, `UPDATE request_rate_limits SET window_started_at=?,attempts=?,updated_at=? WHERE action=? AND ip_hash=?`, started, attempts+1, now, rule.action, ipHash)
	if err != nil {
		return false, 0, err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return false, 0, errors.New("rate limit tidak dapat diperbarui")
	}
	if err = tx.Commit(); err != nil {
		return false, 0, fmt.Errorf("commit rate limit: %w", err)
	}
	return true, 0, nil
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
