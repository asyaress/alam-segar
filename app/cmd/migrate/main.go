package main

import (
	"context"
	"log"
	"path/filepath"
	"time"

	"alam-segar/kos-management/internal/config"
	"alam-segar/kos-management/internal/database"
	"alam-segar/kos-management/internal/migrate"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	db, err := database.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := migrate.Up(ctx, db, filepath.Join(cfg.ProjectRoot, "app", "migrations")); err != nil {
		log.Fatal(err)
	}
	log.Println("migrasi database selesai")
}
