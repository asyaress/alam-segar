package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/mail"
	"os"
	"strings"
	"time"

	"alam-segar/kos-management/internal/auth"
	"alam-segar/kos-management/internal/config"
	"alam-segar/kos-management/internal/database"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	name := strings.TrimSpace(os.Getenv("OWNER_NAME"))
	email := strings.ToLower(strings.TrimSpace(os.Getenv("OWNER_EMAIL")))
	password := os.Getenv("OWNER_PASSWORD")
	if name == "" || email == "" || password == "" {
		log.Fatal("OWNER_NAME, OWNER_EMAIL, dan OWNER_PASSWORD wajib diisi")
	}
	address, err := mail.ParseAddress(email)
	if err != nil || !strings.EqualFold(address.Address, email) {
		log.Fatal("OWNER_EMAIL tidak valid")
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		if cfg.Environment != "development" || !strings.EqualFold(os.Getenv("ALLOW_WEAK_DEV_PASSWORD"), "true") {
			log.Fatalf("kata sandi ditolak: %v", err)
		}
		weakHash, hashErr := bcrypt.GenerateFromPassword([]byte(password), 12)
		if hashErr != nil {
			log.Fatal(hashErr)
		}
		passwordHash = string(weakHash)
		log.Print("peringatan: pengecualian kata sandi lemah hanya digunakan untuk akun development lokal")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	db, err := database.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	var id int64
	var existingRole string
	err = tx.QueryRowContext(ctx, `SELECT id, role FROM users WHERE email = ? FOR UPDATE`, email).Scan(&id, &existingRole)
	action := "owner.account_created"
	switch {
	case errors.Is(err, sql.ErrNoRows):
		result, insertErr := tx.ExecContext(ctx, `
			INSERT INTO users
				(role, status, name, email, password_hash, email_verified_at, approved_at)
			VALUES ('owner', 'active', ?, ?, ?, UTC_TIMESTAMP(6), UTC_TIMESTAMP(6))`, name, email, passwordHash)
		if insertErr != nil {
			log.Fatal(insertErr)
		}
		id, _ = result.LastInsertId()
	case err != nil:
		log.Fatal(err)
	case existingRole != "owner":
		log.Fatalf("email %s sudah digunakan akun dengan role %s", email, existingRole)
	default:
		action = "owner.account_updated"
		if _, err := tx.ExecContext(ctx, `
			UPDATE users
			SET status='active', name=?, password_hash=?, email_verified_at=UTC_TIMESTAMP(6),
				approved_at=UTC_TIMESTAMP(6), deleted_at=NULL
			WHERE id=?`, name, passwordHash, id); err != nil {
			log.Fatal(err)
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM user_sessions WHERE user_id = ?`, id); err != nil {
		log.Fatal(err)
	}
	after, _ := json.Marshal(map[string]any{"email": email, "name": name, "role": "owner", "status": "active"})
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO audit_logs (actor_user_id, action, entity_type, entity_id, after_json)
		VALUES (NULL, ?, 'user', ?, ?)`, action, fmt.Sprint(id), string(after)); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
	log.Printf("akun owner siap: %s (id=%d)", email, id)
}
