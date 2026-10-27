package httpapp

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"alam-segar/kos-management/internal/auth"
	"alam-segar/kos-management/internal/domain"
)

type LeaseListItem struct {
	ID, TenantID, RoomID                                        int64
	Number, TenantName, PropertyName, RoomNumber, Cycle, Status string
	StartDate, EndDate                                          time.Time
	AgreedRent                                                  int64
}

type LeaseDetailView struct {
	LeaseListItem
	DueDay, AnchorDay, InstallmentCount  int
	PartialAllowed                       bool
	SpecialTerms                         string
	ActualCheckIn, ActualCheckOut        sql.NullTime
	Invoices                             []InvoiceView
	Versions                             []LeaseVersionView
	Inspections                          []InspectionView
	EstimatedFuture, EstimatedRefundable int64
	Termination                          *LeaseTerminationView
}

type LeaseVersionView struct {
	Version                         int
	LeaseNumber, ChangeType, Reason string
	EffectiveDate                   time.Time
}

type InspectionView struct {
	ID                       int64
	LeaseNumber, Type, Notes string
	InspectedAt              time.Time
}

type LeaseTerminationView struct {
	ID, LeaseID, CancelledUnpaid, SettledPaid int64
	Number, LeaseNumber, Decision, Reason     string
	EffectiveDate                             time.Time
	HasProof                                  bool
}

type LeaseOption struct {
	ID    int64
	Label string
}

type LeasePage struct {
	Leases                   []LeaseListItem
	Detail                   LeaseDetailView
	Tenants, Rooms, Bookings []LeaseOption
	Accounts                 []FinanceAccountView
	Form                     map[string]string
}

type InvoiceView struct {
	ID                                                         int64
	Number, TenantName, PropertyName, RoomNumber, Type, Status string
	IssueDate, DueDate                                         time.Time
	Total, Paid, CreditAvailable                               int64
}

