package httpapp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"alam-segar/kos-management/internal/auth"
)

type DocumentLine struct {
	Description string `json:"description"`
	Quantity    string `json:"quantity,omitempty"`
	Amount      int64  `json:"amount,omitempty"`
}

type DigitalDocumentSnapshot struct {
	BusinessName, SignerName, SignerTitle                     string
	SourceNumber, Title, Type, TenantName, PropertyName, Room string
	IssueDate, DueDate, PeriodStart, PeriodEnd, Status        string
	Amount, PaidAmount, Outstanding                           int64
	Notes                                                     string
	Lines                                                     []DocumentLine
}

type DigitalDocumentView struct {
	ID, SourceID, Version                           int64
	Number, Type, Title, Tenant, Property, Room     string
	Status, VerificationCode, VoidReason            string
	IssuedAt, OwnerSignedAt, TenantSignedAt         time.Time
	OwnerSigned, TenantSigned, NeedsTenantSignature bool
}

type DocumentSourceOption struct {
	ID    int64
	Label string
}

type DocumentCenterPage struct {
	Documents                               []DigitalDocumentView
	Invoices, Receipts, Leases, Inspections []DocumentSourceOption
}

type DocumentDetailPage struct {
	Document    DigitalDocumentView
	Snapshot    DigitalDocumentSnapshot
	VerifyURL   string
	IsOwner     bool
	IntegrityOK bool
}

type DocumentVerifyPage struct {
	Found                   bool
	Document                DigitalDocumentView
	BusinessName, HashShort string
	IntegrityOK             bool
}

func (a *App) ownerDocuments(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "owner")
	if s == nil {
		return
	}
	page, err := a.loadOwnerDocumentCenter(r.Context(), s.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	data := PageData{Title: "Dokumen Digital — Alam Segar", User: s.User, CSRFToken: s.CSRFToken, Payload: page}
	if r.URL.Query().Get("berhasil") != "" {
		data.Success = "Dokumen digital berhasil diproses."
	}
	a.render(w, http.StatusOK, "owner_documents", data)
}

func (a *App) tenantDocuments(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "tenant")
	if s == nil {
		return
	}
	page := DocumentCenterPage{}
	rows, err := a.db.QueryContext(r.Context(), documentListSQL+` WHERE d.tenant_id=? ORDER BY d.issued_at DESC`, s.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		v, scanErr := scanDocumentView(rows)
		if scanErr != nil {
			a.serverError(w, r, scanErr)
			return
		}
		page.Documents = append(page.Documents, v)
	}
	data := PageData{Title: "My Docs — Alam Segar", User: s.User, CSRFToken: s.CSRFToken, Payload: page}
	if r.URL.Query().Get("berhasil") != "" {
		data.Success = "Dokumen berhasil ditandatangani."
	}
	a.render(w, http.StatusOK, "tenant_documents", data)
}

const documentListSQL = `SELECT d.id,d.source_id,d.version_number,d.document_number,d.document_type,d.title,u.name,p.name,COALESCE(rm.number,''),d.status,d.verification_code,COALESCE(d.void_reason,''),d.issued_at,d.owner_signed_at,d.tenant_signed_at FROM digital_documents d JOIN users u ON u.id=d.tenant_id JOIN properties p ON p.id=d.property_id LEFT JOIN leases l ON d.document_type IN ('lease','inspection') AND l.id=CASE WHEN d.document_type='lease' THEN d.source_id ELSE (SELECT i2.lease_id FROM inspections i2 WHERE i2.id=d.source_id) END LEFT JOIN rooms rm ON rm.id=COALESCE(l.room_id,(SELECT i3.room_id FROM invoices i3 WHERE d.document_type IN ('invoice','receipt') AND i3.id=CASE WHEN d.document_type='invoice' THEN d.source_id ELSE (SELECT pa.invoice_id FROM payment_allocations pa WHERE pa.payment_id=d.source_id ORDER BY pa.id LIMIT 1) END))`

type rowScanner interface{ Scan(...any) error }

