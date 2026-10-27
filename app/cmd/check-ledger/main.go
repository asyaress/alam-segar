package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"alam-segar/kos-management/internal/config"
	"alam-segar/kos-management/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := database.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var entries, debit, credit, unbalanced int64
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT je.id),COALESCE(SUM(jl.debit),0),COALESCE(SUM(jl.credit),0)
		FROM journal_entries je JOIN journal_lines jl ON jl.journal_entry_id=je.id
		WHERE je.status='posted'`).Scan(&entries, &debit, &credit); err != nil {
		log.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM (
			SELECT je.id FROM journal_entries je JOIN journal_lines jl ON jl.journal_entry_id=je.id
			WHERE je.status='posted' GROUP BY je.id HAVING SUM(jl.debit)<>SUM(jl.credit)
		) tidak_seimbang`).Scan(&unbalanced); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("jurnal_posted=%d total_debit=%d total_kredit=%d jurnal_tidak_seimbang=%d\n", entries, debit, credit, unbalanced)
	if debit != credit || unbalanced != 0 {
		log.Fatal("ledger tidak seimbang")
	}
}
