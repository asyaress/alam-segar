package migrate

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func Up(ctx context.Context, db *sql.DB, directory string) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			checksum CHAR(64) NOT NULL,
			applied_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("read migration directory: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		content, readErr := os.ReadFile(filepath.Join(directory, name))
		if readErr != nil {
			return fmt.Errorf("read migration %s: %w", name, readErr)
		}
		sum := sha256.Sum256(content)
		checksum := hex.EncodeToString(sum[:])

		var existing string
		err = db.QueryRowContext(ctx, `SELECT checksum FROM schema_migrations WHERE version = ?`, name).Scan(&existing)
		if err == nil {
			if existing != checksum {
				return fmt.Errorf("migration %s pernah diterapkan tetapi checksum berubah", name)
			}
			continue
		}
		if err != sql.ErrNoRows {
			return fmt.Errorf("check migration %s: %w", name, err)
		}

		start := time.Now()
		if _, err = db.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if _, err = db.ExecContext(ctx,
			`INSERT INTO schema_migrations (version, checksum) VALUES (?, ?)`, name, checksum); err != nil {
			return fmt.Errorf("record migration %s: %w", name, err)
		}
		fmt.Printf("diterapkan %-35s %s\n", name, time.Since(start).Round(time.Millisecond))
	}

	return nil
}