func (a *App) ownerLeaseList(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	page := LeasePage{}
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT l.id,l.lease_number,l.tenant_id,l.room_id,u.name,p.name,rm.number,l.billing_cycle,l.start_date,l.end_date,l.agreed_rent,l.status
		FROM leases l JOIN users u ON u.id=l.tenant_id JOIN properties p ON p.id=l.property_id JOIN rooms rm ON rm.id=l.room_id
		WHERE p.owner_id=? ORDER BY FIELD(l.status,'active','draft','expiring','ended','terminated'),l.start_date DESC`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item LeaseListItem
		if err := rows.Scan(&item.ID, &item.Number, &item.TenantID, &item.RoomID, &item.TenantName, &item.PropertyName, &item.RoomNumber, &item.Cycle, &item.StartDate, &item.EndDate, &item.AgreedRent, &item.Status); err != nil {
			a.serverError(w, r, err)
			return
		}
		page.Leases = append(page.Leases, item)
	}
	data := PageData{Title: "Kontrak — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page}
	if r.URL.Query().Get("berhasil") != "" {
		data.Success = "Kontrak berhasil diperbarui."
	}
	a.render(w, http.StatusOK, "owner_leases", data)
}

func (a *App) ownerLeaseCreateForm(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	form := map[string]string{"billing_cycle": "monthly", "installment_count": "1", "partial_allowed": "0"}
	if value := r.URL.Query().Get("booking_id"); value != "" {
		form["booking_id"] = value
		var tenantID, roomID, monthlyRate int64
		var startDate, endDate time.Time
		err := a.db.QueryRowContext(r.Context(), `
			SELECT COALESCE(b.applicant_user_id,0),b.room_id,b.planned_start_date,b.planned_end_date,rm.monthly_rate
			FROM booking_requests b JOIN properties p ON p.id=b.property_id JOIN rooms rm ON rm.id=b.room_id
			JOIN users u ON u.id=b.applicant_user_id AND u.role='tenant' AND u.status='active' AND u.email_verified_at IS NOT NULL
			WHERE b.id=? AND p.owner_id=? AND b.status='confirmed'`, value, session.User.ID).
			Scan(&tenantID, &roomID, &startDate, &endDate, &monthlyRate)
		if err == nil {
			form["tenant_id"] = strconv.FormatInt(tenantID, 10)
			form["room_id"] = strconv.FormatInt(roomID, 10)
			form["start_date"] = startDate.Format("2006-01-02")
			form["end_date"] = endDate.Format("2006-01-02")
			form["anchor_day"] = strconv.Itoa(startDate.Day())
			form["due_day"] = strconv.Itoa(startDate.Day())
			form["agreed_rent"] = strconv.FormatInt(monthlyRate, 10)
		}
	}
	a.renderLeaseForm(w, r, http.StatusOK, session, form, "")
}

func (a *App) ownerLeaseCreate(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	fields := []string{"tenant_id", "room_id", "booking_id", "billing_cycle", "start_date", "end_date", "anchor_day", "due_day", "agreed_rent", "installment_count", "partial_allowed", "special_terms"}
	form := map[string]string{}
	for _, key := range fields {
		form[key] = strings.TrimSpace(r.FormValue(key))
	}
	tenantID, e1 := parseID(form["tenant_id"])
	roomID, e2 := parseID(form["room_id"])
	bookingID := int64(0)
	if form["booking_id"] != "" {
		bookingID, _ = strconv.ParseInt(form["booking_id"], 10, 64)
	}
	start, e3 := parseBusinessDate(form["start_date"])
	end, e4 := parseBusinessDate(form["end_date"])
	anchor, _ := strconv.Atoi(form["anchor_day"])
	due, _ := strconv.Atoi(form["due_day"])
	rent, _ := strconv.ParseInt(form["agreed_rent"], 10, 64)
	installments, _ := strconv.Atoi(form["installment_count"])
	partial := form["partial_allowed"] == "1"
	cycle := form["billing_cycle"]
	charges, revenue, planErr := domain.BuildLeasePlan(start, end, anchor, due, cycle, rent, installments)
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil || planErr != nil || len(form["special_terms"]) > 4000 {
		message := "Data kontrak tidak valid."
		if planErr != nil {
			message = planErr.Error()
		}
		a.renderLeaseForm(w, r, http.StatusUnprocessableEntity, session, form, message)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var propertyID int64
	var condition string
	err = tx.QueryRowContext(r.Context(), `SELECT rm.property_id,rm.condition_status FROM rooms rm JOIN properties p ON p.id=rm.property_id WHERE rm.id=? AND p.owner_id=? AND rm.deleted_at IS NULL FOR UPDATE`, roomID, session.User.ID).Scan(&propertyID, &condition)
	if err != nil || condition != "ready" {
		a.renderLeaseForm(w, r, http.StatusConflict, session, form, "Kamar tidak ditemukan atau belum siap disewakan.")
		return
	}
	var tenantStatus, tenantRole string
	var tenantVerified sql.NullTime
	if err := tx.QueryRowContext(r.Context(), `SELECT status,role,email_verified_at FROM users WHERE id=? AND deleted_at IS NULL FOR UPDATE`, tenantID).Scan(&tenantStatus, &tenantRole, &tenantVerified); err != nil || tenantStatus != "active" || tenantRole != "tenant" || !tenantVerified.Valid {
		a.renderLeaseForm(w, r, http.StatusConflict, session, form, "Penghuni belum aktif, belum memverifikasi email, atau tidak ditemukan.")
		return
	}
	var conflicts int
	err = tx.QueryRowContext(r.Context(), `SELECT (SELECT COUNT(*) FROM leases WHERE room_id=? AND status IN ('draft','active','expiring') AND start_date<? AND end_date>?) + (SELECT COUNT(*) FROM leases WHERE tenant_id=? AND status IN ('draft','active','expiring') AND start_date<? AND end_date>?)`, roomID, end, start, tenantID, end, start).Scan(&conflicts)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if conflicts > 0 {
		a.renderLeaseForm(w, r, http.StatusConflict, session, form, "Kamar atau penghuni memiliki kontrak yang bertabrakan.")
		return
	}
	if bookingID > 0 {
		var bookingRoom, bookingTenant int64
		var bookingStatus string
		err = tx.QueryRowContext(r.Context(), `SELECT room_id,COALESCE(applicant_user_id,0),status FROM booking_requests WHERE id=? AND property_id=? FOR UPDATE`, bookingID, propertyID).Scan(&bookingRoom, &bookingTenant, &bookingStatus)
		if err != nil || bookingRoom != roomID || bookingTenant != tenantID || bookingStatus != "confirmed" {
			a.renderLeaseForm(w, r, http.StatusConflict, session, form, "Booking harus terkonfirmasi, memiliki akun penghuni aktif, dan sesuai kamar sebelum dikonversi.")
			return
		}
	}
	number, _ := referenceNumber("KTR")
	result, err := tx.ExecContext(r.Context(), `INSERT INTO leases (lease_number,property_id,room_id,tenant_id,booking_request_id,billing_cycle,start_date,end_date,original_anchor_day,due_day,agreed_rent,partial_payment_allowed,installment_count,special_terms,status,created_by) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?, 'draft',?)`, number, propertyID, roomID, tenantID, nullablePositive(bookingID), cycle, start, end, anchor, due, rent, partial, installments, nullableString(form["special_terms"]), session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	leaseID, _ := result.LastInsertId()
	after, _ := json.Marshal(map[string]any{"number": number, "room_id": roomID, "tenant_id": tenantID, "start": form["start_date"], "end": form["end_date"], "cycle": cycle, "rent": rent, "installments": installments})
	_, err = tx.ExecContext(r.Context(), `INSERT INTO lease_versions (lease_id,version_number,effective_date,change_type,after_json,reason,created_by) VALUES (?,1,?,'created',?,'Kontrak dibuat',?)`, leaseID, start, string(after), session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	invoiceIDs, err := a.createLeaseInvoicesTx(r.Context(), tx, session.User.ID, leaseID, propertyID, roomID, tenantID, number, charges, revenue)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if bookingID > 0 {
		if err := a.allocateBookingFundTx(r.Context(), tx, session.User.ID, bookingID, leaseID, propertyID, tenantID, invoiceIDs[0]); err != nil {
			a.serverError(w, r, err)
			return
		}
		_, err = tx.ExecContext(r.Context(), `UPDATE booking_requests SET status='converted_to_contract' WHERE id=?`, bookingID)
		if err != nil {
			a.serverError(w, r, err)
			return
		}
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "lease.create", "lease", leaseID, nil, map[string]any{"number": number, "status": "draft"}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := createNotificationTx(r.Context(), tx, tenantID, fmt.Sprintf("lease.created.%d", leaseID), "Kontrak baru", "Kontrak "+number+" telah dibuat. Silakan periksa tagihan Anda.", "/penghuni/kontrak"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/kontrak/"+strconv.FormatInt(leaseID, 10)+"?berhasil=dibuat", http.StatusSeeOther)
}

func (a *App) createLeaseInvoicesTx(ctx context.Context, tx *sql.Tx, ownerID, leaseID, propertyID, roomID, tenantID int64, leaseNumber string, charges []domain.LeaseCharge, revenue []domain.RevenuePeriod) ([]int64, error) {
	arID, err := systemAccountID(ctx, tx, ownerID, "accounts_receivable")
	if err != nil {
		return nil, err
	}
	unearnedID, err := systemAccountID(ctx, tx, ownerID, "unearned_rent")
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(charges))
	for _, charge := range charges {
		number, _ := referenceNumber("INV")
		result, err := tx.ExecContext(ctx, `INSERT INTO invoices (invoice_number,property_id,room_id,tenant_id,lease_id,invoice_type,service_start_date,service_end_date,issue_date,due_date,total_amount,status,issued_by,notes) VALUES (?,?,?,?,?,'rent',?,?,?,?,?,'issued',?,?)`, number, propertyID, roomID, tenantID, leaseID, charge.ServiceStart, charge.ServiceEnd, charge.ServiceStart, charge.DueDate, charge.Amount, ownerID, "Tagihan kontrak "+leaseNumber)
		if err != nil {
			return nil, err
		}
		invoiceID, _ := result.LastInsertId()
		ids = append(ids, invoiceID)
		_, err = tx.ExecContext(ctx, `INSERT INTO invoice_lines (invoice_id,description,quantity,unit_amount,line_total) VALUES (?,'Sewa kamar',1,?,?)`, invoiceID, charge.Amount, charge.Amount)
		if err != nil {
			return nil, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO lease_installments (lease_id,installment_number,due_date,amount,invoice_id) VALUES (?,?,?,?,?)`, leaseID, charge.Number, charge.DueDate, charge.Amount, invoiceID)
		if err != nil {
			return nil, err
		}
		_, err = postJournalTx(ctx, tx, JournalInput{OwnerID: ownerID, PropertyID: sql.NullInt64{Int64: propertyID, Valid: true}, Date: charge.ServiceStart, Description: "Penerbitan " + number, SourceType: "invoice", SourceID: invoiceID, EventType: "issued", ActorID: ownerID, Lines: []JournalLineInput{{AccountID: arID, Debit: charge.Amount, Description: "Piutang sewa"}, {AccountID: unearnedID, Credit: charge.Amount, Description: "Pendapatan diterima di muka"}}})
		if err != nil {
			return nil, err
		}
	}
	for _, period := range revenue {
		_, err := tx.ExecContext(ctx, `INSERT INTO revenue_recognition_schedules (lease_id,property_id,recognition_date,service_start_date,service_end_date,amount) VALUES (?,?,?,?,?,?)`, leaseID, propertyID, period.RecognitionDate, period.ServiceStart, period.ServiceEnd, period.Amount)
		if err != nil {
			return nil, err
		}
	}
	return ids, nil
}

