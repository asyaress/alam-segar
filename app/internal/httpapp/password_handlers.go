package httpapp

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strings"

	"alam-segar/kos-management/internal/auth"
)

type PasswordPage struct {
	Email string
	Token string
}

func (a *App) publicSession(w http.ResponseWriter, r *http.Request) (*auth.Session, error) {
	session := sessionFrom(r.Context())
	if session != nil {
		return session, nil
	}
	return a.sessions.Start(r.Context(), w, nil)
}

func (a *App) forgotPasswordForm(w http.ResponseWriter, r *http.Request) {
	session, err := a.publicSession(w, r)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	a.render(w, http.StatusOK, "password", PageData{Title: "Reset Password — Alam Segar", CSRFToken: session.CSRFToken, Payload: PasswordPage{}})
}

func (a *App) forgotPassword(w http.ResponseWriter, r *http.Request) {
	session := sessionFrom(r.Context())
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	data := PageData{Title: "Reset Password — Alam Segar", CSRFToken: session.CSRFToken, Payload: PasswordPage{Email: email}}
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		data.Error = "Alamat email tidak valid."
		a.render(w, http.StatusUnprocessableEntity, "password", data)
		return
	}
	var userID int64
	err = a.db.QueryRowContext(r.Context(), `SELECT id FROM users WHERE email=? AND status='active' AND deleted_at IS NULL`, email).Scan(&userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		a.serverError(w, r, err)
		return
	}
	if err == nil {
		rawToken, tokenErr := randomKey(32)
		if tokenErr != nil {
			a.serverError(w, r, tokenErr)
			return
		}
		hash := sha256.Sum256([]byte(rawToken))
		tx, txErr := a.db.BeginTx(r.Context(), nil)
		if txErr != nil {
			a.serverError(w, r, txErr)
			return
		}
		defer tx.Rollback()
		if _, txErr = tx.ExecContext(r.Context(), `UPDATE password_reset_tokens SET used_at=UTC_TIMESTAMP(6) WHERE user_id=? AND used_at IS NULL`, userID); txErr == nil {
			_, txErr = tx.ExecContext(r.Context(), `INSERT INTO password_reset_tokens (user_id,token_hash,expires_at) VALUES (?,?,UTC_TIMESTAMP(6)+INTERVAL 60 MINUTE)`, userID, hex.EncodeToString(hash[:]))
		}
		resetURL := strings.TrimRight(a.cfg.BaseURL, "/") + "/reset-password?token=" + rawToken
		if txErr == nil {
			_, txErr = tx.ExecContext(r.Context(), `INSERT INTO notification_deliveries (user_id,channel,event_key,recipient,subject,body) VALUES (?,'email',?,?,?,?)`, userID, fmt.Sprintf("password.reset.%x", hash[:8]), email, "Reset kata sandi Alam Segar", "Atur ulang kata sandi melalui tautan berikut (berlaku 60 menit): "+resetURL)
		}
		if txErr != nil {
			a.serverError(w, r, txErr)
			return
		}
		if txErr = tx.Commit(); txErr != nil {
			a.serverError(w, r, txErr)
			return
		}
		if a.cfg.Environment == "development" {
			data.DevLink = resetURL
		}
	}
	data.Success = "Jika email terdaftar dan aktif, tautan reset telah dimasukkan ke antrean email."
	data.Payload = PasswordPage{}
	a.render(w, http.StatusOK, "password", data)
}

func (a *App) resetPasswordForm(w http.ResponseWriter, r *http.Request) {
	session, err := a.publicSession(w, r)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	rawToken := strings.TrimSpace(r.URL.Query().Get("token"))
	if len(rawToken) != 64 {
		http.Error(w, "Tautan reset tidak valid.", http.StatusBadRequest)
		return
	}
	hash := sha256.Sum256([]byte(rawToken))
	var count int
	if err := a.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM password_reset_tokens WHERE token_hash=? AND used_at IS NULL AND expires_at>UTC_TIMESTAMP(6)`, hex.EncodeToString(hash[:])).Scan(&count); err != nil {
		a.serverError(w, r, err)
		return
	}
	if count != 1 {
		http.Error(w, "Tautan reset sudah dipakai atau kedaluwarsa.", http.StatusGone)
		return
	}
	a.render(w, http.StatusOK, "password", PageData{Title: "Bikin Password Baru — Alam Segar", CSRFToken: session.CSRFToken, Payload: PasswordPage{Token: rawToken}})
}

func (a *App) resetPassword(w http.ResponseWriter, r *http.Request) {
	session := sessionFrom(r.Context())
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	rawToken := strings.TrimSpace(r.FormValue("token"))
	password := r.FormValue("password")
	data := PageData{Title: "Bikin Password Baru — Alam Segar", CSRFToken: session.CSRFToken, Payload: PasswordPage{Token: rawToken}}
	if rawToken == "" || password != r.FormValue("password_confirmation") {
		data.Error = "Konfirmasi kata sandi tidak cocok."
		a.render(w, http.StatusUnprocessableEntity, "password", data)
		return
	}
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		data.Error = err.Error()
		a.render(w, http.StatusUnprocessableEntity, "password", data)
		return
	}
	hash := sha256.Sum256([]byte(rawToken))
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var tokenID, userID int64
	err = tx.QueryRowContext(r.Context(), `SELECT id,user_id FROM password_reset_tokens WHERE token_hash=? AND used_at IS NULL AND expires_at>UTC_TIMESTAMP(6) FOR UPDATE`, hex.EncodeToString(hash[:])).Scan(&tokenID, &userID)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Tautan reset sudah dipakai atau kedaluwarsa.", http.StatusGone)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if _, err = tx.ExecContext(r.Context(), `UPDATE users SET password_hash=?,email_verified_at=COALESCE(email_verified_at,UTC_TIMESTAMP(6)) WHERE id=? AND status='active'`, passwordHash, userID); err == nil {
		_, err = tx.ExecContext(r.Context(), `UPDATE password_reset_tokens SET used_at=UTC_TIMESTAMP(6) WHERE id=?`, tokenID)
	}
	if err == nil {
		_, err = tx.ExecContext(r.Context(), `DELETE FROM user_sessions WHERE user_id=?`, userID)
	}
	if err == nil {
		err = insertAuditTx(r.Context(), tx, userID, "auth.password_reset", "user", userID, nil, map[string]any{"sessions_revoked": true}, a.sessions.HashIP(r))
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	data.Success = "Kata sandi berhasil diubah. Semua sesi lama telah dikeluarkan."
	data.Payload = PasswordPage{}
	a.render(w, http.StatusOK, "password", data)
}