func scanDocumentView(row rowScanner) (DigitalDocumentView, error) {
	var v DigitalDocumentView
	var ownerSigned, tenantSigned sql.NullTime
	err := row.Scan(&v.ID, &v.SourceID, &v.Version, &v.Number, &v.Type, &v.Title, &v.Tenant, &v.Property, &v.Room, &v.Status, &v.VerificationCode, &v.VoidReason, &v.IssuedAt, &ownerSigned, &tenantSigned)
	if ownerSigned.Valid {
		v.OwnerSigned = true
		v.OwnerSignedAt = ownerSigned.Time
	}
	if tenantSigned.Valid {
		v.TenantSigned = true
		v.TenantSignedAt = tenantSigned.Time
	}
	v.NeedsTenantSignature = v.Type == "lease" || v.Type == "inspection"
	return v, err
}

func (a *App) loadOwnerDocumentCenter(ctx context.Context, ownerID int64) (DocumentCenterPage, error) {
	var page DocumentCenterPage
	rows, err := a.db.QueryContext(ctx, documentListSQL+` WHERE d.owner_id=? ORDER BY d.issued_at DESC`, ownerID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		v, e := scanDocumentView(rows)
		if e != nil {
			rows.Close()
			return page, e
		}
		page.Documents = append(page.Documents, v)
	}
	if err = rows.Close(); err != nil {
		return page, err
	}
	queries := []struct {
		target *[]DocumentSourceOption
		sql    string
	}{
		{&page.Invoices, `SELECT i.id,CONCAT(i.invoice_number,' — ',u.name,' — Rp',FORMAT(i.total_amount,0,'id_ID')) FROM invoices i JOIN users u ON u.id=i.tenant_id JOIN properties p ON p.id=i.property_id WHERE p.owner_id=? AND i.status<>'draft' AND NOT EXISTS(SELECT 1 FROM digital_documents d WHERE d.document_type='invoice' AND d.source_id=i.id AND d.status<>'void') ORDER BY i.issue_date DESC LIMIT 200`},
		{&page.Receipts, `SELECT pay.id,CONCAT(pay.receipt_number,' — ',u.name,' — Rp',FORMAT(pay.amount,0,'id_ID')) FROM payments pay JOIN users u ON u.id=pay.tenant_id JOIN payment_allocations pa ON pa.payment_id=pay.id JOIN invoices i ON i.id=pa.invoice_id JOIN properties p ON p.id=i.property_id WHERE p.owner_id=? AND NOT EXISTS(SELECT 1 FROM digital_documents d WHERE d.document_type='receipt' AND d.source_id=pay.id AND d.status<>'void') GROUP BY pay.id,pay.receipt_number,u.name,pay.amount,pay.payment_date ORDER BY pay.payment_date DESC LIMIT 200`},
		{&page.Leases, `SELECT l.id,CONCAT(l.lease_number,' — ',u.name,' — Kamar ',rm.number) FROM leases l JOIN users u ON u.id=l.tenant_id JOIN properties p ON p.id=l.property_id JOIN rooms rm ON rm.id=l.room_id WHERE p.owner_id=? AND NOT EXISTS(SELECT 1 FROM digital_documents d WHERE d.document_type='lease' AND d.source_id=l.id AND d.status<>'void') ORDER BY l.created_at DESC LIMIT 200`},
		{&page.Inspections, `SELECT ins.id,CONCAT(l.lease_number,' — ',IF(ins.inspection_type='check_in','Serah terima masuk','Serah terima keluar'),' — ',u.name) FROM inspections ins JOIN leases l ON l.id=ins.lease_id JOIN users u ON u.id=l.tenant_id JOIN properties p ON p.id=l.property_id WHERE p.owner_id=? AND NOT EXISTS(SELECT 1 FROM digital_documents d WHERE d.document_type='inspection' AND d.source_id=ins.id AND d.status<>'void') ORDER BY ins.inspected_at DESC LIMIT 200`},
	}
	for _, q := range queries {
		rows, e := a.db.QueryContext(ctx, q.sql, ownerID)
		if e != nil {
			return page, e
		}
		for rows.Next() {
			var v DocumentSourceOption
			if e = rows.Scan(&v.ID, &v.Label); e != nil {
				rows.Close()
				return page, e
			}
			*q.target = append(*q.target, v)
		}
		if e = rows.Close(); e != nil {
			return page, e
		}
	}
	return page, nil
}

