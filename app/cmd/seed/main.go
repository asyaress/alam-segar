package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"alam-segar/kos-management/internal/auth"
	"alam-segar/kos-management/internal/config"
	"alam-segar/kos-management/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if cfg.Environment != "development" {
		log.Fatal("seed akun lokal hanya boleh dijalankan pada APP_ENV=development")
	}

	name := envOr("DEV_OWNER_NAME", "Pemilik Alam Segar")
	email := strings.ToLower(envOr("DEV_OWNER_EMAIL", "pemilik@alamsegar.local"))
	password := envOr("DEV_OWNER_PASSWORD", "Ganti-Sebelum-Deploy-123!")
	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatalf("kata sandi seed tidak valid: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	db, err := database.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var id int64
	err = db.QueryRowContext(ctx, `SELECT id FROM users WHERE email = ? AND deleted_at IS NULL`, email).Scan(&id)
	switch {
	case err == nil:
		log.Printf("akun owner development sudah ada: %s", email)
	case err == sql.ErrNoRows:
		result, insertErr := db.ExecContext(ctx, `
			INSERT INTO users
				(role, status, name, email, password_hash, email_verified_at, approved_at)
			VALUES ('owner', 'active', ?, ?, ?, UTC_TIMESTAMP(6), UTC_TIMESTAMP(6))`, name, email, hash)
		if insertErr != nil {
			log.Fatal(insertErr)
		}
		id, _ = result.LastInsertId()
		log.Printf("akun owner development dibuat: %s (id=%d)", email, id)
	default:
		log.Fatal(err)
	}

	log.Println("kata sandi mengikuti DEV_OWNER_PASSWORD; jangan gunakan akun seed ini untuk data asli")

	seedDemo, _ := strconv.ParseBool(envOr("DEV_SEED_DEMO_DATA", "false"))
	if seedDemo {
		if err := seedDemoData(ctx, db, id); err != nil {
			log.Fatal(err)
		}
		log.Println("data properti dan kamar DEMO tersedia; hapus/ganti sebelum memasukkan data nyata")
	}
}

func seedDemoData(ctx context.Context, db *sql.DB, ownerID int64) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO properties (owner_id, code, slug, name, address, description, contact_phone, is_published, status)
		VALUES (?, 'DEMO', 'alam-segar-demo', 'Alam Segar',
			'Alamat demo — ganti melalui data properti sebelum penggunaan nyata',
			'Kos nyaman, rapi, dan semua urusan bisa di-track dari HP.',
			'080000000000', 1, 'active')
		ON DUPLICATE KEY UPDATE owner_id = VALUES(owner_id)`, ownerID)
	if err != nil {
		return err
	}
	var propertyID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM properties WHERE owner_id = ? AND code = 'DEMO'`, ownerID).Scan(&propertyID); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO room_types (property_id, name, description, base_monthly_rate, base_annual_rate)
		VALUES (?, 'Standar Demo', 'Tipe kamar untuk pengujian lokal', 1200000, 12000000)
		ON DUPLICATE KEY UPDATE description = VALUES(description)`, propertyID)
	if err != nil {
		return err
	}
	var typeID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM room_types WHERE property_id = ? AND name = 'Standar Demo'`, propertyID).Scan(&typeID); err != nil {
		return err
	}
	for _, number := range []string{"01", "02"} {
		slug := "kamar-demo-" + number
		_, err = tx.ExecContext(ctx, `
			INSERT INTO rooms
				(property_id, room_type_id, number, slug, floor_name, monthly_rate, annual_rate, capacity,
				 condition_status, is_published, description)
			VALUES (?, ?, ?, ?, 'Lantai Demo', 1200000, 12000000, 1, 'ready', 1,
				'Kamar nyaman buat satu orang. Clean, private, dan praktis.')
			ON DUPLICATE KEY UPDATE room_type_id = VALUES(room_type_id)`, propertyID, typeID, number, slug)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
