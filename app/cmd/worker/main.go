package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"alam-segar/kos-management/internal/config"
	"alam-segar/kos-management/internal/database"
)

type delivery struct {
	id                       int64
	recipient, subject, body string
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	db, err := database.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	processed := 0
	for processed < 100 {
		item, ok, err := claim(ctx, db)
		if err != nil {
			log.Fatal(err)
		}
		if !ok {
			break
		}
		err = send(cfg, item)
		if err != nil {
			_, _ = db.ExecContext(ctx, `UPDATE notification_deliveries SET status='failed',last_error=?,claimed_at=NULL WHERE id=? AND status='processing'`, truncate(err.Error(), 1000), item.id)
			log.Printf("email id=%d gagal: %v", item.id, err)
		} else {
			_, err = db.ExecContext(ctx, `UPDATE notification_deliveries SET status='sent',sent_at=UTC_TIMESTAMP(6),last_error=NULL,claimed_at=NULL WHERE id=? AND status='processing'`, item.id)
			if err != nil {
				log.Fatal(err)
			}
		}
		processed++
	}
	log.Printf("worker selesai; %d email diproses", processed)
}

func claim(ctx context.Context, db *sql.DB) (delivery, bool, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return delivery{}, false, err
	}
	defer tx.Rollback()
	_, _ = tx.ExecContext(ctx, `UPDATE notification_deliveries SET status='failed',claimed_at=NULL,last_error='Klaim worker kedaluwarsa' WHERE status='processing' AND claimed_at<UTC_TIMESTAMP(6)-INTERVAL 15 MINUTE`)
	var item delivery
	err = tx.QueryRowContext(ctx, `SELECT id,recipient,subject,body FROM notification_deliveries WHERE status IN ('queued','failed') AND attempts<5 ORDER BY created_at,id LIMIT 1 FOR UPDATE`).Scan(&item.id, &item.recipient, &item.subject, &item.body)
	if err == sql.ErrNoRows {
		return delivery{}, false, nil
	}
	if err != nil {
		return delivery{}, false, err
	}
	result, err := tx.ExecContext(ctx, `UPDATE notification_deliveries SET status='processing',attempts=attempts+1,claimed_at=UTC_TIMESTAMP(6) WHERE id=? AND status IN ('queued','failed')`, item.id)
	if err != nil {
		return delivery{}, false, err
	}
	affected, _ := result.RowsAffected()
	if affected != 1 {
		return delivery{}, false, nil
	}
	if err := tx.Commit(); err != nil {
		return delivery{}, false, err
	}
	return item, true, nil
}

func send(cfg config.Config, item delivery) error {
	if cfg.MailDriver == "log" {
		log.Printf("MAIL_DRIVER=log: email id=%d untuk %s, subjek=%q dianggap terkirim (isi tidak dicetak)", item.id, item.recipient, item.subject)
		return nil
	}
	recipient, err := mail.ParseAddress(item.recipient)
	if err != nil {
		return err
	}
	from, err := mail.ParseAddress(cfg.SMTPFrom)
	if err != nil {
		return err
	}
	subject := strings.NewReplacer("\r", " ", "\n", " ").Replace(item.subject)
	body := strings.ReplaceAll(strings.ReplaceAll(item.body, "\r\n", "\n"), "\n", "\r\n")
	message := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s", cfg.SMTPFrom, item.recipient, subject, body))
	var auth smtp.Auth
	if cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)
	}
	return smtp.SendMail(net.JoinHostPort(cfg.SMTPHost, cfg.SMTPPort), auth, from.Address, []string{recipient.Address}, message)
}

func truncate(value string, length int) string {
	if len(value) <= length {
		return value
	}
	return value[:length]
}
