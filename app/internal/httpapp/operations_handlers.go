package httpapp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"alam-segar/kos-management/internal/auth"
)

type AssetView struct {
	ID                                                     int64
	Code, Name, Category, Property, Room, Condition, Notes string
	Cost                                                   int64
}
type AssetPage struct {
	Assets            []AssetView
	Properties, Rooms []LeaseOption
}
type TicketView struct {
	ID                                                                                     int64
	Number, Tenant, Property, Room, Category, Urgency, Subject, Description, Status, Photo string
	CreatedAt                                                                              time.Time
	Updates                                                                                []TicketUpdateView
}
type TicketUpdateView struct {
	Status, Message, Vendor, Actor string
	CreatedAt                      time.Time
}
type TicketPage struct {
	Tickets       []TicketView
	ActiveLeaseID int64
}
type AnnouncementView struct {
	ID                    int64
	Title, Body, Property string
	PublishedAt           time.Time
}
type AnnouncementPage struct {
	Items      []AnnouncementView
	Properties []LeaseOption
}
type TenantRoomPage struct {
	HasRoom                                                                                                                 bool
	Property, Room, Type, Rules, Description, Electricity, Water, Block, Floor, Area, CoverPhoto, LeaseNumber, BillingCycle string
	AgreedRent                                                                                                              int64
	LeaseStart, LeaseEnd, CheckIn                                                                                           time.Time
	HasCheckIn                                                                                                              bool
	Amenities                                                                                                               []string
	Assets                                                                                                                  []AssetView
}
type TenantLeasePage struct {
	Leases       []LeaseListItem
	Versions     []LeaseVersionView
	Inspections  []InspectionView
	Terminations []LeaseTerminationView
}
type NotificationView struct {
	ID               int64
	Title, Body, URL string
	Read             bool
	CreatedAt        time.Time
}
type TenantProfilePage struct{ Phone, EmergencyName, EmergencyPhone string }
type TenantDashboardPage struct {
	HasLease                                                                          bool
	Property, Room, LeaseStatus, InvoiceNumber, LeaseNumber, BillingCycle, CoverPhoto string
	LeaseStart, LeaseEnd, DueDate                                                     time.Time
	AgreedRent, Outstanding                                                           int64
	UnreadNotifications, OpenTickets                                                  int64
	SmartActions, WaitingParcels, OpenEmergencies                                     int64
	PendingDocuments                                                                  int64
}

func (a *App) loadTenantDashboard(ctx context.Context, tenantID int64) (TenantDashboardPage, error) {
	var page TenantDashboardPage
	err := a.db.QueryRowContext(ctx, `SELECT p.name,rm.number,l.status,l.lease_number,l.billing_cycle,l.start_date,l.end_date,l.agreed_rent,COALESCE((SELECT rp.storage_path FROM room_photos rp WHERE rp.room_id=rm.id AND rp.is_public=1 ORDER BY rp.sort_order,rp.id LIMIT 1),'') FROM leases l JOIN properties p ON p.id=l.property_id JOIN rooms rm ON rm.id=l.room_id WHERE l.tenant_id=? AND l.status IN ('active','expiring') ORDER BY l.start_date DESC LIMIT 1`, tenantID).Scan(&page.Property, &page.Room, &page.LeaseStatus, &page.LeaseNumber, &page.BillingCycle, &page.LeaseStart, &page.LeaseEnd, &page.AgreedRent, &page.CoverPhoto)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return page, err
	}
	page.HasLease = err == nil
	err = a.db.QueryRowContext(ctx, `SELECT invoice_number,due_date,total_amount-paid_amount FROM invoices WHERE tenant_id=? AND status IN ('issued','partially_paid','overdue') AND total_amount>paid_amount ORDER BY due_date LIMIT 1`, tenantID).Scan(&page.InvoiceNumber, &page.DueDate, &page.Outstanding)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return page, err
	}
	if err := a.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id=? AND read_at IS NULL`, tenantID).Scan(&page.UnreadNotifications); err != nil {
		return page, err
	}
	if err := a.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM maintenance_tickets WHERE tenant_id=? AND status NOT IN ('resolved','closed')`, tenantID).Scan(&page.OpenTickets); err != nil {
		return page, err
	}
	if err := a.db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM visitor_passes WHERE tenant_id=? AND status IN ('pending','approved','checked_in'))+
		(SELECT COUNT(*) FROM facility_reservations WHERE tenant_id=? AND status IN ('pending','approved'))+
		(SELECT COUNT(*) FROM resident_requests WHERE tenant_id=? AND status IN ('submitted','reviewing','approved')),
		(SELECT COUNT(*) FROM parcel_deliveries WHERE tenant_id=? AND status IN ('arrived','notified')),
		(SELECT COUNT(*) FROM emergency_alerts WHERE tenant_id=? AND status IN ('open','acknowledged'))`, tenantID, tenantID, tenantID, tenantID, tenantID).Scan(&page.SmartActions, &page.WaitingParcels, &page.OpenEmergencies); err != nil {
		return page, err
	}
	if err := a.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM digital_documents WHERE tenant_id=? AND status='awaiting_tenant'`, tenantID).Scan(&page.PendingDocuments); err != nil {
		return page, err
	}
	return page, nil
}

