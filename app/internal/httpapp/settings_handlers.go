package httpapp

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type OwnerSettingsPage struct {
	BusinessName, SignerName, SignerTitle, GoLiveDate string
	ReminderBeforeDays, ReminderOverdueDays           int
	LeaseExpiringDays                                 int
}

func ownerSettingKey(ownerID int64, name string) string {
	return fmt.Sprintf("owner.%d.%s", ownerID, name)
}

func (a *App) ownerSettings(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	page, err := a.loadOwnerSettings(r, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	data := PageData{Title: "Pengaturan — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page}
	if r.URL.Query().Get("berhasil") != "" {
		data.Success = "Pengaturan berhasil disimpan."
	}
	a.render(w, http.StatusOK, "owner_settings", data)
}

func (a *App) loadOwnerSettings(r *http.Request, ownerID int64) (OwnerSettingsPage, error) {
	page := OwnerSettingsPage{BusinessName: "Alam Segar", SignerTitle: "Pemilik", ReminderBeforeDays: 3, ReminderOverdueDays: 3, LeaseExpiringDays: 30}
	_ = a.db.QueryRowContext(r.Context(), `SELECT name FROM users WHERE id=?`, ownerID).Scan(&page.SignerName)
	rows, err := a.db.QueryContext(r.Context(), `SELECT setting_key,setting_value FROM app_settings WHERE setting_key LIKE ?`, fmt.Sprintf("owner.%d.%%", ownerID))
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return page, err
		}
		switch key {
		case ownerSettingKey(ownerID, "business_name"):
			page.BusinessName = value
		case ownerSettingKey(ownerID, "signer_name"):
			page.SignerName = value
		case ownerSettingKey(ownerID, "signer_title"):
			page.SignerTitle = value
		case ownerSettingKey(ownerID, "go_live_date"):
			page.GoLiveDate = value
		case ownerSettingKey(ownerID, "reminder_before_days"):
			page.ReminderBeforeDays, _ = strconv.Atoi(value)
		case ownerSettingKey(ownerID, "reminder_overdue_days"):
			page.ReminderOverdueDays, _ = strconv.Atoi(value)
		case ownerSettingKey(ownerID, "lease_expiring_days"):
			page.LeaseExpiringDays, _ = strconv.Atoi(value)
		}
	}
	return page, rows.Err()
}

func (a *App) ownerSettingsUpdate(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	businessName := strings.TrimSpace(r.FormValue("business_name"))
	signerName := strings.TrimSpace(r.FormValue("signer_name"))
	signerTitle := strings.TrimSpace(r.FormValue("signer_title"))
	goLiveDate := strings.TrimSpace(r.FormValue("go_live_date"))
	beforeDays, e1 := strconv.Atoi(r.FormValue("reminder_before_days"))
	overdueDays, e2 := strconv.Atoi(r.FormValue("reminder_overdue_days"))
	expiringDays, e3 := strconv.Atoi(r.FormValue("lease_expiring_days"))
	if len([]rune(businessName)) < 3 || len([]rune(businessName)) > 150 || len([]rune(signerName)) < 3 || len([]rune(signerName)) > 150 || len([]rune(signerTitle)) < 2 || len([]rune(signerTitle)) > 100 || e1 != nil || e2 != nil || e3 != nil || beforeDays < 1 || beforeDays > 30 || overdueDays < 1 || overdueDays > 30 || expiringDays < 7 || expiringDays > 180 {
		http.Error(w, "Nama usaha atau rentang jadwal pengingat tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	if goLiveDate != "" {
		if _, err := parseBusinessDate(goLiveDate); err != nil {
			http.Error(w, "Tanggal mulai penggunaan tidak valid.", http.StatusUnprocessableEntity)
			return
		}
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	settings := map[string]string{
		"business_name":         businessName,
		"signer_name":           signerName,
		"signer_title":          signerTitle,
		"go_live_date":          goLiveDate,
		"reminder_before_days":  strconv.Itoa(beforeDays),
		"reminder_overdue_days": strconv.Itoa(overdueDays),
		"lease_expiring_days":   strconv.Itoa(expiringDays),
	}
	for key, value := range settings {
		if _, err = tx.ExecContext(r.Context(), `INSERT INTO app_settings (setting_key,setting_value,updated_by) VALUES (?,?,?) ON DUPLICATE KEY UPDATE setting_value=VALUES(setting_value),updated_by=VALUES(updated_by)`, ownerSettingKey(session.User.ID, key), value, session.User.ID); err != nil {
			a.serverError(w, r, err)
			return
		}
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "settings.update", "app_settings", session.User.ID, nil, settings, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/pengaturan?berhasil=1", http.StatusSeeOther)
}