func (a *App) allocateBookingFundTx(ctx context.Context, tx *sql.Tx, ownerID, bookingID, leaseID, propertyID, tenantID, firstInvoiceID int64) error {
	var fundID, amount int64
	var usage string
	err := tx.QueryRowContext(ctx, `SELECT bf.id,bf.remaining_liability,b.booking_usage FROM booking_funds bf JOIN booking_requests b ON b.id=bf.booking_request_id WHERE bf.booking_request_id=? AND bf.status='held' FOR UPDATE`, bookingID).Scan(&fundID, &amount, &usage)
	if err != nil {
		return err
	}
	liabilityID, err := systemAccountID(ctx, tx, ownerID, "booking_liability")
	if err != nil {
		return err
	}
	if usage == "additional_booking_fee" {
		revenueID, err := systemAccountID(ctx, tx, ownerID, "booking_revenue")
		if err != nil {
			return err
		}
		_, err = postJournalTx(ctx, tx, JournalInput{OwnerID: ownerID, PropertyID: sql.NullInt64{Int64: propertyID, Valid: true}, Date: businessToday(a.location), Description: "Pengakuan biaya booking", SourceType: "booking_fund", SourceID: fundID, EventType: "fee_recognized", ActorID: ownerID, Lines: []JournalLineInput{{AccountID: liabilityID, Debit: amount}, {AccountID: revenueID, Credit: amount}}})
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO booking_fund_allocations (booking_fund_id,lease_id,amount,allocation_type) VALUES (?,?,?,'booking_fee')`, fundID, leaseID, amount)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE booking_funds SET remaining_liability=0,status='fee_recognized' WHERE id=?`, fundID)
		return err
	}
	var total, paid int64
	if err := tx.QueryRowContext(ctx, `SELECT total_amount,paid_amount FROM invoices WHERE id=? FOR UPDATE`, firstInvoiceID).Scan(&total, &paid); err != nil {
		return err
	}
	outstanding := total - paid
	apply := amount
	if apply > outstanding {
		apply = outstanding
	}
	excess := amount - apply
	arID, err := systemAccountID(ctx, tx, ownerID, "accounts_receivable")
	if err != nil {
		return err
	}
	lines := []JournalLineInput{{AccountID: liabilityID, Debit: amount}}
	if apply > 0 {
		lines = append(lines, JournalLineInput{AccountID: arID, Credit: apply})
	}
	if excess > 0 {
		creditID, err := systemAccountID(ctx, tx, ownerID, "tenant_credit")
		if err != nil {
			return err
		}
		lines = append(lines, JournalLineInput{AccountID: creditID, Credit: excess})
	}
	_, err = postJournalTx(ctx, tx, JournalInput{OwnerID: ownerID, PropertyID: sql.NullInt64{Int64: propertyID, Valid: true}, Date: businessToday(a.location), Description: "Alokasi uang booking ke kontrak", SourceType: "booking_fund", SourceID: fundID, EventType: "allocated", ActorID: ownerID, Lines: lines})
	if err != nil {
		return err
	}
	if apply > 0 {
		status := "partially_paid"
		if paid+apply == total {
			status = "paid"
		}
		_, err = tx.ExecContext(ctx, `UPDATE invoices SET paid_amount=paid_amount+?,status=? WHERE id=?`, apply, status, firstInvoiceID)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO booking_fund_allocations (booking_fund_id,lease_id,invoice_id,amount,allocation_type) VALUES (?,?,?,?,'invoice_credit')`, fundID, leaseID, firstInvoiceID, apply)
		if err != nil {
			return err
		}
	}
	if excess > 0 {
		_, err = tx.ExecContext(ctx, `INSERT INTO tenant_credits (tenant_id,property_id,source_type,source_id,amount,remaining_amount) VALUES (?,?,'booking_fund',?,?,?)`, tenantID, propertyID, fundID, excess, excess)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO booking_fund_allocations (booking_fund_id,lease_id,amount,allocation_type) VALUES (?,?,?,'tenant_credit')`, fundID, leaseID, excess)
		if err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `UPDATE booking_funds SET remaining_liability=0,status='allocated' WHERE id=?`, fundID)
	return err
}

