package main

import (
	"context"
	"log"
	"time"

	"alam-segar/kos-management/internal/config"
	"alam-segar/kos-management/internal/database"
	"alam-segar/kos-management/internal/httpapp"
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
	app, err := httpapp.New(cfg, db)
	if err != nil {
		log.Fatal(err)
	}
	if err := app.RunScheduledJobs(ctx); err != nil {
		log.Fatal(err)
	}
	log.Print("pekerjaan terjadwal selesai")
}