func (a *App) ownerAssetList(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	page := AssetPage{}
	rows, err := a.db.QueryContext(r.Context(), `SELECT a.id,a.asset_code,a.name,a.category,p.name,COALESCE(rm.number,'Area Umum'),a.condition_status,COALESCE(a.notes,''),COALESCE(a.acquired_cost,0) FROM assets a JOIN properties p ON p.id=a.property_id LEFT JOIN rooms rm ON rm.id=a.room_id WHERE a.owner_id=? ORDER BY p.name,rm.number,a.name`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var v AssetView
		if err := rows.Scan(&v.ID, &v.Code, &v.Name, &v.Category, &v.Property, &v.Room, &v.Condition, &v.Notes, &v.Cost); err != nil {
			rows.Close()
			a.serverError(w, r, err)
			return
		}
		page.Assets = append(page.Assets, v)
	}
	rows.Close()
	page.Properties, page.Rooms, err = a.ownerPropertyRoomOptions(r.Context(), session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	data := PageData{Title: "Inventaris — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page}
	if r.URL.Query().Get("berhasil") != "" {
		data.Success = "Aset berhasil ditambahkan."
	}
	a.render(w, http.StatusOK, "owner_assets", data)
}

func (a *App) ownerCreateAsset(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	propertyID, e1 := parseID(r.FormValue("property_id"))
	roomID := optionalID(r.FormValue("room_id"))
	code := strings.ToUpper(strings.TrimSpace(r.FormValue("code")))
	name := strings.TrimSpace(r.FormValue("name"))
	category := strings.TrimSpace(r.FormValue("category"))
	condition := r.FormValue("condition")
	cost, _ := strconv.ParseInt(r.FormValue("cost"), 10, 64)
	if e1 != nil || len(code) < 2 || len(name) < 2 || len(category) < 2 || !map[string]bool{"good": true, "fair": true, "damaged": true, "lost": true, "disposed": true}[condition] {
		http.Error(w, "Data aset tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var valid int
	if roomID.Valid {
		_ = tx.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM rooms rm JOIN properties p ON p.id=rm.property_id WHERE rm.id=? AND p.id=? AND p.owner_id=?`, roomID.Int64, propertyID, session.User.ID).Scan(&valid)
	} else {
		_ = tx.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM properties WHERE id=? AND owner_id=?`, propertyID, session.User.ID).Scan(&valid)
	}
	if valid != 1 {
		http.Error(w, "Lokasi aset tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	result, err := tx.ExecContext(r.Context(), `INSERT INTO assets (owner_id,property_id,room_id,asset_code,name,category,serial_number,acquired_date,acquired_cost,condition_status,is_fixed_asset,notes) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, session.User.ID, propertyID, nullableInt64(roomID), code, name, category, nullableString(strings.TrimSpace(r.FormValue("serial_number"))), nullableDate(r.FormValue("acquired_date")), nullablePositive(cost), condition, r.FormValue("is_fixed_asset") == "1", nullableString(strings.TrimSpace(r.FormValue("notes"))))
	if err != nil {
		if isDuplicate(err) {
			http.Error(w, "Kode aset sudah digunakan.", http.StatusConflict)
			return
		}
		a.serverError(w, r, err)
		return
	}
	id, _ := result.LastInsertId()
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "asset.create", "asset", id, nil, map[string]any{"code": code, "name": name}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/inventaris?berhasil=dibuat", http.StatusSeeOther)
}

func (a *App) ownerTicketList(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	page := TicketPage{}
	rows, err := a.db.QueryContext(r.Context(), `SELECT mt.id,mt.ticket_number,u.name,p.name,rm.number,mt.category,mt.urgency,mt.subject,mt.description,mt.status,COALESCE(mt.photo_storage_path,''),mt.created_at FROM maintenance_tickets mt JOIN users u ON u.id=mt.tenant_id JOIN properties p ON p.id=mt.property_id JOIN rooms rm ON rm.id=mt.room_id WHERE p.owner_id=? ORDER BY FIELD(mt.status,'open','in_progress','waiting_for_tenant','resolved','closed'),mt.created_at DESC`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var v TicketView
		if err := rows.Scan(&v.ID, &v.Number, &v.Tenant, &v.Property, &v.Room, &v.Category, &v.Urgency, &v.Subject, &v.Description, &v.Status, &v.Photo, &v.CreatedAt); err != nil {
			rows.Close()
			a.serverError(w, r, err)
			return
		}
		updates, queryErr := a.ticketUpdates(r.Context(), v.ID)
		if queryErr != nil {
			rows.Close()
			a.serverError(w, r, queryErr)
			return
		}
		v.Updates = updates
		page.Tickets = append(page.Tickets, v)
	}
	rows.Close()
	data := PageData{Title: "Keluhan & Pemeliharaan — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page}
	if r.URL.Query().Get("berhasil") != "" {
		data.Success = "Status tiket berhasil diperbarui."
	}
	a.render(w, http.StatusOK, "owner_tickets", data)
}

func (a *App) ownerUpdateTicket(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	id, err := parseID(r.PathValue("id"))
	newStatus := r.FormValue("status")
	message := strings.TrimSpace(r.FormValue("message"))
	if err != nil || !map[string]bool{"open": true, "in_progress": true, "waiting_for_tenant": true, "resolved": true, "closed": true}[newStatus] || len(message) < 3 {
		http.Error(w, "Status atau tanggapan tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var tenantID int64
	var oldStatus string
	err = tx.QueryRowContext(r.Context(), `SELECT mt.tenant_id,mt.status FROM maintenance_tickets mt JOIN properties p ON p.id=mt.property_id WHERE mt.id=? AND p.owner_id=? FOR UPDATE`, id, session.User.ID).Scan(&tenantID, &oldStatus)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_, err = tx.ExecContext(r.Context(), `UPDATE maintenance_tickets SET status=?,closed_at=IF(?='closed',UTC_TIMESTAMP(6),NULL) WHERE id=?`, newStatus, newStatus, id)
	if err == nil {
		_, err = tx.ExecContext(r.Context(), `INSERT INTO maintenance_updates (ticket_id,actor_user_id,status_from,status_to,message,vendor_name) VALUES (?,?,?,?,?,?)`, id, session.User.ID, oldStatus, newStatus, message, nullableString(strings.TrimSpace(r.FormValue("vendor"))))
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := createNotificationTx(r.Context(), tx, tenantID, fmt.Sprintf("ticket.status.%d.%s", id, newStatus), "Status keluhan diperbarui", message, "/penghuni/keluhan"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "ticket.update", "maintenance_ticket", id, map[string]any{"status": oldStatus}, map[string]any{"status": newStatus, "message": message}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/keluhan?berhasil=status", http.StatusSeeOther)
}

func (a *App) tenantTickets(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "tenant")
	if session == nil {
		return
	}
	page := TicketPage{}
	_ = a.db.QueryRowContext(r.Context(), `SELECT id FROM leases WHERE tenant_id=? AND status IN ('active','expiring') ORDER BY start_date DESC LIMIT 1`, session.User.ID).Scan(&page.ActiveLeaseID)
	rows, err := a.db.QueryContext(r.Context(), `SELECT id,ticket_number,'','','',category,urgency,subject,description,status,COALESCE(photo_storage_path,''),created_at FROM maintenance_tickets WHERE tenant_id=? ORDER BY created_at DESC`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var v TicketView
		if err := rows.Scan(&v.ID, &v.Number, &v.Tenant, &v.Property, &v.Room, &v.Category, &v.Urgency, &v.Subject, &v.Description, &v.Status, &v.Photo, &v.CreatedAt); err != nil {
			rows.Close()
			a.serverError(w, r, err)
			return
		}
		updates, queryErr := a.ticketUpdates(r.Context(), v.ID)
		if queryErr != nil {
			rows.Close()
			a.serverError(w, r, queryErr)
			return
		}
		v.Updates = updates
		page.Tickets = append(page.Tickets, v)
	}
	rows.Close()
	data := PageData{Title: "Keluhan Saya — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page}
	if r.URL.Query().Get("berhasil") != "" {
		data.Success = "Keluhan berhasil dikirim."
	}
	a.render(w, http.StatusOK, "tenant_tickets", data)
}

func (a *App) tenantCreateTicket(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "tenant")
	if session == nil {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20)
	if err := r.ParseMultipartForm(8 << 20); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	category, urgency := r.FormValue("category"), r.FormValue("urgency")
	subject := strings.TrimSpace(r.FormValue("subject"))
	description := strings.TrimSpace(r.FormValue("description"))
	if !map[string]bool{"electricity": true, "water": true, "ac": true, "furniture": true, "security": true, "cleanliness": true, "other": true}[category] || !map[string]bool{"low": true, "normal": true, "high": true, "emergency": true}[urgency] || len(subject) < 3 || len(description) < 5 {
		http.Error(w, "Data keluhan tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	upload, err := a.savePrivateUpload(r, "photo", "tickets", 5<<20, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var leaseID, propertyID, roomID int64
	err = tx.QueryRowContext(r.Context(), `SELECT id,property_id,room_id FROM leases WHERE tenant_id=? AND status IN ('active','expiring') ORDER BY start_date DESC LIMIT 1 FOR UPDATE`, session.User.ID).Scan(&leaseID, &propertyID, &roomID)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Keluhan hanya dapat dibuat saat kontrak aktif.", http.StatusConflict)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	number, _ := referenceNumber("TKT")
	result, err := tx.ExecContext(r.Context(), `INSERT INTO maintenance_tickets (ticket_number,tenant_id,lease_id,property_id,room_id,category,urgency,subject,description,photo_storage_path) VALUES (?,?,?,?,?,?,?,?,?,?)`, number, session.User.ID, leaseID, propertyID, roomID, category, urgency, subject, description, nullableString(upload.RelativePath))
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	id, _ := result.LastInsertId()
	_, err = tx.ExecContext(r.Context(), `INSERT INTO maintenance_updates (ticket_id,actor_user_id,status_to,message) VALUES (?,?,'open','Keluhan dibuat oleh penghuni')`, id, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := createNotificationForOwnerTx(r.Context(), tx, propertyID, fmt.Sprintf("ticket.created.%d", id), "Keluhan baru", subject, "/pemilik/keluhan"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/penghuni/keluhan?berhasil=dibuat", http.StatusSeeOther)
}

func (a *App) tenantReplyTicket(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "tenant")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	id, idErr := parseID(r.PathValue("id"))
	message := strings.TrimSpace(r.FormValue("message"))
	if idErr != nil || len(message) < 3 || len([]rune(message)) > 2000 {
		http.Error(w, "Tanggapan harus berisi 3 sampai 2.000 karakter.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var propertyID int64
	var status string
	err = tx.QueryRowContext(r.Context(), `SELECT property_id,status FROM maintenance_tickets WHERE id=? AND tenant_id=? FOR UPDATE`, id, session.User.ID).Scan(&propertyID, &status)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if status == "closed" || status == "resolved" {
		http.Error(w, "Tiket sudah selesai. Buat tiket baru jika masalah muncul kembali.", http.StatusConflict)
		return
	}
	newStatus := status
	if status == "waiting_for_tenant" {
		newStatus = "in_progress"
	}
	_, err = tx.ExecContext(r.Context(), `INSERT INTO maintenance_updates (ticket_id,actor_user_id,status_from,status_to,message) VALUES (?,?,?,?,?)`, id, session.User.ID, status, newStatus, message)
	if err == nil && newStatus != status {
		_, err = tx.ExecContext(r.Context(), `UPDATE maintenance_tickets SET status=? WHERE id=?`, newStatus, id)
	}
	if err == nil {
		err = createNotificationForOwnerTx(r.Context(), tx, propertyID, fmt.Sprintf("ticket.reply.%d.%d", id, time.Now().UnixNano()), "Tanggapan penghuni", message, "/pemilik/keluhan")
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/penghuni/keluhan?berhasil=balasan", http.StatusSeeOther)
}

func (a *App) ticketUpdates(ctx context.Context, id int64) ([]TicketUpdateView, error) {
	rows, err := a.db.QueryContext(ctx, `SELECT mu.status_to,mu.message,COALESCE(mu.vendor_name,''),u.name,mu.created_at FROM maintenance_updates mu JOIN users u ON u.id=mu.actor_user_id WHERE mu.ticket_id=? ORDER BY mu.created_at`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []TicketUpdateView
	for rows.Next() {
		var v TicketUpdateView
		if err := rows.Scan(&v.Status, &v.Message, &v.Vendor, &v.Actor, &v.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (a *App) ownerAnnouncementList(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	page := AnnouncementPage{}
	rows, err := a.db.QueryContext(r.Context(), `SELECT a.id,a.title,a.body,COALESCE(p.name,'Semua properti'),a.published_at FROM announcements a LEFT JOIN properties p ON p.id=a.property_id WHERE a.owner_id=? ORDER BY a.published_at DESC`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var v AnnouncementView
		if err := rows.Scan(&v.ID, &v.Title, &v.Body, &v.Property, &v.PublishedAt); err != nil {
			rows.Close()
			a.serverError(w, r, err)
			return
		}
		page.Items = append(page.Items, v)
	}
	rows.Close()
	props, _, err := a.ownerPropertyRoomOptions(r.Context(), session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	page.Properties = props
	data := PageData{Title: "Pengumuman — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page}
	if r.URL.Query().Get("berhasil") != "" {
		data.Success = "Pengumuman diterbitkan."
	}
	a.render(w, http.StatusOK, "owner_announcements", data)
}

func (a *App) ownerCreateAnnouncement(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	title := strings.TrimSpace(r.FormValue("title"))
	body := strings.TrimSpace(r.FormValue("body"))
	propertyID := optionalID(r.FormValue("property_id"))
	if len(title) < 3 || len(body) < 5 {
		http.Error(w, "Judul dan isi pengumuman wajib diisi.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(r.Context(), `INSERT INTO announcements (owner_id,property_id,title,body,published_at) VALUES (?,?,?,?,UTC_TIMESTAMP(6))`, session.User.ID, nullableInt64(propertyID), title, body)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	id, _ := result.LastInsertId()
	query := `SELECT DISTINCT u.id,u.email FROM users u JOIN leases l ON l.tenant_id=u.id JOIN properties p ON p.id=l.property_id WHERE p.owner_id=? AND u.status='active' AND l.status IN ('active','expiring')`
	args := []any{session.User.ID}
	if propertyID.Valid {
		query += ` AND l.property_id=?`
		args = append(args, propertyID.Int64)
	}
	rows, err := tx.QueryContext(r.Context(), query, args...)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	type recipient struct {
		id    int64
		email string
	}
	var recipients []recipient
	for rows.Next() {
		var v recipient
		_ = rows.Scan(&v.id, &v.email)
		recipients = append(recipients, v)
	}
	rows.Close()
	for _, v := range recipients {
		event := fmt.Sprintf("announcement.%d", id)
		if err := createNotificationTx(r.Context(), tx, v.id, event, title, body, "/penghuni/notifikasi"); err != nil {
			a.serverError(w, r, err)
			return
		}
		_, err = tx.ExecContext(r.Context(), `INSERT INTO notification_deliveries (user_id,channel,event_key,recipient,subject,body) VALUES (?,'email',?,?,?,?)`, v.id, event, v.email, title, body)
		if err != nil {
			a.serverError(w, r, err)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/pengumuman?berhasil=terbit", http.StatusSeeOther)
}

func (a *App) tenantRoom(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "tenant")
	if session == nil {
		return
	}
	page := TenantRoomPage{}
	var roomID int64
	var checkIn sql.NullTime
	err := a.db.QueryRowContext(r.Context(), `SELECT rm.id,p.name,rm.number,COALESCE(rt.name,'Kamar'),COALESCE(p.rules_text,''),COALESCE(rm.description,''),COALESCE(rm.electricity_customer_number,''),COALESCE(rm.water_customer_number,''),COALESCE(rm.block_name,''),COALESCE(rm.floor_name,''),COALESCE(CAST(rm.area_m2 AS CHAR),''),COALESCE((SELECT rp.storage_path FROM room_photos rp WHERE rp.room_id=rm.id AND rp.is_public=1 ORDER BY rp.sort_order,rp.id LIMIT 1),''),l.lease_number,l.billing_cycle,l.agreed_rent,l.start_date,l.end_date,l.actual_check_in_at FROM leases l JOIN rooms rm ON rm.id=l.room_id JOIN properties p ON p.id=l.property_id LEFT JOIN room_types rt ON rt.id=rm.room_type_id WHERE l.tenant_id=? AND l.status IN ('active','expiring') ORDER BY l.start_date DESC LIMIT 1`, session.User.ID).Scan(&roomID, &page.Property, &page.Room, &page.Type, &page.Rules, &page.Description, &page.Electricity, &page.Water, &page.Block, &page.Floor, &page.Area, &page.CoverPhoto, &page.LeaseNumber, &page.BillingCycle, &page.AgreedRent, &page.LeaseStart, &page.LeaseEnd, &checkIn)
	if errors.Is(err, sql.ErrNoRows) {
		a.render(w, http.StatusOK, "tenant_room", PageData{Title: "My Room — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page})
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	page.HasRoom = true
	if checkIn.Valid {
		page.CheckIn, page.HasCheckIn = checkIn.Time, true
	}
	amenities, err := a.db.QueryContext(r.Context(), `SELECT am.name FROM amenities am JOIN room_amenities ra ON ra.amenity_id=am.id WHERE ra.room_id=? ORDER BY am.name`, roomID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for amenities.Next() {
		var name string
		if err := amenities.Scan(&name); err != nil {
			amenities.Close()
			a.serverError(w, r, err)
			return
		}
		page.Amenities = append(page.Amenities, name)
	}
	amenities.Close()
	rows, err := a.db.QueryContext(r.Context(), `SELECT id,asset_code,name,category,'','',condition_status,COALESCE(notes,''),COALESCE(acquired_cost,0) FROM assets WHERE room_id=? AND condition_status<>'disposed' ORDER BY name`, roomID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var v AssetView
		_ = rows.Scan(&v.ID, &v.Code, &v.Name, &v.Category, &v.Property, &v.Room, &v.Condition, &v.Notes, &v.Cost)
		page.Assets = append(page.Assets, v)
	}
	rows.Close()
	a.render(w, http.StatusOK, "tenant_room", PageData{Title: "My Room — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page})
}

func (a *App) tenantLeases(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "tenant")
	if session == nil {
		return
	}
	page := TenantLeasePage{}
	rows, err := a.db.QueryContext(r.Context(), `SELECT l.id,l.lease_number,l.tenant_id,l.room_id,u.name,p.name,rm.number,l.billing_cycle,l.start_date,l.end_date,l.agreed_rent,l.status FROM leases l JOIN users u ON u.id=l.tenant_id JOIN properties p ON p.id=l.property_id JOIN rooms rm ON rm.id=l.room_id WHERE l.tenant_id=? ORDER BY l.start_date DESC`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var v LeaseListItem
		if err := rows.Scan(&v.ID, &v.Number, &v.TenantID, &v.RoomID, &v.TenantName, &v.PropertyName, &v.RoomNumber, &v.Cycle, &v.StartDate, &v.EndDate, &v.AgreedRent, &v.Status); err != nil {
			rows.Close()
			a.serverError(w, r, err)
			return
		}
		page.Leases = append(page.Leases, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(r.Context(), `SELECT lv.version_number,l.lease_number,lv.change_type,lv.reason,lv.effective_date FROM lease_versions lv JOIN leases l ON l.id=lv.lease_id WHERE l.tenant_id=? ORDER BY lv.effective_date DESC,lv.version_number DESC`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var v LeaseVersionView
		if err := rows.Scan(&v.Version, &v.LeaseNumber, &v.ChangeType, &v.Reason, &v.EffectiveDate); err != nil {
			rows.Close()
			a.serverError(w, r, err)
			return
		}
		page.Versions = append(page.Versions, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(r.Context(), `SELECT i.id,l.lease_number,i.inspection_type,i.inspected_at,COALESCE(i.overall_notes,'') FROM inspections i JOIN leases l ON l.id=i.lease_id WHERE l.tenant_id=? ORDER BY i.inspected_at DESC`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var v InspectionView
		if err := rows.Scan(&v.ID, &v.LeaseNumber, &v.Type, &v.InspectedAt, &v.Notes); err != nil {
			rows.Close()
			a.serverError(w, r, err)
			return
		}
		page.Inspections = append(page.Inspections, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(r.Context(), `SELECT lt.id,lt.lease_id,lt.termination_number,l.lease_number,lt.settlement_decision,lt.reason,lt.effective_date,lt.cancelled_unpaid_amount,lt.settled_paid_amount,lt.proof_storage_path IS NOT NULL FROM lease_terminations lt JOIN leases l ON l.id=lt.lease_id WHERE lt.tenant_id=? ORDER BY lt.effective_date DESC`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var v LeaseTerminationView
		if err := rows.Scan(&v.ID, &v.LeaseID, &v.Number, &v.LeaseNumber, &v.Decision, &v.Reason, &v.EffectiveDate, &v.CancelledUnpaid, &v.SettledPaid, &v.HasProof); err != nil {
			rows.Close()
			a.serverError(w, r, err)
			return
		}
		page.Terminations = append(page.Terminations, v)
	}
	rows.Close()
	a.render(w, http.StatusOK, "tenant_leases", PageData{Title: "My Contract — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page})
}

func (a *App) tenantAnnouncements(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "tenant")
	if session == nil {
		return
	}
	page := AnnouncementPage{}
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT a.id,a.title,a.body,COALESCE(p.name,'Semua properti'),a.published_at
		FROM announcements a
		LEFT JOIN properties p ON p.id=a.property_id
		WHERE EXISTS (
			SELECT 1 FROM leases l JOIN properties lp ON lp.id=l.property_id
			WHERE l.tenant_id=? AND lp.owner_id=a.owner_id
		)
		AND (a.property_id IS NULL OR EXISTS (
			SELECT 1 FROM leases l2 WHERE l2.tenant_id=? AND l2.property_id=a.property_id
		))
		ORDER BY a.published_at DESC`, session.User.ID, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var v AnnouncementView
		if err := rows.Scan(&v.ID, &v.Title, &v.Body, &v.Property, &v.PublishedAt); err != nil {
			a.serverError(w, r, err)
			return
		}
		page.Items = append(page.Items, v)
	}
	a.render(w, http.StatusOK, "tenant_announcements", PageData{Title: "Info — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page})
}

func (a *App) tenantNotifications(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "tenant")
	if session == nil {
		return
	}
	var items []NotificationView
	rows, err := a.db.QueryContext(r.Context(), `SELECT id,title,body,COALESCE(action_url,''),read_at IS NOT NULL,created_at FROM notifications WHERE user_id=? ORDER BY created_at DESC LIMIT 100`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var v NotificationView
		_ = rows.Scan(&v.ID, &v.Title, &v.Body, &v.URL, &v.Read, &v.CreatedAt)
		items = append(items, v)
	}
	rows.Close()
	a.render(w, http.StatusOK, "tenant_notifications", PageData{Title: "Notif — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: items})
}
func (a *App) tenantReadNotification(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "tenant")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_, err = a.db.ExecContext(r.Context(), `UPDATE notifications SET read_at=COALESCE(read_at,UTC_TIMESTAMP(6)) WHERE id=? AND user_id=?`, id, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/penghuni/notifikasi", http.StatusSeeOther)
}

func (a *App) tenantProfile(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "tenant")
	if session == nil {
		return
	}
	page := TenantProfilePage{}
	err := a.db.QueryRowContext(r.Context(), `SELECT COALESCE(u.phone,''),COALESCE(tp.emergency_contact_name,''),COALESCE(tp.emergency_contact_phone,'') FROM users u LEFT JOIN tenant_profiles tp ON tp.user_id=u.id WHERE u.id=?`, session.User.ID).Scan(&page.Phone, &page.EmergencyName, &page.EmergencyPhone)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	data := PageData{Title: "Profil & Keamanan — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page}
	if r.URL.Query().Get("berhasil") != "" {
		data.Success = "Profil berhasil diperbarui."
	}
	a.render(w, http.StatusOK, "tenant_profile", data)
}
func (a *App) tenantUpdateProfile(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "tenant")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	phone := strings.TrimSpace(r.FormValue("phone"))
	emergency := strings.TrimSpace(r.FormValue("emergency_name"))
	emergencyPhone := strings.TrimSpace(r.FormValue("emergency_phone"))
	if len(phone) < 8 || (emergencyPhone != "" && len(emergencyPhone) < 8) {
		http.Error(w, "Nomor telepon tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(r.Context(), `UPDATE users SET phone=? WHERE id=?`, phone, session.User.ID)
	if err == nil {
		_, err = tx.ExecContext(r.Context(), `INSERT INTO tenant_profiles (user_id,emergency_contact_name,emergency_contact_phone) VALUES (?,?,?) ON DUPLICATE KEY UPDATE emergency_contact_name=VALUES(emergency_contact_name),emergency_contact_phone=VALUES(emergency_contact_phone)`, session.User.ID, nullableString(emergency), nullableString(emergencyPhone))
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/penghuni/profil?berhasil=profil", http.StatusSeeOther)
}
func (a *App) tenantChangePassword(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "tenant")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	var currentHash string
	if err := a.db.QueryRowContext(r.Context(), `SELECT password_hash FROM users WHERE id=?`, session.User.ID).Scan(&currentHash); err != nil {
		a.serverError(w, r, err)
		return
	}
	if !auth.VerifyPassword(currentHash, r.FormValue("current_password")) {
		http.Error(w, "Kata sandi saat ini salah.", http.StatusUnauthorized)
		return
	}
	newHash, err := auth.HashPassword(r.FormValue("new_password"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(r.Context(), `UPDATE users SET password_hash=? WHERE id=?`, newHash, session.User.ID)
	if err == nil {
		_, err = tx.ExecContext(r.Context(), `DELETE FROM user_sessions WHERE user_id=? AND id<>?`, session.User.ID, session.ID)
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/penghuni/profil?berhasil=password", http.StatusSeeOther)
}

func (a *App) ownerPrivateFile(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	kind := r.PathValue("kind")
	var path string
	switch kind {
	case "payment":
		err = a.db.QueryRowContext(r.Context(), `SELECT ps.proof_storage_path FROM payment_submissions ps JOIN bank_accounts ba ON ba.id=ps.bank_account_id WHERE ps.id=? AND ba.owner_id=?`, id, session.User.ID).Scan(&path)
	case "booking-payment":
		err = a.db.QueryRowContext(r.Context(), `SELECT bps.proof_storage_path FROM booking_payment_submissions bps JOIN booking_requests b ON b.id=bps.booking_request_id JOIN properties p ON p.id=b.property_id WHERE bps.id=? AND p.owner_id=?`, id, session.User.ID).Scan(&path)
	case "expense":
		err = a.db.QueryRowContext(r.Context(), `SELECT proof_storage_path FROM expenses WHERE id=? AND owner_id=?`, id, session.User.ID).Scan(&path)
	case "inspection":
		err = a.db.QueryRowContext(r.Context(), `SELECT i.photo_storage_path FROM inspections i JOIN leases l ON l.id=i.lease_id JOIN properties p ON p.id=l.property_id WHERE i.id=? AND p.owner_id=?`, id, session.User.ID).Scan(&path)
	case "ticket":
		err = a.db.QueryRowContext(r.Context(), `SELECT mt.photo_storage_path FROM maintenance_tickets mt JOIN properties p ON p.id=mt.property_id WHERE mt.id=? AND p.owner_id=?`, id, session.User.ID).Scan(&path)
	case "payment-refund":
		err = a.db.QueryRowContext(r.Context(), `SELECT pr.proof_storage_path FROM payment_refunds pr JOIN invoices i ON i.id=pr.invoice_id JOIN properties p ON p.id=i.property_id WHERE pr.id=? AND p.owner_id=?`, id, session.User.ID).Scan(&path)
	case "credit-refund":
		err = a.db.QueryRowContext(r.Context(), `SELECT proof_storage_path FROM tenant_credit_refunds WHERE id=? AND owner_id=?`, id, session.User.ID).Scan(&path)
	case "termination-refund":
		err = a.db.QueryRowContext(r.Context(), `SELECT proof_storage_path FROM lease_terminations WHERE id=? AND owner_id=? AND settlement_decision='refund'`, id, session.User.ID).Scan(&path)
	case "booking-refund":
		err = a.db.QueryRowContext(r.Context(), `SELECT br.proof_storage_path FROM booking_refunds br JOIN booking_funds bf ON bf.id=br.booking_fund_id JOIN booking_requests b ON b.id=bf.booking_request_id JOIN properties p ON p.id=b.property_id WHERE br.id=? AND p.owner_id=?`, id, session.User.ID).Scan(&path)
	default:
		http.NotFound(w, r)
		return
	}
	if errors.Is(err, sql.ErrNoRows) || path == "" {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	a.servePrivateFile(w, r, path)
}
func (a *App) tenantPrivateFile(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "tenant")
	if session == nil {
		return
	}
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	kind := r.PathValue("kind")
	var path string
	switch kind {
	case "payment":
		err = a.db.QueryRowContext(r.Context(), `SELECT proof_storage_path FROM payment_submissions WHERE id=? AND tenant_id=?`, id, session.User.ID).Scan(&path)
	case "ticket":
		err = a.db.QueryRowContext(r.Context(), `SELECT photo_storage_path FROM maintenance_tickets WHERE id=? AND tenant_id=?`, id, session.User.ID).Scan(&path)
	case "inspection":
		err = a.db.QueryRowContext(r.Context(), `SELECT i.photo_storage_path FROM inspections i JOIN leases l ON l.id=i.lease_id WHERE i.id=? AND l.tenant_id=?`, id, session.User.ID).Scan(&path)
	case "payment-refund":
		err = a.db.QueryRowContext(r.Context(), `SELECT pr.proof_storage_path FROM payment_refunds pr JOIN payments p ON p.id=pr.payment_id WHERE pr.id=? AND p.tenant_id=?`, id, session.User.ID).Scan(&path)
	case "credit-refund":
		err = a.db.QueryRowContext(r.Context(), `SELECT proof_storage_path FROM tenant_credit_refunds WHERE id=? AND tenant_id=?`, id, session.User.ID).Scan(&path)
	case "termination-refund":
		err = a.db.QueryRowContext(r.Context(), `SELECT proof_storage_path FROM lease_terminations WHERE id=? AND tenant_id=? AND settlement_decision='refund'`, id, session.User.ID).Scan(&path)
	default:
		http.NotFound(w, r)
		return
	}
	if errors.Is(err, sql.ErrNoRows) || path == "" {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	a.servePrivateFile(w, r, path)
}

func (a *App) ownerPropertyRoomOptions(ctx context.Context, ownerID int64) ([]LeaseOption, []LeaseOption, error) {
	var properties, roomsResult []LeaseOption
	rows, err := a.db.QueryContext(ctx, `SELECT id,name FROM properties WHERE owner_id=? AND deleted_at IS NULL ORDER BY name`, ownerID)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var v LeaseOption
		_ = rows.Scan(&v.ID, &v.Label)
		properties = append(properties, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(ctx, `SELECT rm.id,CONCAT(p.name,' / Kamar ',rm.number) FROM rooms rm JOIN properties p ON p.id=rm.property_id WHERE p.owner_id=? AND rm.deleted_at IS NULL ORDER BY p.name,rm.number`, ownerID)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var v LeaseOption
		_ = rows.Scan(&v.ID, &v.Label)
		roomsResult = append(roomsResult, v)
	}
	rows.Close()
	return properties, roomsResult, nil
}
func nullableDate(value string) any {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return nil
	}
	return parsed
}