func (a *App) renderLeaseForm(w http.ResponseWriter, r *http.Request, status int, session *auth.Session, form map[string]string, message string) {
	page := LeasePage{Form: form}
	rows, err := a.db.QueryContext(r.Context(), `SELECT id,name FROM users WHERE role='tenant' AND status='active' AND email_verified_at IS NOT NULL AND deleted_at IS NULL ORDER BY name`)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var o LeaseOption
		_ = rows.Scan(&o.ID, &o.Label)
		page.Tenants = append(page.Tenants, o)
	}
	rows.Close()
	rows, err = a.db.QueryContext(r.Context(), `SELECT rm.id,CONCAT(p.name,' / Kamar ',rm.number) FROM rooms rm JOIN properties p ON p.id=rm.property_id WHERE p.owner_id=? AND rm.condition_status='ready' AND rm.deleted_at IS NULL ORDER BY p.name,rm.number`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var o LeaseOption
		_ = rows.Scan(&o.ID, &o.Label)
		page.Rooms = append(page.Rooms, o)
	}
	rows.Close()
	rows, err = a.db.QueryContext(r.Context(), `SELECT b.id,CONCAT(b.reference_number,' — ',b.applicant_name,' / Kamar ',rm.number) FROM booking_requests b JOIN rooms rm ON rm.id=b.room_id JOIN properties p ON p.id=b.property_id WHERE p.owner_id=? AND b.status='confirmed' ORDER BY b.created_at`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var o LeaseOption
		_ = rows.Scan(&o.ID, &o.Label)
		page.Bookings = append(page.Bookings, o)
	}
	rows.Close()
	a.render(w, status, "owner_lease_form", PageData{Title: "Buat Kontrak — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Error: message, Payload: page})
}

func (a *App) ownerLeaseDetail(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	page := LeasePage{}
	finance, financeErr := a.loadFinancePage(r.Context(), session.User.ID)
	if financeErr != nil {
		a.serverError(w, r, financeErr)
		return
	}
	page.Accounts = finance.Accounts
	d := &page.Detail
	err = a.db.QueryRowContext(r.Context(), `SELECT l.id,l.lease_number,l.tenant_id,l.room_id,u.name,p.name,rm.number,l.billing_cycle,l.start_date,l.end_date,l.agreed_rent,l.status,l.due_day,l.original_anchor_day,l.installment_count,l.partial_payment_allowed,COALESCE(l.special_terms,''),l.actual_check_in_at,l.actual_check_out_at FROM leases l JOIN users u ON u.id=l.tenant_id JOIN properties p ON p.id=l.property_id JOIN rooms rm ON rm.id=l.room_id WHERE l.id=? AND p.owner_id=?`, id, session.User.ID).Scan(&d.ID, &d.Number, &d.TenantID, &d.RoomID, &d.TenantName, &d.PropertyName, &d.RoomNumber, &d.Cycle, &d.StartDate, &d.EndDate, &d.AgreedRent, &d.Status, &d.DueDay, &d.AnchorDay, &d.InstallmentCount, &d.PartialAllowed, &d.SpecialTerms, &d.ActualCheckIn, &d.ActualCheckOut)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	rows, err := a.db.QueryContext(r.Context(), `SELECT id,invoice_number,'', '', '',invoice_type,status,issue_date,due_date,total_amount,paid_amount FROM invoices WHERE lease_id=? ORDER BY due_date`, id)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var v InvoiceView
		_ = rows.Scan(&v.ID, &v.Number, &v.TenantName, &v.PropertyName, &v.RoomNumber, &v.Type, &v.Status, &v.IssueDate, &v.DueDate, &v.Total, &v.Paid)
		d.Invoices = append(d.Invoices, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(r.Context(), `SELECT version_number,change_type,reason,effective_date FROM lease_versions WHERE lease_id=? ORDER BY version_number DESC`, id)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var v LeaseVersionView
		_ = rows.Scan(&v.Version, &v.ChangeType, &v.Reason, &v.EffectiveDate)
		d.Versions = append(d.Versions, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(r.Context(), `SELECT id,inspection_type,inspected_at,COALESCE(overall_notes,'') FROM inspections WHERE lease_id=? ORDER BY inspected_at`, id)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var v InspectionView
		_ = rows.Scan(&v.ID, &v.Type, &v.InspectedAt, &v.Notes)
		d.Inspections = append(d.Inspections, v)
	}
	rows.Close()
	_ = a.db.QueryRowContext(r.Context(), `SELECT COALESCE(SUM(amount),0) FROM revenue_recognition_schedules WHERE lease_id=? AND status='scheduled' AND service_start_date>=CURRENT_DATE`, id).Scan(&d.EstimatedFuture)
	var recognized, paid int64
	_ = a.db.QueryRowContext(r.Context(), `SELECT COALESCE(SUM(amount),0) FROM revenue_recognition_schedules WHERE lease_id=? AND status='posted'`, id).Scan(&recognized)
	_ = a.db.QueryRowContext(r.Context(), `SELECT COALESCE(SUM(paid_amount),0) FROM invoices WHERE lease_id=? AND invoice_type='rent'`, id).Scan(&paid)
	d.EstimatedRefundable = paid - recognized
	if d.EstimatedRefundable < 0 {
		d.EstimatedRefundable = 0
	}
	if d.EstimatedRefundable > d.EstimatedFuture {
		d.EstimatedRefundable = d.EstimatedFuture
	}
	var termination LeaseTerminationView
	err = a.db.QueryRowContext(r.Context(), `SELECT id,lease_id,termination_number,'',settlement_decision,reason,effective_date,cancelled_unpaid_amount,settled_paid_amount,proof_storage_path IS NOT NULL FROM lease_terminations WHERE lease_id=?`, id).Scan(&termination.ID, &termination.LeaseID, &termination.Number, &termination.LeaseNumber, &termination.Decision, &termination.Reason, &termination.EffectiveDate, &termination.CancelledUnpaid, &termination.SettledPaid, &termination.HasProof)
	if err == nil {
		d.Termination = &termination
	} else if !errors.Is(err, sql.ErrNoRows) {
		a.serverError(w, r, err)
		return
	}
	data := PageData{Title: "Detail Kontrak — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page}
	if r.URL.Query().Get("berhasil") != "" {
		data.Success = "Tindakan kontrak berhasil disimpan."
	}
	a.render(w, http.StatusOK, "owner_lease_detail", data)
}

func (a *App) ownerLeaseCheckIn(w http.ResponseWriter, r *http.Request) {
	a.ownerLeaseInspection(w, r, "check_in")
}
func (a *App) ownerLeaseCheckOut(w http.ResponseWriter, r *http.Request) {
	a.ownerLeaseInspection(w, r, "check_out")
}
func (a *App) ownerLeaseInspection(w http.ResponseWriter, r *http.Request, kind string) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20)
	if err := r.ParseMultipartForm(8 << 20); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	leaseID, err := parseID(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	upload, err := a.savePrivateUpload(r, "photo", "inspections", 5<<20, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	notes := strings.TrimSpace(r.FormValue("notes"))
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var propertyID, roomID int64
	var startDate, endDate time.Time
	var status string
	err = tx.QueryRowContext(r.Context(), `SELECT l.property_id,l.room_id,l.start_date,l.end_date,l.status FROM leases l JOIN properties p ON p.id=l.property_id WHERE l.id=? AND p.owner_id=? FOR UPDATE`, leaseID, session.User.ID).Scan(&propertyID, &roomID, &startDate, &endDate, &status)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if kind == "check_in" && status != "draft" {
		http.Error(w, "Hanya kontrak draft yang dapat check-in.", http.StatusConflict)
		return
	}
	today := businessToday(a.location)
	if kind == "check_in" && today.Before(startDate) {
		http.Error(w, "Check-in belum dapat dilakukan sebelum tanggal mulai kontrak.", http.StatusConflict)
		return
	}
	if kind == "check_out" && status != "active" && status != "expiring" {
		http.Error(w, "Kontrak belum aktif atau sudah selesai.", http.StatusConflict)
		return
	}
	if kind == "check_out" && today.Before(endDate) {
		http.Error(w, "Kontrak belum mencapai tanggal selesai. Gunakan alur terminasi dini agar penyelesaian keuangannya tercatat.", http.StatusConflict)
		return
	}
	result, err := tx.ExecContext(r.Context(), `INSERT INTO inspections (lease_id,inspection_type,inspected_at,overall_notes,photo_storage_path,recorded_by) VALUES (?,?,UTC_TIMESTAMP(6),?,?,?)`, leaseID, kind, nullableString(notes), upload.RelativePath, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	inspectionID, _ := result.LastInsertId()
	rows, err := tx.QueryContext(r.Context(), `SELECT id,name,condition_status FROM assets WHERE room_id=? AND condition_status<>'disposed'`, roomID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var assetID int64
		var name, condition string
		_ = rows.Scan(&assetID, &name, &condition)
		_, err = tx.ExecContext(r.Context(), `INSERT INTO inspection_items (inspection_id,asset_id,item_name_snapshot,condition_snapshot) VALUES (?,?,?,?)`, inspectionID, assetID, name, condition)
		if err != nil {
			rows.Close()
			a.serverError(w, r, err)
			return
		}
	}
	rows.Close()
	if kind == "check_in" {
		_, err = tx.ExecContext(r.Context(), `UPDATE leases SET status='active',actual_check_in_at=UTC_TIMESTAMP(6),checked_in_by=? WHERE id=?`, session.User.ID, leaseID)
	} else {
		_, err = tx.ExecContext(r.Context(), `UPDATE leases SET status='ended',actual_check_out_at=UTC_TIMESTAMP(6),checked_out_by=? WHERE id=?`, session.User.ID, leaseID)
		if err == nil {
			_, err = tx.ExecContext(r.Context(), `UPDATE rooms SET condition_status='maintenance' WHERE id=?`, roomID)
		}
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "lease."+kind, "lease", leaseID, map[string]any{"status": status}, map[string]any{"inspection_id": inspectionID}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/kontrak/"+strconv.FormatInt(leaseID, 10)+"?berhasil="+kind, http.StatusSeeOther)
}

func (a *App) ownerLeaseTerminate(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 14<<20)
	if err := r.ParseMultipartForm(14 << 20); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	leaseID, idErr := parseID(r.PathValue("id"))
	effectiveDate, dateErr := parseBusinessDate(r.FormValue("effective_date"))
	decision := r.FormValue("settlement_decision")
	reason := strings.TrimSpace(r.FormValue("reason"))
	if idErr != nil || dateErr != nil || !map[string]bool{"refund": true, "retain": true}[decision] || len(reason) < 10 {
		http.Error(w, "Tanggal, keputusan penyelesaian, dan alasan minimal 10 karakter wajib diisi.", http.StatusUnprocessableEntity)
		return
	}
	inspectionPhoto, err := a.savePrivateUpload(r, "photo", "inspections", 5<<20, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	refundProof := savedUpload{}
	var bankID int64
	if decision == "refund" {
		bankID, idErr = parseID(r.FormValue("bank_account_id"))
		if idErr != nil {
			http.Error(w, "Rekening refund wajib dipilih.", http.StatusUnprocessableEntity)
			return
		}
		refundProof, err = a.savePrivateUpload(r, "refund_proof", "lease-termination-refunds", 5<<20, true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
	}
	today := businessToday(a.location)
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var propertyID, roomID, tenantID int64
	var number, status string
	var startDate, oldEnd time.Time
	err = tx.QueryRowContext(r.Context(), `SELECT l.property_id,l.room_id,l.tenant_id,l.lease_number,l.start_date,l.end_date,l.status FROM leases l JOIN properties p ON p.id=l.property_id WHERE l.id=? AND p.owner_id=? FOR UPDATE`, leaseID, session.User.ID).Scan(&propertyID, &roomID, &tenantID, &number, &startDate, &oldEnd, &status)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if status != "active" && status != "expiring" {
		http.Error(w, "Hanya kontrak aktif yang dapat diterminasi.", http.StatusConflict)
		return
	}
	if !effectiveDate.After(startDate) || !effectiveDate.Before(oldEnd) || effectiveDate.After(today) {
		http.Error(w, "Tanggal terminasi harus setelah tanggal mulai, sebelum tanggal akhir, dan tidak boleh di masa depan.", http.StatusUnprocessableEntity)
		return
	}
	type scheduleRow struct {
		id     int64
		start  time.Time
		amount int64
		status string
	}
	rows, err := tx.QueryContext(r.Context(), `SELECT id,service_start_date,amount,status FROM revenue_recognition_schedules WHERE lease_id=? FOR UPDATE`, leaseID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	var futureAmount, recognizedAmount, currentScheduledAmount int64
	var cancelScheduleIDs []int64
	for rows.Next() {
		var v scheduleRow
		if err := rows.Scan(&v.id, &v.start, &v.amount, &v.status); err != nil {
			rows.Close()
			a.serverError(w, r, err)
			return
		}
		if v.status == "posted" {
			recognizedAmount += v.amount
		}
		if v.status == "scheduled" && !v.start.Before(effectiveDate) {
			futureAmount += v.amount
			cancelScheduleIDs = append(cancelScheduleIDs, v.id)
		} else if v.status == "scheduled" && v.start.Before(effectiveDate) {
			currentScheduledAmount += v.amount
		}
	}
	rows.Close()
	rows, err = tx.QueryContext(r.Context(), `SELECT id,service_start_date,total_amount,paid_amount,status FROM invoices WHERE lease_id=? AND invoice_type='rent' FOR UPDATE`, leaseID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	var totalPaid int64
	var voidInvoiceIDs []int64
	for rows.Next() {
		var invoiceID, total, paid int64
		var serviceStart time.Time
		var invoiceStatus string
		if err := rows.Scan(&invoiceID, &serviceStart, &total, &paid, &invoiceStatus); err != nil {
			rows.Close()
			a.serverError(w, r, err)
			return
		}
		totalPaid += paid
		if !serviceStart.Before(effectiveDate) && paid == 0 && (invoiceStatus == "issued" || invoiceStatus == "overdue") {
			voidInvoiceIDs = append(voidInvoiceIDs, invoiceID)
		}
	}
	rows.Close()
	settledPaid := totalPaid - recognizedAmount - currentScheduledAmount
	if settledPaid < 0 {
		settledPaid = 0
	}
	if settledPaid > futureAmount {
		settledPaid = futureAmount
	}
	cancelledUnpaid := futureAmount - settledPaid
	var bankLedger int64
	if decision == "refund" && settledPaid > 0 {
		bankLedger, err = financialAccountLedgerID(r.Context(), tx, session.User.ID, "bank", bankID, true)
		if err != nil {
			http.Error(w, "Rekening refund tidak ditemukan.", http.StatusNotFound)
			return
		}
		balance, balanceErr := ledgerBalance(r.Context(), tx, bankLedger)
		if balanceErr != nil {
			a.serverError(w, r, balanceErr)
			return
		}
		if balance < settledPaid {
			http.Error(w, "Saldo pembukuan rekening tidak mencukupi untuk refund terminasi.", http.StatusConflict)
			return
		}
	}
	terminationNumber, _ := referenceNumber("TRM")
	result, err := tx.ExecContext(r.Context(), `INSERT INTO lease_terminations (termination_number,lease_id,owner_id,tenant_id,property_id,effective_date,settlement_decision,cancelled_unpaid_amount,settled_paid_amount,bank_account_id,proof_storage_path,reason,created_by) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`, terminationNumber, leaseID, session.User.ID, tenantID, propertyID, effectiveDate, decision, cancelledUnpaid, settledPaid, nullablePositive(bankID), nullableString(refundProof.RelativePath), reason, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	terminationID, _ := result.LastInsertId()
	if futureAmount > 0 {
		unearnedID, accountErr := systemAccountID(r.Context(), tx, session.User.ID, "unearned_rent")
		if accountErr != nil {
			a.serverError(w, r, accountErr)
			return
		}
		lines := []JournalLineInput{{AccountID: unearnedID, Debit: futureAmount, Description: "Pelepasan pendapatan diterima di muka"}}
		if cancelledUnpaid > 0 {
			arID, accountErr := systemAccountID(r.Context(), tx, session.User.ID, "accounts_receivable")
			if accountErr != nil {
				a.serverError(w, r, accountErr)
				return
			}
			lines = append(lines, JournalLineInput{AccountID: arID, Credit: cancelledUnpaid, Description: "Pembatalan piutang masa depan"})
		}
		if settledPaid > 0 {
			if decision == "refund" {
				lines = append(lines, JournalLineInput{AccountID: bankLedger, Credit: settledPaid, Description: "Refund terminasi dini"})
			} else {
				revenueID, accountErr := systemAccountID(r.Context(), tx, session.User.ID, "rent_revenue")
				if accountErr != nil {
					a.serverError(w, r, accountErr)
					return
				}
				lines = append(lines, JournalLineInput{AccountID: revenueID, Credit: settledPaid, Description: "Dana dipertahankan sesuai keputusan terminasi"})
			}
		}
		_, err = postJournalTx(r.Context(), tx, JournalInput{OwnerID: session.User.ID, PropertyID: sql.NullInt64{Int64: propertyID, Valid: true}, Date: effectiveDate, Description: "Penyelesaian terminasi " + terminationNumber, SourceType: "lease_termination", SourceID: terminationID, EventType: decision, ActorID: session.User.ID, Lines: lines})
		if errors.Is(err, ErrClosedPeriod) {
			http.Error(w, "Tanggal terminasi berada pada periode tertutup.", http.StatusConflict)
			return
		}
		if err != nil {
			a.serverError(w, r, err)
			return
		}
	}
	for _, scheduleID := range cancelScheduleIDs {
		if _, err = tx.ExecContext(r.Context(), `UPDATE revenue_recognition_schedules SET status='cancelled' WHERE id=? AND status='scheduled'`, scheduleID); err != nil {
			a.serverError(w, r, err)
			return
		}
	}
	for _, invoiceID := range voidInvoiceIDs {
		if _, err = tx.ExecContext(r.Context(), `UPDATE invoices SET status='void',voided_at=UTC_TIMESTAMP(6),voided_by=?,void_reason=? WHERE id=?`, session.User.ID, "Terminasi "+terminationNumber+": "+reason, invoiceID); err != nil {
			a.serverError(w, r, err)
			return
		}
	}
	inspectionResult, err := tx.ExecContext(r.Context(), `INSERT INTO inspections (lease_id,inspection_type,inspected_at,overall_notes,photo_storage_path,recorded_by) VALUES (?,'check_out',UTC_TIMESTAMP(6),?,?,?)`, leaseID, nullableString("Terminasi "+terminationNumber+": "+reason), inspectionPhoto.RelativePath, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	inspectionID, _ := inspectionResult.LastInsertId()
	assetRows, err := tx.QueryContext(r.Context(), `SELECT id,name,condition_status FROM assets WHERE room_id=? AND condition_status<>'disposed'`, roomID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for assetRows.Next() {
		var assetID int64
		var assetName, condition string
		if err := assetRows.Scan(&assetID, &assetName, &condition); err != nil {
			assetRows.Close()
			a.serverError(w, r, err)
			return
		}
		if _, err = tx.ExecContext(r.Context(), `INSERT INTO inspection_items (inspection_id,asset_id,item_name_snapshot,condition_snapshot) VALUES (?,?,?,?)`, inspectionID, assetID, assetName, condition); err != nil {
			assetRows.Close()
			a.serverError(w, r, err)
			return
		}
	}
	assetRows.Close()
	var version int
	_ = tx.QueryRowContext(r.Context(), `SELECT COALESCE(MAX(version_number),0)+1 FROM lease_versions WHERE lease_id=?`, leaseID).Scan(&version)
	before, _ := json.Marshal(map[string]any{"end_date": oldEnd.Format("2006-01-02"), "status": status})
	after, _ := json.Marshal(map[string]any{"end_date": effectiveDate.Format("2006-01-02"), "status": "terminated", "settlement_decision": decision, "settled_paid": settledPaid, "cancelled_unpaid": cancelledUnpaid})
	if _, err = tx.ExecContext(r.Context(), `INSERT INTO lease_versions (lease_id,version_number,effective_date,change_type,before_json,after_json,reason,created_by) VALUES (?,?,?,'terminated',?,?,?,?)`, leaseID, version, effectiveDate, string(before), string(after), reason, session.User.ID); err != nil {
		a.serverError(w, r, err)
		return
	}
	if _, err = tx.ExecContext(r.Context(), `UPDATE leases SET end_date=?,status='terminated',actual_check_out_at=UTC_TIMESTAMP(6),checked_out_by=? WHERE id=?`, effectiveDate, session.User.ID, leaseID); err != nil {
		a.serverError(w, r, err)
		return
	}
	if _, err = tx.ExecContext(r.Context(), `UPDATE rooms SET condition_status='maintenance' WHERE id=?`, roomID); err != nil {
		a.serverError(w, r, err)
		return
	}
	message := fmt.Sprintf("Kontrak %s diterminasi per %s. Keputusan: %s. %s", number, effectiveDate.Format("02-01-2006"), decision, reason)
	if err := createNotificationTx(r.Context(), tx, tenantID, fmt.Sprintf("lease.terminated.%d", terminationID), "Kontrak diterminasi", message, "/penghuni/kontrak"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "lease.terminate", "lease_termination", terminationID, string(before), string(after), a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/kontrak/"+strconv.FormatInt(leaseID, 10)+"?berhasil=terminasi", http.StatusSeeOther)
}

func (a *App) ownerLeaseExtend(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	leaseID, err := parseID(r.PathValue("id"))
	newEnd, dateErr := parseBusinessDate(r.FormValue("new_end_date"))
	newRent, _ := strconv.ParseInt(r.FormValue("new_rent"), 10, 64)
	reason := strings.TrimSpace(r.FormValue("reason"))
	if err != nil || dateErr != nil || newRent <= 0 || len(reason) < 5 {
		http.Error(w, "Tanggal, tarif, dan alasan perpanjangan wajib valid.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var propertyID, roomID, tenantID int64
	var number, cycle, status string
	var oldEnd time.Time
	var anchor, due, installments int
	var oldRent int64
	var partial bool
	err = tx.QueryRowContext(r.Context(), `SELECT l.property_id,l.room_id,l.tenant_id,l.lease_number,l.billing_cycle,l.status,l.end_date,l.original_anchor_day,l.due_day,l.installment_count,l.agreed_rent,l.partial_payment_allowed FROM leases l JOIN properties p ON p.id=l.property_id WHERE l.id=? AND p.owner_id=? FOR UPDATE`, leaseID, session.User.ID).Scan(&propertyID, &roomID, &tenantID, &number, &cycle, &status, &oldEnd, &anchor, &due, &installments, &oldRent, &partial)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if status != "active" && status != "expiring" {
		http.Error(w, "Hanya kontrak aktif yang dapat diperpanjang.", http.StatusConflict)
		return
	}
	charges, revenue, planErr := domain.BuildLeasePlan(oldEnd, newEnd, anchor, due, cycle, newRent, installments)
	if planErr != nil {
		http.Error(w, planErr.Error(), http.StatusUnprocessableEntity)
		return
	}
	var version int
	_ = tx.QueryRowContext(r.Context(), `SELECT COALESCE(MAX(version_number),0)+1 FROM lease_versions WHERE lease_id=?`, leaseID).Scan(&version)
	before, _ := json.Marshal(map[string]any{"end_date": oldEnd.Format("2006-01-02"), "rent": oldRent})
	after, _ := json.Marshal(map[string]any{"end_date": newEnd.Format("2006-01-02"), "rent": newRent})
	_, err = tx.ExecContext(r.Context(), `INSERT INTO lease_versions (lease_id,version_number,effective_date,change_type,before_json,after_json,reason,created_by) VALUES (?,?,?,'extended',?,?,?,?)`, leaseID, version, oldEnd, string(before), string(after), reason, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	_, err = tx.ExecContext(r.Context(), `UPDATE leases SET end_date=?,agreed_rent=?,status='active' WHERE id=?`, newEnd, newRent, leaseID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	_, err = a.createLeaseInvoicesTx(r.Context(), tx, session.User.ID, leaseID, propertyID, roomID, tenantID, number, charges, revenue)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "lease.extend", "lease", leaseID, string(before), string(after), a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/kontrak/"+strconv.FormatInt(leaseID, 10)+"?berhasil=perpanjang", http.StatusSeeOther)
}

func (a *App) ownerDamageCharge(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	leaseID, err := parseID(r.PathValue("id"))
	amount, _ := strconv.ParseInt(r.FormValue("amount"), 10, 64)
	basis := r.FormValue("basis")
	reason := strings.TrimSpace(r.FormValue("reason"))
	dueDate, dateErr := parseBusinessDate(r.FormValue("due_date"))
	if err != nil || dateErr != nil || amount <= 0 || !map[string]bool{"actual_cost": true, "agreement": true}[basis] || len(reason) < 5 {
		http.Error(w, "Data tagihan kerusakan tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var propertyID, roomID, tenantID int64
	err = tx.QueryRowContext(r.Context(), `SELECT l.property_id,l.room_id,l.tenant_id FROM leases l JOIN properties p ON p.id=l.property_id WHERE l.id=? AND p.owner_id=? FOR UPDATE`, leaseID, session.User.ID).Scan(&propertyID, &roomID, &tenantID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	number, _ := referenceNumber("INV-DMG")
	today := businessToday(a.location)
	result, err := tx.ExecContext(r.Context(), `INSERT INTO invoices (invoice_number,property_id,room_id,tenant_id,lease_id,invoice_type,issue_date,due_date,total_amount,status,issued_by,notes) VALUES (?,?,?,?,?,'damage',?,?,?,'issued',?,?)`, number, propertyID, roomID, tenantID, leaseID, today, dueDate, amount, session.User.ID, reason)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	invoiceID, _ := result.LastInsertId()
	_, err = tx.ExecContext(r.Context(), `INSERT INTO invoice_lines (invoice_id,description,quantity,unit_amount,line_total) VALUES (?, ?,1,?,?)`, invoiceID, "Ganti rugi kerusakan: "+reason, amount, amount)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	result, err = tx.ExecContext(r.Context(), `INSERT INTO damage_charges (lease_id,invoice_id,amount,basis,reason,created_by) VALUES (?,?,?,?,?,?)`, leaseID, invoiceID, amount, basis, reason, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	damageID, _ := result.LastInsertId()
	arID, err := systemAccountID(r.Context(), tx, session.User.ID, "accounts_receivable")
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	recoveryID, err := systemAccountID(r.Context(), tx, session.User.ID, "damage_recovery")
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	_, err = postJournalTx(r.Context(), tx, JournalInput{OwnerID: session.User.ID, PropertyID: sql.NullInt64{Int64: propertyID, Valid: true}, Date: today, Description: "Tagihan ganti rugi " + number, SourceType: "damage_charge", SourceID: damageID, EventType: "issued", ActorID: session.User.ID, Lines: []JournalLineInput{{AccountID: arID, Debit: amount}, {AccountID: recoveryID, Credit: amount}}})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := createNotificationTx(r.Context(), tx, tenantID, fmt.Sprintf("damage.issued.%d", damageID), "Tagihan ganti rugi", "Tagihan "+number+" diterbitkan: "+reason, "/penghuni/tagihan"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "damage_charge.create", "damage_charge", damageID, nil, map[string]any{"invoice_id": invoiceID, "amount": amount, "basis": basis}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/kontrak/"+strconv.FormatInt(leaseID, 10)+"?berhasil=kerusakan", http.StatusSeeOther)
}

func createNotificationTx(ctx context.Context, tx *sql.Tx, userID int64, eventKey, title, body, url string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO notifications (user_id,event_key,title,body,action_url) VALUES (?,?,?,?,?)`, userID, eventKey, title, body, nullableString(url))
	return err
}
