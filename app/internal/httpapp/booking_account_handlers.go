package httpapp

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"alam-segar/kos-management/internal/auth"
)

// provisionBookingTenant mengubah pemohon publik menjadi akun tenant hanya
// setelah uang booking dikonfirmasi pemilik. Tautan aktivasi dikirim ke email;
// akun belum dapat masuk sebelum pemilik email memilih kata sandi.
func (a *App) provisionBookingTenant(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	bookingID, err := parseID(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()

	var linkedUser sql.NullInt64
	var name, email, phone, bookingStatus string
	err = tx.QueryRowContext(r.Context(), `
		SELECT b.applicant_user_id,b.applicant_name,LOWER(b.applicant_email),b.applicant_phone,b.status
		FROM booking_requests b JOIN properties p ON p.id=b.property_id
		WHERE b.id=? AND p.owner_id=? FOR UPDATE`, bookingID, session.User.ID).
		Scan(&linkedUser, &name, &email, &phone, &bookingStatus)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if bookingStatus != "confirmed" {
		http.Error(w, "Akun baru dapat dibuat setelah dana booking diverifikasi.", http.StatusConflict)
		return
	}

	userID := linkedUser.Int64
	var role, userStatus string
	var verified sql.NullTime
	if userID > 0 {
		err = tx.QueryRowContext(r.Context(), `SELECT role,status,email_verified_at FROM users WHERE id=? AND deleted_at IS NULL FOR UPDATE`, userID).
			Scan(&role, &userStatus, &verified)
	} else {
		err = tx.QueryRowContext(r.Context(), `SELECT id,role,status,email_verified_at FROM users WHERE email=? AND deleted_at IS NULL FOR UPDATE`, email).
			Scan(&userID, &role, &userStatus, &verified)
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		a.serverError(w, r, err)
		return
	}
	if err == nil && role != "tenant" {
		http.Error(w, "Email pemohon sudah dipakai oleh akun selain penghuni.", http.StatusConflict)
		return
	}

	isNew := errors.Is(err, sql.ErrNoRows)
	if isNew {
		randomSecret, randomErr := randomKey(32)
		if randomErr != nil {
			a.serverError(w, r, randomErr)
			return
		}
		randomPassword := "A!" + randomSecret + "z9"
		hash, hashErr := auth.HashPassword(randomPassword)
		if hashErr != nil {
			a.serverError(w, r, hashErr)
			return
		}
		result, createErr := tx.ExecContext(r.Context(), `
			INSERT INTO users (role,status,name,email,phone,password_hash,approved_at,approved_by)
			VALUES ('tenant','active',?,?,?,?,UTC_TIMESTAMP(6),?)`, name, email, phone, hash, session.User.ID)
		if createErr != nil {
			if isDuplicate(createErr) {
				http.Error(w, "Email pemohon baru saja digunakan. Muat ulang halaman.", http.StatusConflict)
				return
			}
			a.serverError(w, r, createErr)
			return
		}
		userID, _ = result.LastInsertId()
		if _, err = tx.ExecContext(r.Context(), `INSERT INTO tenant_profiles (user_id) VALUES (?)`, userID); err == nil {
			_, err = tx.ExecContext(r.Context(), `INSERT INTO registration_requests (user_id,status,reviewed_by,reviewed_at,review_reason) VALUES (?,'approved',?,UTC_TIMESTAMP(6),'Disetujui dari booking dengan dana terverifikasi')`, userID, session.User.ID)
		}
	} else if userStatus != "active" {
		_, err = tx.ExecContext(r.Context(), `UPDATE users SET status='active',approved_at=UTC_TIMESTAMP(6),approved_by=? WHERE id=?`, session.User.ID, userID)
		if err == nil {
			_, err = tx.ExecContext(r.Context(), `UPDATE registration_requests SET status='approved',reviewed_by=?,reviewed_at=UTC_TIMESTAMP(6),review_reason='Disetujui dari booking dengan dana terverifikasi' WHERE user_id=? AND status IN ('pending_email','pending_review','rejected')`, session.User.ID, userID)
		}
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if _, err = tx.ExecContext(r.Context(), `UPDATE booking_requests SET applicant_user_id=? WHERE id=?`, userID, bookingID); err != nil {
		a.serverError(w, r, err)
		return
	}

	eventKey := fmt.Sprintf("booking.account.%d", bookingID)
	if !verified.Valid {
		rawToken, tokenErr := randomKey(32)
		if tokenErr != nil {
			a.serverError(w, r, tokenErr)
			return
		}
		tokenHash := sha256.Sum256([]byte(rawToken))
		if _, err = tx.ExecContext(r.Context(), `UPDATE password_reset_tokens SET used_at=UTC_TIMESTAMP(6) WHERE user_id=? AND used_at IS NULL`, userID); err == nil {
			_, err = tx.ExecContext(r.Context(), `INSERT INTO password_reset_tokens (user_id,token_hash,expires_at) VALUES (?,?,UTC_TIMESTAMP(6)+INTERVAL 24 HOUR)`, userID, hex.EncodeToString(tokenHash[:]))
		}
		activationURL := strings.TrimRight(a.cfg.BaseURL, "/") + "/reset-password?token=" + rawToken
		if err == nil {
			_, err = tx.ExecContext(r.Context(), `INSERT INTO notification_deliveries (user_id,channel,event_key,recipient,subject,body) VALUES (?,'email',?,?,?,?) ON DUPLICATE KEY UPDATE status='queued',attempts=0,claimed_at=NULL,last_error=NULL,sent_at=NULL,body=VALUES(body)`, userID, eventKey, email, "Aktifkan akun penghuni Alam Segar", "Booking Anda telah dikonfirmasi. Buat kata sandi melalui tautan berikut (berlaku 24 jam): "+activationURL)
		}
	} else {
		if err = createNotificationTx(r.Context(), tx, userID, eventKey, "Booking terhubung", "Booking Anda telah dikonfirmasi dan terhubung ke akun penghuni.", "/penghuni"); err == nil {
			_, err = tx.ExecContext(r.Context(), `INSERT INTO notification_deliveries (user_id,channel,event_key,recipient,subject,body) VALUES (?,'email',?,?,?,?) ON DUPLICATE KEY UPDATE status='queued',attempts=0,claimed_at=NULL,last_error=NULL,sent_at=NULL`, userID, eventKey, email, "Booking terhubung ke akun Alam Segar", "Booking Anda telah dikonfirmasi. Masuk ke portal penghuni untuk melihat proses kontrak.")
		}
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = insertAuditTx(r.Context(), tx, session.User.ID, "booking.provision_tenant", "booking_request", bookingID, nil, map[string]any{"tenant_id": userID, "new_account": isNew, "activation_required": !verified.Valid}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/booking?berhasil=akun", http.StatusSeeOther)
}
