package config

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Environment       string
	Address           string
	BaseURL           string
	SecureCookie      bool
	HashKey           string
	TrustedProxyCIDRs []string
	ProjectRoot       string
	DatabaseDSN       string
	AdminAssetsRoot   string
	PublicAssetsRoot  string
	PWAAssetsRoot     string
	PrivateFilesRoot  string
	PublicFilesRoot   string
	MailDriver        string
	SMTPHost          string
	SMTPPort          string
	SMTPUser          string
	SMTPPassword      string
	SMTPFrom          string
}

func Load() (Config, error) {
	projectRoot, err := resolveProjectRoot()
	if err != nil {
		return Config{}, err
	}

	loadDotEnv(filepath.Join(projectRoot, "app", ".env"))

	if value := strings.TrimSpace(os.Getenv("APP_PROJECT_ROOT")); value != "" {
		projectRoot, err = filepath.Abs(value)
		if err != nil {
			return Config{}, fmt.Errorf("resolve APP_PROJECT_ROOT: %w", err)
		}
	}

	secureCookie, err := strconv.ParseBool(envOr("APP_SECURE_COOKIE", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("APP_SECURE_COOKIE harus true atau false: %w", err)
	}

	cfg := Config{
		Environment:       envOr("APP_ENV", "development"),
		Address:           envOr("APP_ADDR", "127.0.0.1:8080"),
		BaseURL:           envOr("APP_BASE_URL", "http://127.0.0.1:8080"),
		SecureCookie:      secureCookie,
		HashKey:           envOr("APP_HASH_KEY", "local-development-only-change-me"),
		TrustedProxyCIDRs: splitCSV(os.Getenv("APP_TRUSTED_PROXY_CIDRS")),
		ProjectRoot:       projectRoot,
		DatabaseDSN: envOr("DATABASE_DSN",
			"root@tcp(127.0.0.1:3306)/alam-segar?parseTime=true&loc=UTC&charset=utf8mb4&collation=utf8mb4_general_ci&multiStatements=true"),
		AdminAssetsRoot: filepath.Join(projectRoot, "package", "assets"),
		PublicAssetsRoot: filepath.Join(projectRoot,
			"hotel-booking-boostrap-5-template-seaside-2025-08-25-10-30-06-utc",
			"Seaside", "Seaside HTML"),
		PWAAssetsRoot:    filepath.Join(projectRoot, "app", "static", "pwa"),
		PrivateFilesRoot: filepath.Join(projectRoot, "app", "storage", "private"),
		PublicFilesRoot:  filepath.Join(projectRoot, "app", "storage", "public"),
		MailDriver:       strings.ToLower(envOr("MAIL_DRIVER", "log")),
		SMTPHost:         envOr("SMTP_HOST", ""),
		SMTPPort:         envOr("SMTP_PORT", "587"),
		SMTPUser:         envOr("SMTP_USER", ""),
		SMTPPassword:     envOr("SMTP_PASSWORD", ""),
		SMTPFrom:         envOr("SMTP_FROM", "noreply@alamsegar.local"),
	}
	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	for _, directory := range []string{cfg.PWAAssetsRoot, cfg.PrivateFilesRoot, cfg.PublicFilesRoot} {
		if err := os.MkdirAll(directory, 0700); err != nil {
			return Config{}, fmt.Errorf("buat direktori penyimpanan %s: %w", directory, err)
		}
	}
	for label, path := range map[string]string{
		"aset admin Travl":    cfg.AdminAssetsRoot,
		"aset publik Seaside": cfg.PublicAssetsRoot,
		"aset PWA aplikasi":   cfg.PWAAssetsRoot,
		"penyimpanan privat":  cfg.PrivateFilesRoot,
		"penyimpanan publik":  cfg.PublicFilesRoot,
	} {
		info, statErr := os.Stat(path)
		if statErr != nil || !info.IsDir() {
			return Config{}, fmt.Errorf("%s tidak ditemukan di %s", label, path)
		}
	}

	return cfg, nil
}

func validate(cfg Config) error {
	if cfg.MailDriver != "log" && cfg.MailDriver != "smtp" {
		return errors.New("MAIL_DRIVER harus log atau smtp")
	}
	if cfg.MailDriver == "smtp" && (cfg.SMTPHost == "" || cfg.SMTPFrom == "") {
		return errors.New("SMTP_HOST dan SMTP_FROM wajib diisi saat MAIL_DRIVER=smtp")
	}
	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil || baseURL.Host == "" || (baseURL.Scheme != "http" && baseURL.Scheme != "https") || baseURL.RawQuery != "" || baseURL.Fragment != "" {
		return errors.New("APP_BASE_URL harus URL http/https yang valid tanpa query atau fragmen")
	}
	if cfg.Environment != "development" {
		if baseURL.Scheme != "https" {
			return errors.New("APP_BASE_URL wajib HTTPS di luar mode development")
		}
		if !cfg.SecureCookie {
			return errors.New("APP_SECURE_COOKIE wajib true di luar mode development")
		}
		if cfg.HashKey == "local-development-only-change-me" || len(cfg.HashKey) < 32 {
			return errors.New("APP_HASH_KEY wajib unik dan minimal 32 karakter di luar mode development")
		}
	}
	for _, cidr := range cfg.TrustedProxyCIDRs {
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			return fmt.Errorf("APP_TRUSTED_PROXY_CIDRS berisi CIDR tidak valid %q: %w", cidr, err)
		}
	}
	return nil
}

func splitCSV(value string) []string {
	var values []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			values = append(values, item)
		}
	}
	return values
}

func resolveProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("read working directory: %w", err)
	}

	current := wd
	for range 6 {
		if _, statErr := os.Stat(filepath.Join(current, "KONTEKS_LENGKAP_CODEX_KOS_MANAGEMENT.md")); statErr == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", errors.New("root proyek tidak ditemukan; set APP_PROJECT_ROOT")
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) == "" {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
