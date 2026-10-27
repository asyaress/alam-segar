package httpapp

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"alam-segar/kos-management/internal/auth"
)

type RegistrationView struct {
	ID        int64
	UserID    int64
	Name      string
	Email     string
	Phone     string
	Status    string
	Verified  bool
	CreatedAt time.Time
}

type TenantView struct {
	ID             int64
	Name           string
	Email          string
	Phone          string
	Status         string
	ActiveLease    string
	EmergencyName  string
	EmergencyPhone string
}

type PeoplePage struct {
	Registrations []RegistrationView
	Tenants       []TenantView
	Form          map[string]string
}

func (a *App) registrationForm(w http.ResponseWriter, r *http.Request) {
	session := sessionFrom(r.Context())
	if session != nil && session.User != nil {
		a.redirectForRole(w, r, session.User.Role)
		return
	}
	if session == nil {
		var err error
		session, err = a.sessions.Start(r.Context(), w, nil)
		if err != nil {
			a.serverError(w, r, err)
			return
		}
	}
	a.render(w, http.StatusOK, "public_register", PageData{Title: "Bikin Akun — Alam Segar", CSRFToken: session.CSRFToken, Payload: PeoplePage{Form: map[string]string{}}})
}

func (a *App) registerTenant(w http.ResponseWriter, r *http.Request) {
	session := sessionFrom(r.Context())
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	form := map[string]string{
		"name": strings.TrimSpace(r.FormValue("name")), "email": strings.ToLower(strings.TrimSpace(r.FormValue("email"))),
		"phone": strings.TrimSpace(r.FormValue("phone")), "emergency_name": strings.TrimSpace(r.FormValue("emergency_name")),
		"emergency_phone": strings.TrimSpace(r.FormValue("emergency_phone")),
	}
	data := PageData{Title: "Bikin Akun — Alam Segar", CSRFToken: session.CSRFToken, Payload: PeoplePage{Form: form}}
	parsed, emailErr := mail.ParseAddress(form["email"])
	password := r.FormValue("password")
	passwordHash, passwordErr := auth.HashPassword(password)
	switch {
	case len([]rune(form["name"])) < 3 || len([]rune(form["name"])) > 150:
		data.Error = "Nama harus berisi 3 sampai 150 karakter."
	case emailErr != nil || parsed.Address != form["email"] || len(form["email"]) > 190:
		data.Error = "Email tidak valid."
	case len(form["phone"]) < 8 || len(form["phone"]) > 40:
		data.Error = "Nomor telepon harus berisi 8 sampai 40 karakter."
	case passwordErr != nil:
		data.Error = passwordErr.Error()
	case form["emergency_phone"] != "" && len(form["emergency_phone"]) < 8:
		data.Error = "Nomor kontak darurat tidak valid."
	}
	if data.Error != "" {
		a.render(w, http.StatusUnprocessableEntity, "public_register", data)
		return
	}

	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(r.Context(), `
		INSERT INTO users (role, status, name, email, phone, password_hash)
		VALUES ('tenant','pending',?,?,?,?)`, form["name"], form["email"], nullableString(form["phone"]), passwordHash)
	if err != nil {
		if isDuplicate(err) {
			data.Error = "Email sudah terdaftar. Gunakan halaman masuk atau hubungi pemilik."
			a.render(w, http.StatusConflict, "public_register", data)
			return
		}
		a.serverError(w, r, err)
		return
	}
	userID, _ := result.LastInsertId()
	if _, err := tx.ExecContext(r.Context(), `
		INSERT INTO tenant_profiles (user_id, emergency_contact_name, emergency_contact_phone)
		VALUES (?, ?, ?)`, userID, nullableString(form["emergency_name"]), nullableString(form["emergency_phone"])); err != nil {
		a.serverError(w, r, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `INSERT INTO registration_requests (user_id,status) VALUES (?,'pending_email')`, userID); err != nil {
		a.serverError(w, r, err)
		return
	}
	rawToken, err := randomKey(32)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	tokenHash := sha256.Sum256([]byte(rawToken))
	if _, err := tx.ExecContext(r.Context(), `
		INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
		VALUES (?, ?, UTC_TIMESTAMP(6) + INTERVAL 24 HOUR)`, userID, hex.EncodeToString(tokenHash[:])); err != nil {
		a.serverError(w, r, err)
		return
	}
	verificationURL := strings.TrimRight(a.cfg.BaseURL, "/") + "/verifikasi-email?token=" + rawToken
	body := "Verifikasi email pendaftaran Alam Segar melalui tautan berikut: " + verificationURL
	if _, err := tx.ExecContext(r.Context(), `
		INSERT INTO notification_deliveries (user_id, channel, event_key, recipient, subject, body)
		VALUES (?, 'email', ?, ?, 'Verifikasi email Alam Segar', ?)`, userID, fmt.Sprintf("registration.verify.%d", userID), form["email"], body); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := insertAuditTx(r.Context(), tx, userID, "registration.submit", "user", userID, nil, map[string]any{"email": form["email"], "status": "pending_email"}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	data.Success = "Pendaftaran diterima. Verifikasi email, lalu tunggu persetujuan pemilik."
	data.Payload = PeoplePage{Form: map[string]string{}}
	if a.cfg.Environment == "development" {
		data.DevLink = verificationURL
	}
	a.render(w, http.StatusCreated, "public_register", data)
}

func (a *App) verifyEmail(w http.ResponseWriter, r *http.Request) {
	rawToken := strings.TrimSpace(r.URL.Query().Get("token"))
	if len(rawToken) != 64 {
		http.Error(w, "Tautan verifikasi tidak valid.", http.StatusBadRequest)
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
	err = tx.QueryRowContext(r.Context(), `
		SELECT id, user_id FROM email_verification_tokens
		WHERE token_hash=? AND used_at IS NULL AND expires_at > UTC_TIMESTAMP(6) FOR UPDATE`, hex.EncodeToString(hash[:])).Scan(&tokenID, &userID)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Tautan verifikasi sudah dipakai atau kedaluwarsa.", http.StatusGone)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `UPDATE email_verification_tokens SET used_at=UTC_TIMESTAMP(6) WHERE id=?`, tokenID); err != nil {
		a.serverError(w, r, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `UPDATE users SET email_verified_at=UTC_TIMESTAMP(6) WHERE id=?`, userID); err != nil {
		a.serverError(w, r, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `UPDATE registration_requests SET status='pending_review' WHERE user_id=? AND status='pending_email'`, userID); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	a.render(w, http.StatusOK, "public_register", PageData{Title: "Email Terverifikasi — Alam Segar", Success: "Email berhasil diverifikasi. Pemilik akan meninjau pendaftaran Anda.", Payload: PeoplePage{Form: map[string]string{}}})
}

func (a *App) ownerTenantList(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	page := PeoplePage{}
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT rr.id,u.id,u.name,u.email,COALESCE(u.phone,''),rr.status,u.email_verified_at IS NOT NULL,rr.created_at
		FROM registration_requests rr JOIN users u ON u.id=rr.user_id
		WHERE u.role='tenant' ORDER BY FIELD(rr.status,'pending_review','pending_email','approved','rejected'),rr.created_at DESC`)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var item RegistrationView
		if err := rows.Scan(&item.ID, &item.UserID, &item.Name, &item.Email, &item.Phone, &item.Status, &item.Verified, &item.CreatedAt); err != nil {
			rows.Close()
			a.serverError(w, r, err)
			return
		}
		page.Registrations = append(page.Registrations, item)
	}
	rows.Close()
	rows, err = a.db.QueryContext(r.Context(), `
		SELECT u.id,u.name,u.email,COALESCE(u.phone,''),u.status,
			COALESCE(CONCAT(p.name,' / Kamar ',rm.number),'Belum ada kontrak aktif'),
			COALESCE(tp.emergency_contact_name,''),COALESCE(tp.emergency_contact_phone,'')
		FROM users u LEFT JOIN tenant_profiles tp ON tp.user_id=u.id
		LEFT JOIN leases l ON l.tenant_id=u.id AND l.status IN ('active','expiring')
		LEFT JOIN rooms rm ON rm.id=l.room_id LEFT JOIN properties p ON p.id=l.property_id
		WHERE u.role='tenant' AND u.deleted_at IS NULL ORDER BY u.name`)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item TenantView
		if err := rows.Scan(&item.ID, &item.Name, &item.Email, &item.Phone, &item.Status, &item.ActiveLease, &item.EmergencyName, &item.EmergencyPhone); err != nil {
			a.serverError(w, r, err)
			return
		}
		page.Tenants = append(page.Tenants, item)
	}
	data := PageData{Title: "Penghuni — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page}
	if r.URL.Query().Get("berhasil") != "" {
		data.Success = "Data penghuni berhasil diperbarui."
	}
	a.render(w, http.StatusOK, "owner_tenants", data)
}

func (a *App) ownerTenantCreateForm(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	a.render(w, http.StatusOK, "owner_tenant_form", PageData{Title: "Tambah Penghuni — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: PeoplePage{Form: map[string]string{}}})
}

func (a *App) ownerTenantCreate(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	form := map[string]string{"name": strings.TrimSpace(r.FormValue("name")), "email": strings.ToLower(strings.TrimSpace(r.FormValue("email"))), "phone": strings.TrimSpace(r.FormValue("phone")), "emergency_name": strings.TrimSpace(r.FormValue("emergency_name")), "emergency_phone": strings.TrimSpace(r.FormValue("emergency_phone"))}
	data := PageData{Title: "Tambah Penghuni — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: PeoplePage{Form: form}}
	hash, err := auth.HashPassword(r.FormValue("password"))
	if len(form["name"]) < 3 || len(form["phone"]) < 8 {
		data.Error = "Nama dan nomor telepon wajib diisi dengan benar."
	} else if parsed, parseErr := mail.ParseAddress(form["email"]); parseErr != nil || parsed.Address != form["email"] {
		data.Error = "Email tidak valid."
	} else if err != nil {
		data.Error = err.Error()
	}
	if data.Error != "" {
		a.render(w, http.StatusUnprocessableEntity, "owner_tenant_form", data)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(r.Context(), `
		INSERT INTO users (role,status,name,email,phone,password_hash,email_verified_at,approved_at,approved_by)
		VALUES ('tenant','active',?,?,?,?,UTC_TIMESTAMP(6),UTC_TIMESTAMP(6),?)`, form["name"], form["email"], form["phone"], hash, session.User.ID)
	if err != nil {
		if isDuplicate(err) {
			data.Error = "Email sudah digunakan."
			a.render(w, http.StatusConflict, "owner_tenant_form", data)
			return
		}
		a.serverError(w, r, err)
		return
	}
	userID, _ := result.LastInsertId()
	_, err = tx.ExecContext(r.Context(), `INSERT INTO tenant_profiles (user_id,emergency_contact_name,emergency_contact_phone) VALUES (?,?,?)`, userID, nullableString(form["emergency_name"]), nullableString(form["emergency_phone"]))
	if err == nil {
		_, err = tx.ExecContext(r.Context(), `INSERT INTO registration_requests (user_id,status,reviewed_by,reviewed_at,review_reason) VALUES (?,'approved',?,UTC_TIMESTAMP(6),'Dibuat langsung oleh pemilik')`, userID, session.User.ID)
	}
	if err == nil {
		err = insertAuditTx(r.Context(), tx, session.User.ID, "tenant.create", "user", userID, nil, map[string]any{"email": form["email"], "status": "active"}, a.sessions.HashIP(r))
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/penghuni?berhasil=dibuat", http.StatusSeeOther)
}

func (a *App) ownerApproveRegistration(w http.ResponseWriter, r *http.Request) {
	a.ownerReviewRegistration(w, r, true)
}

func (a *App) ownerRejectRegistration(w http.ResponseWriter, r *http.Request) {
	a.ownerReviewRegistration(w, r, false)
}

func (a *App) ownerReviewRegistration(w http.ResponseWriter, r *http.Request, approve bool) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	registrationID, err := parseID(r.PathValue("id"))
	reason := strings.TrimSpace(r.FormValue("reason"))
	if err != nil || (!approve && len(reason) < 5) {
		http.Error(w, "Alasan penolakan minimal 5 karakter.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var userID int64
	var status string
	var verified sql.NullTime
	var email string
	err = tx.QueryRowContext(r.Context(), `
		SELECT rr.user_id,rr.status,u.email_verified_at,u.email FROM registration_requests rr JOIN users u ON u.id=rr.user_id
		WHERE rr.id=? FOR UPDATE`, registrationID).Scan(&userID, &status, &verified, &email)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if status != "pending_review" || (approve && !verified.Valid) {
		http.Error(w, "Registrasi belum siap ditinjau atau sudah diproses.", http.StatusConflict)
		return
	}
	newStatus, userStatus, action := "rejected", "rejected", "registration.reject"
	if approve {
		newStatus, userStatus, action, reason = "approved", "active", "registration.approve", "Disetujui pemilik"
	}
	if _, err := tx.ExecContext(r.Context(), `UPDATE registration_requests SET status=?,reviewed_by=?,reviewed_at=UTC_TIMESTAMP(6),review_reason=? WHERE id=?`, newStatus, session.User.ID, reason, registrationID); err != nil {
		a.serverError(w, r, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `UPDATE users SET status=?,approved_at=IF(?='active',UTC_TIMESTAMP(6),approved_at),approved_by=? WHERE id=?`, userStatus, userStatus, session.User.ID, userID); err != nil {
		a.serverError(w, r, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `INSERT INTO notifications (user_id,event_key,title,body,action_url) VALUES (?,?,?,?,?)`, userID, fmt.Sprintf("%s.%d", action, registrationID), "Status pendaftaran", "Pendaftaran Anda "+map[bool]string{true: "disetujui", false: "ditolak"}[approve]+". "+reason, "/masuk"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `INSERT INTO notification_deliveries (user_id,channel,event_key,recipient,subject,body) VALUES (?,'email',?,?,?,?)`, userID, fmt.Sprintf("%s.%d", action, registrationID), email, "Status pendaftaran Alam Segar", "Pendaftaran Anda "+newStatus+". "+reason); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, action, "registration_request", registrationID, map[string]any{"status": status}, map[string]any{"status": newStatus, "reason": reason}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/penghuni?berhasil=ditinjau", http.StatusSeeOther)
}

func nullableTime(value sql.NullTime) any {
	if !value.Valid {
		return nil
	}
	return value.Time
}

var _ = strconv.IntSize
