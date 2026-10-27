package httpapp

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

// RunScheduledJobs menjalankan pekerjaan harian secara idempoten: status jatuh
// tempo, pengingat tenant, penanda kontrak mendekati akhir, dan pengakuan sewa.
func (a *App) RunScheduledJobs(ctx context.Context) error {
	if err := a.cleanupExpiredSecurityData(ctx); err != nil {
		return fmt.Errorf("bersihkan data keamanan kedaluwarsa: %w", err)
	}
	today := businessToday(a.location)
	ownerRows, err := a.db.QueryContext(ctx, `SELECT id FROM users WHERE role='owner' AND status='active' AND deleted_at IS NULL ORDER BY id`)
	if err != nil {
		return err
	}
	var ownerIDs []int64
	for ownerRows.Next() {
		var id int64
		if err := ownerRows.Scan(&id); err != nil {
			ownerRows.Close()
			return err
		}
		ownerIDs = append(ownerIDs, id)
	}
	if err := ownerRows.Close(); err != nil {
		return err
	}
	for _, ownerID := range ownerIDs {
		tx, err := a.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			return err
		}
		if err = a.runOwnerScheduledJobsTx(ctx, tx, ownerID, today); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("owner %d: %w", ownerID, err)
		}
		if err = tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) cleanupExpiredSecurityData(ctx context.Context) error {
	queries := []string{
		`DELETE FROM user_sessions WHERE expires_at<=UTC_TIMESTAMP(6)`,
		`DELETE FROM password_reset_tokens WHERE expires_at<UTC_TIMESTAMP(6)-INTERVAL 7 DAY OR (used_at IS NOT NULL AND used_at<UTC_TIMESTAMP(6)-INTERVAL 7 DAY)`,
		`DELETE FROM email_verification_tokens WHERE expires_at<UTC_TIMESTAMP(6)-INTERVAL 7 DAY OR (used_at IS NOT NULL AND used_at<UTC_TIMESTAMP(6)-INTERVAL 7 DAY)`,
		`DELETE FROM login_attempts WHERE attempted_at<UTC_TIMESTAMP(6)-INTERVAL 30 DAY`,
		`DELETE FROM booking_request_attempts WHERE attempted_at<UTC_TIMESTAMP(6)-INTERVAL 30 DAY`,
		`DELETE FROM request_rate_limits WHERE updated_at<UTC_TIMESTAMP(6)-INTERVAL 2 DAY`,
	}
	for _, query := range queries {
		if _, err := a.db.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) runOwnerScheduledJobsTx(ctx context.Context, tx *sql.Tx, ownerID int64, today time.Time) error {
	reminderBefore := ownerIntSettingTx(ctx, tx, ownerID, "reminder_before_days", 3, 1, 30)
	reminderOverdue := ownerIntSettingTx(ctx, tx, ownerID, "reminder_overdue_days", 3, 1, 30)
	leaseExpiring := ownerIntSettingTx(ctx, tx, ownerID, "lease_expiring_days", 30, 7, 180)
	if _, err := tx.ExecContext(ctx, `UPDATE invoices i JOIN properties p ON p.id=i.property_id SET i.status='overdue' WHERE p.owner_id=? AND i.status IN ('issued','partially_paid') AND i.due_date<?`, ownerID, today); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE leases l JOIN properties p ON p.id=l.property_id SET l.status='expiring' WHERE p.owner_id=? AND l.status='active' AND l.end_date>? AND l.end_date<=?`, ownerID, today, today.AddDate(0, 0, leaseExpiring)); err != nil {
		return err
	}
	if err := a.recognizeRevenueThroughTx(ctx, tx, ownerID, today, ownerID); err != nil {
		return err
	}
	if err := a.runSmartLivingScheduledJobsTx(ctx, tx, ownerID, today); err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT i.id,i.invoice_number,i.tenant_id,u.email,i.due_date,i.total_amount-i.paid_amount,
			DATEDIFF(i.due_date,?)
		FROM invoices i JOIN properties p ON p.id=i.property_id JOIN users u ON u.id=i.tenant_id
		WHERE p.owner_id=? AND i.status IN ('issued','partially_paid','overdue')
		  AND DATEDIFF(i.due_date,?) IN (?,0,?)
		ORDER BY i.due_date,i.id FOR UPDATE`, today, ownerID, today, reminderBefore, -reminderOverdue)
	if err != nil {
		return err
	}
	type reminder struct {
		id, tenantID, outstanding int64
		number, email             string
		due                       time.Time
		days                      int
	}
	var reminders []reminder
	for rows.Next() {
		var v reminder
		if err := rows.Scan(&v.id, &v.number, &v.tenantID, &v.email, &v.due, &v.outstanding, &v.days); err != nil {
			rows.Close()
			return err
		}
		reminders = append(reminders, v)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, v := range reminders {
		phase := "overdue"
		title := "Tagihan melewati jatuh tempo"
		if v.days == reminderBefore {
			phase, title = fmt.Sprintf("due%d", reminderBefore), fmt.Sprintf("Tagihan jatuh tempo %d hari lagi", reminderBefore)
		} else if v.days == 0 {
			phase, title = "due", "Tagihan jatuh tempo hari ini"
		} else {
			phase, title = fmt.Sprintf("overdue%d", reminderOverdue), fmt.Sprintf("Tagihan melewati jatuh tempo %d hari", reminderOverdue)
		}
		event := fmt.Sprintf("invoice.%s.%d.%s", phase, v.id, v.due.Format("20060102"))
		body := fmt.Sprintf("Tagihan %s sebesar %s jatuh tempo %s. Pembayaran transfer tetap menunggu verifikasi pemilik.", v.number, formatRupiah(v.outstanding), v.due.In(a.location).Format("02 Jan 2006"))
		result, err := tx.ExecContext(ctx, `INSERT IGNORE INTO notification_deliveries (user_id,channel,event_key,recipient,subject,body) VALUES (?,'email',?,?,?,?)`, v.tenantID, event, v.email, title, body)
		if err != nil {
			return err
		}
		inserted, _ := result.RowsAffected()
		if inserted == 1 {
			if err := createNotificationTx(ctx, tx, v.tenantID, event, title, body, "/penghuni/tagihan"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *App) runSmartLivingScheduledJobsTx(ctx context.Context, tx *sql.Tx, ownerID int64, today time.Time) error {
	if _, err := tx.ExecContext(ctx, `UPDATE visitor_passes v JOIN properties p ON p.id=v.property_id SET v.status='expired',v.owner_note=COALESCE(v.owner_note,'Jadwal kunjungan telah lewat') WHERE p.owner_id=? AND v.status IN ('pending','approved') AND v.visit_date<?`, ownerID, today); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE facility_reservations fr JOIN properties p ON p.id=fr.property_id SET fr.status='completed',fr.owner_note=COALESCE(fr.owner_note,'Selesai otomatis setelah jadwal') WHERE p.owner_id=? AND fr.status='approved' AND fr.reservation_date<?`, ownerID, today); err != nil {
		return err
	}
	type smartReminder struct {
		userID                         int64
		event, email, title, body, url string
	}
	var reminders []smartReminder
	rows, err := tx.QueryContext(ctx, `SELECT v.id,v.tenant_id,u.email,v.pass_number,v.visitor_name,DATE_FORMAT(v.starts_at,'%H:%i'),COALESCE(v.access_code,'') FROM visitor_passes v JOIN users u ON u.id=v.tenant_id JOIN properties p ON p.id=v.property_id WHERE p.owner_id=? AND v.status='approved' AND v.visit_date=? FOR UPDATE`, ownerID, today)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int64
		var v smartReminder
		var number, name, start, code string
		if err := rows.Scan(&id, &v.userID, &v.email, &number, &name, &start, &code); err != nil {
			rows.Close()
			return err
		}
		v.event = fmt.Sprintf("smart.visitor.today.%d.%s", id, today.Format("20060102"))
		v.title = "Pengingat kunjungan hari ini"
		v.body = fmt.Sprintf("%s (%s) dijadwalkan pukul %s. Kode akses: %s.", name, number, start, code)
		v.url = "/penghuni/layanan"
		reminders = append(reminders, v)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	rows, err = tx.QueryContext(ctx, `SELECT fr.id,fr.tenant_id,u.email,fr.reservation_number,sf.name,DATE_FORMAT(fr.starts_at,'%H:%i') FROM facility_reservations fr JOIN shared_facilities sf ON sf.id=fr.facility_id JOIN users u ON u.id=fr.tenant_id JOIN properties p ON p.id=fr.property_id WHERE p.owner_id=? AND fr.status='approved' AND fr.reservation_date=? FOR UPDATE`, ownerID, today)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int64
		var v smartReminder
		var number, name, start string
		if err := rows.Scan(&id, &v.userID, &v.email, &number, &name, &start); err != nil {
			rows.Close()
			return err
		}
		v.event = fmt.Sprintf("smart.reservation.today.%d.%s", id, today.Format("20060102"))
		v.title = "Pengingat reservasi hari ini"
		v.body = fmt.Sprintf("Reservasi %s untuk %s dimulai pukul %s.", number, name, start)
		v.url = "/penghuni/layanan"
		reminders = append(reminders, v)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	rows, err = tx.QueryContext(ctx, `SELECT pd.id,pd.tenant_id,u.email,pd.parcel_number,pd.courier,DATEDIFF(?,DATE(pd.received_at)) FROM parcel_deliveries pd JOIN users u ON u.id=pd.tenant_id JOIN properties p ON p.id=pd.property_id WHERE p.owner_id=? AND pd.status IN ('arrived','notified') AND DATEDIFF(?,DATE(pd.received_at)) IN (1,3) FOR UPDATE`, today, ownerID, today)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id, days int64
		var v smartReminder
		var number, courier string
		if err := rows.Scan(&id, &v.userID, &v.email, &number, &courier, &days); err != nil {
			rows.Close()
			return err
		}
		v.event = fmt.Sprintf("smart.parcel.reminder.%d.day%d", id, days)
		v.title = "Paket masih menunggu diambil"
		v.body = fmt.Sprintf("Paket %s dari %s sudah menunggu %d hari.", number, courier, days)
		v.url = "/penghuni/layanan#paket-saya"
		reminders = append(reminders, v)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, v := range reminders {
		result, err := tx.ExecContext(ctx, `INSERT IGNORE INTO notification_deliveries(user_id,channel,event_key,recipient,subject,body) VALUES(?,'email',?,?,?,?)`, v.userID, v.event, v.email, v.title, v.body)
		if err != nil {
			return err
		}
		inserted, _ := result.RowsAffected()
		if inserted == 1 {
			if err := createNotificationTx(ctx, tx, v.userID, v.event, v.title, v.body, v.url); err != nil {
				return err
			}
		}
	}
	return nil
}

func ownerIntSettingTx(ctx context.Context, tx *sql.Tx, ownerID int64, name string, fallback, minimum, maximum int) int {
	var value string
	if err := tx.QueryRowContext(ctx, `SELECT setting_value FROM app_settings WHERE setting_key=?`, ownerSettingKey(ownerID, name)).Scan(&value); err != nil {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < minimum || parsed > maximum {
		return fallback
	}
	return parsed
}