func (a *App) ownerIssueDocument(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "owner")
	if s == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(s, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	kind := r.FormValue("document_type")
	sourceID, e := parseID(r.FormValue("source_id"))
	if e != nil || !map[string]bool{"invoice": true, "receipt": true, "lease": true, "inspection": true}[kind] {
		http.Error(w, "Sumber dokumen tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	snapshot, ownerID, propertyID, tenantID, err := a.buildDocumentSnapshotTx(r.Context(), tx, s.User.ID, kind, sourceID)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	var active, maxVersion int64
	if err = tx.QueryRowContext(r.Context(), `SELECT COUNT(CASE WHEN status<>'void' THEN 1 END),COALESCE(MAX(version_number),0) FROM digital_documents WHERE document_type=? AND source_id=? FOR UPDATE`, kind, sourceID).Scan(&active, &maxVersion); err != nil {
		a.serverError(w, r, err)
		return
	}
	if active > 0 {
		http.Error(w, "Dokumen aktif untuk sumber ini sudah tersedia.", http.StatusConflict)
		return
	}
	body, err := json.Marshal(snapshot)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	sum := sha256.Sum256(body)
	code, err := randomKey(16)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	number, err := referenceNumber("DOK")
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	result, err := tx.ExecContext(r.Context(), `INSERT INTO digital_documents(owner_id,property_id,tenant_id,document_type,source_id,document_number,version_number,title,snapshot_json,snapshot_sha256,verification_code,issued_by,issued_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,UTC_TIMESTAMP(6))`, ownerID, propertyID, tenantID, kind, sourceID, number, maxVersion+1, snapshot.Title, string(body), hex.EncodeToString(sum[:]), code, s.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	id, _ := result.LastInsertId()
	if err = notifyUserTx(r.Context(), tx, tenantID, fmt.Sprintf("document.issued.%d", id), "Dokumen digital baru", snapshot.Title+" telah diterbitkan dan menunggu tanda tangan pemilik.", "/penghuni/dokumen"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = insertAuditTx(r.Context(), tx, s.User.ID, "document.issue", "digital_document", id, nil, map[string]any{"type": kind, "source_id": sourceID, "hash": hex.EncodeToString(sum[:])}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/dokumen?berhasil=terbit", http.StatusSeeOther)
}

func (a *App) buildDocumentSnapshotTx(ctx context.Context, tx *sql.Tx, ownerID int64, kind string, sourceID int64) (DigitalDocumentSnapshot, int64, int64, int64, error) {
	var s DigitalDocumentSnapshot
	s.Type = kind
	_ = tx.QueryRowContext(ctx, `SELECT COALESCE((SELECT setting_value FROM app_settings WHERE setting_key=?),'Alam Segar'),COALESCE((SELECT setting_value FROM app_settings WHERE setting_key=?),(SELECT name FROM users WHERE id=?)),COALESCE((SELECT setting_value FROM app_settings WHERE setting_key=?),'Pemilik')`, ownerSettingKey(ownerID, "business_name"), ownerSettingKey(ownerID, "signer_name"), ownerID, ownerSettingKey(ownerID, "signer_title")).Scan(&s.BusinessName, &s.SignerName, &s.SignerTitle)
	var actualOwner, propertyID, tenantID int64
	switch kind {
	case "invoice":
		var issue, due time.Time
		var serviceStart, serviceEnd sql.NullTime
		var room, notes sql.NullString
		err := tx.QueryRowContext(ctx, `SELECT p.owner_id,i.property_id,i.tenant_id,i.invoice_number,u.name,p.name,rm.number,i.issue_date,i.due_date,i.service_start_date,i.service_end_date,i.total_amount,i.paid_amount,i.status,i.notes FROM invoices i JOIN properties p ON p.id=i.property_id JOIN users u ON u.id=i.tenant_id LEFT JOIN rooms rm ON rm.id=i.room_id WHERE i.id=? AND p.owner_id=? AND i.status<>'draft' FOR UPDATE`, sourceID, ownerID).Scan(&actualOwner, &propertyID, &tenantID, &s.SourceNumber, &s.TenantName, &s.PropertyName, &room, &issue, &due, &serviceStart, &serviceEnd, &s.Amount, &s.PaidAmount, &s.Status, &notes)
		if err != nil {
			return s, 0, 0, 0, err
		}
		s.Room = room.String
		s.IssueDate = documentDate(issue)
		s.DueDate = documentDate(due)
		if serviceStart.Valid {
			s.PeriodStart = documentDate(serviceStart.Time)
		}
		if serviceEnd.Valid {
			s.PeriodEnd = documentDate(serviceEnd.Time)
		}
		s.Outstanding = s.Amount - s.PaidAmount
		s.Notes = notes.String
		s.Title = "Invoice " + s.SourceNumber
		rows, err := tx.QueryContext(ctx, `SELECT description,CAST(quantity AS CHAR),line_total FROM invoice_lines WHERE invoice_id=? ORDER BY id`, sourceID)
		if err != nil {
			return s, 0, 0, 0, err
		}
		for rows.Next() {
			var v DocumentLine
			if err = rows.Scan(&v.Description, &v.Quantity, &v.Amount); err != nil {
				rows.Close()
				return s, 0, 0, 0, err
			}
			s.Lines = append(s.Lines, v)
		}
		rows.Close()
	case "receipt":
		var paid time.Time
		var room sql.NullString
		err := tx.QueryRowContext(ctx, `SELECT p.owner_id,p.id,pay.tenant_id,pay.receipt_number,u.name,p.name,rm.number,pay.payment_date,pay.amount,'verified' FROM payments pay JOIN users u ON u.id=pay.tenant_id JOIN payment_allocations pa ON pa.payment_id=pay.id JOIN invoices i ON i.id=pa.invoice_id JOIN properties p ON p.id=i.property_id LEFT JOIN rooms rm ON rm.id=i.room_id WHERE pay.id=? AND p.owner_id=? ORDER BY pa.id LIMIT 1 FOR UPDATE`, sourceID, ownerID).Scan(&actualOwner, &propertyID, &tenantID, &s.SourceNumber, &s.TenantName, &s.PropertyName, &room, &paid, &s.Amount, &s.Status)
		if err != nil {
			return s, 0, 0, 0, err
		}
		s.Room = room.String
		s.IssueDate = documentDate(paid)
		s.PaidAmount = s.Amount
		s.Title = "Kuitansi " + s.SourceNumber
		rows, err := tx.QueryContext(ctx, `SELECT CONCAT('Pembayaran ',i.invoice_number),'1',pa.amount FROM payment_allocations pa JOIN invoices i ON i.id=pa.invoice_id WHERE pa.payment_id=? ORDER BY pa.id`, sourceID)
		if err != nil {
			return s, 0, 0, 0, err
		}
		for rows.Next() {
			var v DocumentLine
			if err = rows.Scan(&v.Description, &v.Quantity, &v.Amount); err != nil {
				rows.Close()
				return s, 0, 0, 0, err
			}
			s.Lines = append(s.Lines, v)
		}
		rows.Close()
	case "lease":
		var start, end time.Time
		var cycle string
		err := tx.QueryRowContext(ctx, `SELECT p.owner_id,l.property_id,l.tenant_id,l.lease_number,u.name,p.name,rm.number,l.start_date,l.end_date,l.agreed_rent,l.status,l.billing_cycle FROM leases l JOIN users u ON u.id=l.tenant_id JOIN properties p ON p.id=l.property_id JOIN rooms rm ON rm.id=l.room_id WHERE l.id=? AND p.owner_id=? FOR UPDATE`, sourceID, ownerID).Scan(&actualOwner, &propertyID, &tenantID, &s.SourceNumber, &s.TenantName, &s.PropertyName, &s.Room, &start, &end, &s.Amount, &s.Status, &cycle)
		if err != nil {
			return s, 0, 0, 0, err
		}
		s.IssueDate = documentDate(businessToday(time.UTC))
		s.PeriodStart = documentDate(start)
		s.PeriodEnd = documentDate(end)
		s.Title = "Perjanjian Sewa " + s.SourceNumber
		s.Lines = []DocumentLine{{Description: "Sewa " + map[string]string{"monthly": "bulanan", "annual": "tahunan"}[cycle], Quantity: "1", Amount: s.Amount}}
	case "inspection":
		var inspected time.Time
		var inspectionType string
		var notes sql.NullString
		err := tx.QueryRowContext(ctx, `SELECT p.owner_id,l.property_id,l.tenant_id,CONCAT(l.lease_number,'-',IF(ins.inspection_type='check_in','MASUK','KELUAR')),u.name,p.name,rm.number,ins.inspected_at,ins.inspection_type,ins.overall_notes,l.status FROM inspections ins JOIN leases l ON l.id=ins.lease_id JOIN users u ON u.id=l.tenant_id JOIN properties p ON p.id=l.property_id JOIN rooms rm ON rm.id=l.room_id WHERE ins.id=? AND p.owner_id=? FOR UPDATE`, sourceID, ownerID).Scan(&actualOwner, &propertyID, &tenantID, &s.SourceNumber, &s.TenantName, &s.PropertyName, &s.Room, &inspected, &inspectionType, &notes, &s.Status)
		if err != nil {
			return s, 0, 0, 0, err
		}
		s.IssueDate = inspected.In(a.location).Format("02 Jan 2006 15:04 WITA")
		s.Notes = notes.String
		if inspectionType == "check_in" {
			s.Title = "Berita Acara Serah Terima Masuk"
		} else {
			s.Title = "Berita Acara Serah Terima Keluar"
		}
		rows, err := tx.QueryContext(ctx, `SELECT item_name_snapshot,CAST(quantity AS CHAR),0 FROM inspection_items WHERE inspection_id=? ORDER BY id`, sourceID)
		if err != nil {
			return s, 0, 0, 0, err
		}
		for rows.Next() {
			var v DocumentLine
			if err = rows.Scan(&v.Description, &v.Quantity, &v.Amount); err != nil {
				rows.Close()
				return s, 0, 0, 0, err
			}
			s.Lines = append(s.Lines, v)
		}
		rows.Close()
	default:
		return s, 0, 0, 0, sql.ErrNoRows
	}
	return s, actualOwner, propertyID, tenantID, nil
}

func documentDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("02 Jan 2006")
}
func documentNeedsTenant(kind string) bool { return kind == "lease" || kind == "inspection" }

func (a *App) ownerSignDocument(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "owner")
	if s == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(s, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	id, e := parseID(r.PathValue("id"))
	if e != nil {
		http.NotFound(w, r)
		return
	}
	var passwordHash string
	if err := a.db.QueryRowContext(r.Context(), `SELECT password_hash FROM users WHERE id=?`, s.User.ID).Scan(&passwordHash); err != nil {
		a.serverError(w, r, err)
		return
	}
	if !auth.VerifyPassword(passwordHash, r.FormValue("password")) {
		http.Error(w, "Kata sandi owner tidak cocok.", http.StatusUnauthorized)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var tenantID int64
	var kind, hash, status string
	err = tx.QueryRowContext(r.Context(), `SELECT tenant_id,document_type,snapshot_sha256,status FROM digital_documents WHERE id=? AND owner_id=? FOR UPDATE`, id, s.User.ID).Scan(&tenantID, &kind, &hash, &status)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if status != "issued" {
		http.Error(w, "Dokumen sudah ditandatangani atau dibatalkan.", http.StatusConflict)
		return
	}
	signedAt := time.Now().UTC().Truncate(time.Microsecond)
	signature := documentSignature(a.cfg.HashKey, hash, s.User.ID, signedAt)
	next := "completed"
	if documentNeedsTenant(kind) {
		next = "awaiting_tenant"
	}
	_, err = tx.ExecContext(r.Context(), `UPDATE digital_documents SET status=?,owner_signed_by=?,owner_signed_at=?,owner_signature_sha256=? WHERE id=?`, next, s.User.ID, signedAt, signature, id)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	title := "Dokumen telah ditandatangani pemilik"
	body := "Dokumen digital tersedia untuk dilihat."
	if next == "awaiting_tenant" {
		title = "Dokumen menunggu tanda tangan Anda"
		body = "Baca dokumen lalu tanda tangani menggunakan kata sandi akun."
	}
	if err = notifyUserTx(r.Context(), tx, tenantID, fmt.Sprintf("document.owner_signed.%d", id), title, body, "/penghuni/dokumen/"+strconv.FormatInt(id, 10)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = insertAuditTx(r.Context(), tx, s.User.ID, "document.owner_sign", "digital_document", id, map[string]any{"status": status}, map[string]any{"status": next, "signature": signature}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/dokumen/"+strconv.FormatInt(id, 10)+"?berhasil=ttd", http.StatusSeeOther)
}

func (a *App) tenantSignDocument(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "tenant")
	if s == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(s, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	id, e := parseID(r.PathValue("id"))
	if e != nil {
		http.NotFound(w, r)
		return
	}
	var passwordHash string
	if err := a.db.QueryRowContext(r.Context(), `SELECT password_hash FROM users WHERE id=?`, s.User.ID).Scan(&passwordHash); err != nil {
		a.serverError(w, r, err)
		return
	}
	if !auth.VerifyPassword(passwordHash, r.FormValue("password")) {
		http.Error(w, "Kata sandi tidak cocok.", http.StatusUnauthorized)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var propertyID int64
	var hash, status string
	err = tx.QueryRowContext(r.Context(), `SELECT property_id,snapshot_sha256,status FROM digital_documents WHERE id=? AND tenant_id=? FOR UPDATE`, id, s.User.ID).Scan(&propertyID, &hash, &status)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if status != "awaiting_tenant" {
		http.Error(w, "Dokumen tidak sedang menunggu tanda tangan Anda.", http.StatusConflict)
		return
	}
	signedAt := time.Now().UTC().Truncate(time.Microsecond)
	signature := documentSignature(a.cfg.HashKey, hash, s.User.ID, signedAt)
	_, err = tx.ExecContext(r.Context(), `UPDATE digital_documents SET status='completed',tenant_signed_by=?,tenant_signed_at=?,tenant_signature_sha256=? WHERE id=?`, s.User.ID, signedAt, signature, id)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = notifyOwnerForPropertyTx(r.Context(), tx, propertyID, fmt.Sprintf("document.tenant_signed.%d", id), "Dokumen ditandatangani penghuni", s.User.Name+" telah menandatangani dokumen digital.", "/pemilik/dokumen/"+strconv.FormatInt(id, 10)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = insertAuditTx(r.Context(), tx, s.User.ID, "document.tenant_sign", "digital_document", id, map[string]any{"status": status}, map[string]any{"status": "completed", "signature": signature}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/penghuni/dokumen/"+strconv.FormatInt(id, 10)+"?berhasil=ttd", http.StatusSeeOther)
}

func documentSignature(secret, documentHash string, userID int64, at time.Time) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(documentHash + "|" + strconv.FormatInt(userID, 10) + "|" + at.Format(time.RFC3339Nano)))
	return hex.EncodeToString(mac.Sum(nil))
}

func (a *App) ownerVoidDocument(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "owner")
	if s == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(s, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	id, e := parseID(r.PathValue("id"))
	reason := strings.TrimSpace(r.FormValue("reason"))
	if e != nil || len(reason) < 5 || len(reason) > 1000 {
		http.Error(w, "Alasan pembatalan wajib diisi.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var tenantID int64
	var old string
	err = tx.QueryRowContext(r.Context(), `SELECT tenant_id,status FROM digital_documents WHERE id=? AND owner_id=? FOR UPDATE`, id, s.User.ID).Scan(&tenantID, &old)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if old == "void" {
		http.Error(w, "Dokumen sudah dibatalkan.", http.StatusConflict)
		return
	}
	_, err = tx.ExecContext(r.Context(), `UPDATE digital_documents SET status='void',voided_by=?,voided_at=UTC_TIMESTAMP(6),void_reason=? WHERE id=?`, s.User.ID, reason, id)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = notifyUserTx(r.Context(), tx, tenantID, fmt.Sprintf("document.void.%d", id), "Dokumen dibatalkan", reason, "/penghuni/dokumen/"+strconv.FormatInt(id, 10)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = insertAuditTx(r.Context(), tx, s.User.ID, "document.void", "digital_document", id, map[string]any{"status": old}, map[string]any{"status": "void", "reason": reason}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/dokumen?berhasil=batal", http.StatusSeeOther)
}

func (a *App) ownerDocumentDetail(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "owner")
	if s == nil {
		return
	}
	id, e := parseID(r.PathValue("id"))
	if e != nil {
		http.NotFound(w, r)
		return
	}
	page, err := a.loadDocumentDetail(r.Context(), id, "owner_id", s.User.ID)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	page.IsOwner = true
	a.render(w, http.StatusOK, "document_detail", PageData{Title: page.Document.Title, User: s.User, CSRFToken: s.CSRFToken, Payload: page, Success: documentSuccess(r)})
}
func (a *App) tenantDocumentDetail(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "tenant")
	if s == nil {
		return
	}
	id, e := parseID(r.PathValue("id"))
	if e != nil {
		http.NotFound(w, r)
		return
	}
	page, err := a.loadDocumentDetail(r.Context(), id, "tenant_id", s.User.ID)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	a.render(w, http.StatusOK, "document_detail", PageData{Title: page.Document.Title, User: s.User, CSRFToken: s.CSRFToken, Payload: page, Success: documentSuccess(r)})
}
func documentSuccess(r *http.Request) string {
	if r.URL.Query().Get("berhasil") != "" {
		return "Tanda tangan digital berhasil direkam dan dokumen dikunci."
	}
	return ""
}

func (a *App) loadDocumentDetail(ctx context.Context, id int64, scope string, userID int64) (DocumentDetailPage, error) {
	var page DocumentDetailPage
	query := documentListSQL + ` WHERE d.id=? AND d.` + scope + `=?`
	v, err := scanDocumentView(a.db.QueryRowContext(ctx, query, id, userID))
	if err != nil {
		return page, err
	}
	page.Document = v
	var raw string
	if err = a.db.QueryRowContext(ctx, `SELECT snapshot_json FROM digital_documents WHERE id=?`, id).Scan(&raw); err != nil {
		return page, err
	}
	if err = json.Unmarshal([]byte(raw), &page.Snapshot); err != nil {
		return page, err
	}
	page.VerifyURL = strings.TrimRight(a.cfg.BaseURL, "/") + "/verifikasi-dokumen/" + v.VerificationCode
	page.IntegrityOK, _ = a.documentIntegrity(ctx, id)
	return page, nil
}

func (a *App) verifyDocument(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(r.PathValue("code"))
	page := DocumentVerifyPage{}
	if len(code) == 32 {
		var raw, hash string
		v, err := scanDocumentView(a.db.QueryRowContext(r.Context(), documentListSQL+` WHERE d.verification_code=?`, code))
		if err == nil {
			page.Found = true
			page.Document = v
			if e := a.db.QueryRowContext(r.Context(), `SELECT snapshot_json,snapshot_sha256 FROM digital_documents WHERE id=?`, v.ID).Scan(&raw, &hash); e == nil {
				var snapshot DigitalDocumentSnapshot
				_ = json.Unmarshal([]byte(raw), &snapshot)
				page.BusinessName = snapshot.BusinessName
				if len(hash) >= 12 {
					page.HashShort = hash[:12]
				}
				page.IntegrityOK, _ = a.documentIntegrity(r.Context(), v.ID)
			}
		}
	}
	w.Header().Set("Cache-Control", "no-store")
	a.render(w, http.StatusOK, "document_verify", PageData{Title: "Verifikasi Dokumen — Alam Segar", Payload: page})
}

func (a *App) documentIntegrity(ctx context.Context, id int64) (bool, error) {
	var raw, storedHash string
	var ownerID, tenantID sql.NullInt64
	var ownerAt, tenantAt sql.NullTime
	var ownerSignature, tenantSignature sql.NullString
	if err := a.db.QueryRowContext(ctx, `SELECT snapshot_json,snapshot_sha256,owner_signed_by,owner_signed_at,owner_signature_sha256,tenant_signed_by,tenant_signed_at,tenant_signature_sha256 FROM digital_documents WHERE id=?`, id).Scan(&raw, &storedHash, &ownerID, &ownerAt, &ownerSignature, &tenantID, &tenantAt, &tenantSignature); err != nil {
		return false, err
	}
	sum := sha256.Sum256([]byte(raw))
	if !hmac.Equal([]byte(storedHash), []byte(hex.EncodeToString(sum[:]))) {
		return false, nil
	}
	if ownerID.Valid || ownerAt.Valid || ownerSignature.Valid {
		if !ownerID.Valid || !ownerAt.Valid || !ownerSignature.Valid {
			return false, nil
		}
		expected := documentSignature(a.cfg.HashKey, storedHash, ownerID.Int64, ownerAt.Time.UTC())
		if !hmac.Equal([]byte(expected), []byte(ownerSignature.String)) {
			return false, nil
		}
	}
	if tenantID.Valid || tenantAt.Valid || tenantSignature.Valid {
		if !tenantID.Valid || !tenantAt.Valid || !tenantSignature.Valid {
			return false, nil
		}
		expected := documentSignature(a.cfg.HashKey, storedHash, tenantID.Int64, tenantAt.Time.UTC())
		if !hmac.Equal([]byte(expected), []byte(tenantSignature.String)) {
			return false, nil
		}
	}
	return true, nil
}

func documentTypeLabel(value string) string {
	if label, ok := map[string]string{"invoice": "Invoice", "receipt": "Kuitansi", "lease": "Kontrak Sewa", "inspection": "Serah Terima"}[value]; ok {
		return label
	}
	return value
}
