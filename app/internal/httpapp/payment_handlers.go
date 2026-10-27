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

type BankPaymentOption struct {
	ID                 int64
	Bank, Name, Number string
}
type BookingPaymentPage struct {
	Banks            []BankPaymentOption
	Reference, Email string
}
type PaymentSubmissionView struct {
	ID, PaymentID                                            int64
	Number, Tenant, Invoice, Bank, Reference, Status, Reason string
	Date, CreatedAt                                          time.Time
	Amount, Refundable                                       int64
}
type PaymentPage struct {
	Invoices      []InvoiceView
	Submissions   []PaymentSubmissionView
	Banks         []BankPaymentOption
	Accounts      []FinanceAccountView
	Refunds       []RefundView
	Credits       []TenantCreditView
	CreditRefunds []TenantCreditRefundView
}
type RefundView struct {
	ID, Amount              int64
	Number, Invoice, Reason string
	Date                    time.Time
}

type TenantCreditView struct {
	ID, Amount, Remaining            int64
	Tenant, Property, Source, Status string
	CreatedAt                        time.Time
}

type TenantCreditRefundView struct {
	ID, Amount                       int64
	Number, Tenant, Property, Reason string
	Date                             time.Time
}

func (a *App) bookingPaymentForm(w http.ResponseWriter, r *http.Request) {
	session := sessionFrom(r.Context())
	if session == nil {
		var err error
		session, err = a.sessions.Start(r.Context(), w, nil)
		if err != nil {
			a.serverError(w, r, err)
			return
		}
	}
	page := BookingPaymentPage{Reference: strings.TrimSpace(r.URL.Query().Get("referensi")), Email: strings.ToLower(strings.TrimSpace(r.URL.Query().Get("email")))}
	data := PageData{Title: "Upload Bukti — Alam Segar", CSRFToken: session.CSRFToken, Payload: page}
	if page.Reference != "" && page.Email != "" {
		var propertyID int64
		err := a.db.QueryRowContext(r.Context(), `SELECT property_id FROM booking_requests WHERE reference_number=? AND applicant_email=? AND status='approved_waiting_payment'`, page.Reference, page.Email).Scan(&propertyID)
		if errors.Is(err, sql.ErrNoRows) {
			data.Error = "Booking belum disetujui untuk pembayaran atau referensi dan email tidak cocok."
		} else if err != nil {
			a.serverError(w, r, err)
			return
		} else {
			page.Banks, err = a.paymentBankOptionsForProperty(r.Context(), propertyID)
			if err != nil {
				a.serverError(w, r, err)
				return
			}
			data.Payload = page
		}
	}
	a.render(w, http.StatusOK, "public_booking_payment", data)
}

func (a *App) submitBookingPayment(w http.ResponseWriter, r *http.Request) {
	session := sessionFrom(r.Context())
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20)
	if err := r.ParseMultipartForm(8 << 20); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	reference := strings.TrimSpace(r.FormValue("reference"))
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	bankID, e1 := parseID(r.FormValue("bank_account_id"))
	date, e2 := parseBusinessDate(r.FormValue("transfer_date"))
	amount, _ := strconv.ParseInt(r.FormValue("amount"), 10, 64)
	page := BookingPaymentPage{Reference: reference, Email: email}
	data := PageData{Title: "Upload Bukti — Alam Segar", CSRFToken: session.CSRFToken, Payload: page}
	if reference == "" || email == "" || e1 != nil || e2 != nil || amount <= 0 {
		data.Error = "Data transfer belum lengkap atau tidak valid."
		a.render(w, http.StatusUnprocessableEntity, "public_booking_payment", data)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var bookingID, propertyID, expected int64
	var status string
	err = tx.QueryRowContext(r.Context(), `SELECT id,property_id,booking_amount,status FROM booking_requests WHERE reference_number=? AND applicant_email=? FOR UPDATE`, reference, email).Scan(&bookingID, &propertyID, &expected, &status)
	if errors.Is(err, sql.ErrNoRows) {
		data.Error = "Referensi booking dan email tidak cocok."
		a.render(w, http.StatusNotFound, "public_booking_payment", data)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if status != "approved_waiting_payment" || amount != expected {
		data.Error = "Booking belum menunggu pembayaran atau nominal tidak sama dengan keputusan pemilik."
		a.render(w, http.StatusConflict, "public_booking_payment", data)
		return
	}
	page.Banks, _ = a.paymentBankOptionsForProperty(r.Context(), propertyID)
	data.Payload = page
	var bankValid int
	if err := tx.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM bank_accounts ba JOIN properties p ON p.owner_id=ba.owner_id WHERE ba.id=? AND ba.status='active' AND ba.is_accepting_payments=1 AND (ba.property_id IS NULL OR ba.property_id=?) AND p.id=?`, bankID, propertyID, propertyID).Scan(&bankValid); err != nil || bankValid == 0 {
		data.Error = "Rekening tujuan tidak valid untuk properti ini."
		a.render(w, http.StatusUnprocessableEntity, "public_booking_payment", data)
		return
	}
	var active int
	_ = tx.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM booking_payment_submissions WHERE booking_request_id=? AND status IN ('pending_verification','verified')`, bookingID).Scan(&active)
	if active > 0 {
		data.Error = "Sudah ada bukti yang sedang diperiksa atau telah diverifikasi."
		a.render(w, http.StatusConflict, "public_booking_payment", data)
		return
	}
	upload, err := a.savePrivateUpload(r, "proof", "booking-payments", 5<<20, true)
	if err != nil {
		data.Error = err.Error()
		a.render(w, http.StatusUnprocessableEntity, "public_booking_payment", data)
		return
	}
	number, _ := referenceNumber("BPS")
	result, err := tx.ExecContext(r.Context(), `INSERT INTO booking_payment_submissions (submission_number,booking_request_id,bank_account_id,amount,transfer_date,bank_reference,proof_storage_path,proof_sha256) VALUES (?,?,?,?,?,?,?,?)`, number, bookingID, bankID, amount, date, nullableString(strings.TrimSpace(r.FormValue("bank_reference"))), upload.RelativePath, upload.SHA256)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	submissionID, _ := result.LastInsertId()
	_, err = tx.ExecContext(r.Context(), `UPDATE booking_requests SET status='proof_submitted' WHERE id=?`, bookingID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := insertAuditTx(r.Context(), tx, auditActor(session), "booking.payment_submit", "booking_payment_submission", submissionID, nil, map[string]any{"booking_id": bookingID, "amount": amount}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	data.Success = "Bukti berhasil dikirim. Dana belum dianggap diterima sampai pemilik memverifikasi mutasi bank."
	a.render(w, http.StatusCreated, "public_booking_payment", data)
}

func (a *App) verifyBookingPayment(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
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
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var bookingID, bankID, amount, propertyID int64
	var transferDate time.Time
	var status, bookingStatus, applicantEmail, bookingReference string
	err = tx.QueryRowContext(r.Context(), `SELECT bps.booking_request_id,bps.bank_account_id,bps.amount,bps.transfer_date,bps.status,b.property_id,b.status,b.applicant_email,b.reference_number FROM booking_payment_submissions bps JOIN booking_requests b ON b.id=bps.booking_request_id JOIN properties p ON p.id=b.property_id WHERE bps.id=? AND p.owner_id=? FOR UPDATE`, id, session.User.ID).Scan(&bookingID, &bankID, &amount, &transferDate, &status, &propertyID, &bookingStatus, &applicantEmail, &bookingReference)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if _, err = tx.ExecContext(r.Context(), `INSERT INTO notification_deliveries (user_id,channel,event_key,recipient,subject,body) VALUES (NULL,'email',?,?,?,?)`, fmt.Sprintf("booking.payment_verified.%d", bookingID), applicantEmail, "Dana booking Alam Segar terverifikasi", "Dana booking "+bookingReference+" telah diverifikasi pemilik. Pemilik akan menyiapkan akun penghuni dan kontrak Anda."); err != nil {
		a.serverError(w, r, err)
		return
	}
	if status != "pending_verification" || bookingStatus != "proof_submitted" {
		http.Error(w, "Bukti sudah diproses.", http.StatusConflict)
		return
	}
	bankLedger, err := financialAccountLedgerID(r.Context(), tx, session.User.ID, "bank", bankID, true)
	if err != nil {
		http.Error(w, "Rekening belum memiliki akun buku besar.", http.StatusConflict)
		return
	}
	liability, err := systemAccountID(r.Context(), tx, session.User.ID, "booking_liability")
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	result, err := tx.ExecContext(r.Context(), `INSERT INTO booking_funds (booking_request_id,booking_payment_submission_id,received_amount,remaining_liability,verified_at) VALUES (?,?,?,?,UTC_TIMESTAMP(6))`, bookingID, id, amount, amount)
	if err != nil {
		if isDuplicate(err) {
			http.Error(w, "Dana booking sudah pernah diverifikasi.", http.StatusConflict)
			return
		}
		a.serverError(w, r, err)
		return
	}
	fundID, _ := result.LastInsertId()
	_, err = postJournalTx(r.Context(), tx, JournalInput{OwnerID: session.User.ID, PropertyID: sql.NullInt64{Int64: propertyID, Valid: true}, Date: transferDate, Description: "Penerimaan uang booking", SourceType: "booking_fund", SourceID: fundID, EventType: "verified", ActorID: session.User.ID, Lines: []JournalLineInput{{AccountID: bankLedger, Debit: amount}, {AccountID: liability, Credit: amount}}})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	_, err = tx.ExecContext(r.Context(), `UPDATE booking_payment_submissions SET status='verified',reviewed_by=?,reviewed_at=UTC_TIMESTAMP(6) WHERE id=?`, session.User.ID, id)
	if err == nil {
		_, err = tx.ExecContext(r.Context(), `UPDATE booking_requests SET status='confirmed' WHERE id=?`, bookingID)
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "booking.payment_verify", "booking_payment_submission", id, map[string]any{"status": status}, map[string]any{"status": "verified", "fund_id": fundID}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/booking?berhasil=dana", http.StatusSeeOther)
}

func (a *App) rejectBookingPayment(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	id, err := parseID(r.PathValue("id"))
	reason := strings.TrimSpace(r.FormValue("reason"))
	if err != nil || len(reason) < 5 {
		http.Error(w, "Alasan minimal 5 karakter.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var applicantEmail, bookingReference string
	if err := tx.QueryRowContext(r.Context(), `SELECT b.applicant_email,b.reference_number FROM booking_payment_submissions bps JOIN booking_requests b ON b.id=bps.booking_request_id JOIN properties p ON p.id=b.property_id WHERE bps.id=? AND p.owner_id=? FOR UPDATE`, id, session.User.ID).Scan(&applicantEmail, &bookingReference); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		a.serverError(w, r, err)
		return
	}
	result, err := tx.ExecContext(r.Context(), `UPDATE booking_payment_submissions bps JOIN booking_requests b ON b.id=bps.booking_request_id JOIN properties p ON p.id=b.property_id SET bps.status='rejected',bps.reviewed_by=?,bps.reviewed_at=UTC_TIMESTAMP(6),bps.rejection_reason=?,b.status='approved_waiting_payment' WHERE bps.id=? AND p.owner_id=? AND bps.status='pending_verification'`, session.User.ID, reason, id, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	affected, _ := result.RowsAffected()
	if affected != 1 {
		http.Error(w, "Bukti sudah diproses.", http.StatusConflict)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `INSERT INTO notification_deliveries (user_id,channel,event_key,recipient,subject,body) VALUES (NULL,'email',?,?,?,?)`, fmt.Sprintf("booking.payment_rejected.%d", id), applicantEmail, "Bukti uang booking perlu diperbaiki", "Bukti transfer untuk booking "+bookingReference+" ditolak. Alasan: "+reason+". Silakan kirim bukti yang benar melalui halaman konfirmasi uang booking."); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "booking.payment_reject", "booking_payment_submission", id, nil, map[string]any{"status": "rejected", "reason": reason}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/booking?berhasil=bukti-ditolak", http.StatusSeeOther)
}

func (a *App) releaseBooking(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	id, err := parseID(r.PathValue("id"))
	reason := strings.TrimSpace(r.FormValue("reason"))
	if err != nil || len(reason) < 5 {
		http.Error(w, "Alasan pelepasan minimal 5 karakter.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var applicantEmail, bookingReference string
	if err := tx.QueryRowContext(r.Context(), `SELECT b.applicant_email,b.reference_number FROM booking_requests b JOIN properties p ON p.id=b.property_id WHERE b.id=? AND p.owner_id=? FOR UPDATE`, id, session.User.ID).Scan(&applicantEmail, &bookingReference); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		a.serverError(w, r, err)
		return
	}
	result, err := tx.ExecContext(r.Context(), `UPDATE booking_requests b JOIN properties p ON p.id=b.property_id SET b.status='released',b.owner_reason=? WHERE b.id=? AND p.owner_id=? AND b.status='approved_waiting_payment'`, reason, id, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	affected, _ := result.RowsAffected()
	if affected != 1 {
		http.Error(w, "Booking tidak dapat dilepaskan karena sudah memiliki pembayaran atau sudah diproses.", http.StatusConflict)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `INSERT INTO notification_deliveries (user_id,channel,event_key,recipient,subject,body) VALUES (NULL,'email',?,?,?,?)`, fmt.Sprintf("booking.released.%d", id), applicantEmail, "Booking Alam Segar dilepas", "Hold kamar untuk booking "+bookingReference+" telah dilepas manual. Alasan pemilik: "+reason); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "booking.release", "booking_request", id, nil, map[string]any{"status": "released", "reason": reason}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/booking?berhasil=dilepas", http.StatusSeeOther)
}

func (a *App) refundBooking(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20)
	if err := r.ParseMultipartForm(8 << 20); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	bookingID, err := parseID(r.PathValue("id"))
	bankID, e1 := parseID(r.FormValue("bank_account_id"))
	date, e2 := parseBusinessDate(r.FormValue("refund_date"))
	amount, _ := strconv.ParseInt(r.FormValue("amount"), 10, 64)
	reason := strings.TrimSpace(r.FormValue("reason"))
	if err != nil || e1 != nil || e2 != nil || amount <= 0 || len(reason) < 5 {
		http.Error(w, "Data refund tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	upload, err := a.savePrivateUpload(r, "proof", "booking-refunds", 5<<20, true)
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
	var fundID, remaining, propertyID int64
	var bookingStatus, fundStatus, applicantEmail, bookingReference string
	err = tx.QueryRowContext(r.Context(), `SELECT bf.id,bf.remaining_liability,bf.status,b.property_id,b.status,b.applicant_email,b.reference_number FROM booking_funds bf JOIN booking_requests b ON b.id=bf.booking_request_id JOIN properties p ON p.id=b.property_id WHERE b.id=? AND p.owner_id=? FOR UPDATE`, bookingID, session.User.ID).Scan(&fundID, &remaining, &fundStatus, &propertyID, &bookingStatus, &applicantEmail, &bookingReference)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if fundStatus != "held" || bookingStatus != "confirmed" || amount != remaining {
		http.Error(w, "Refund booking harus sebesar seluruh sisa dana yang masih ditahan agar status booking dan kewajiban tetap konsisten.", http.StatusConflict)
		return
	}
	bankLedger, err := financialAccountLedgerID(r.Context(), tx, session.User.ID, "bank", bankID, true)
	if err != nil {
		http.Error(w, "Rekening sumber tidak ditemukan.", http.StatusNotFound)
		return
	}
	balance, _ := ledgerBalance(r.Context(), tx, bankLedger)
	if balance < amount {
		http.Error(w, "Saldo pembukuan rekening tidak mencukupi.", http.StatusConflict)
		return
	}
	liability, err := systemAccountID(r.Context(), tx, session.User.ID, "booking_liability")
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	number, _ := referenceNumber("RFD-BKG")
	result, err := tx.ExecContext(r.Context(), `INSERT INTO booking_refunds (refund_number,booking_fund_id,bank_account_id,amount,refund_date,reason,proof_storage_path,created_by) VALUES (?,?,?,?,?,?,?,?)`, number, fundID, bankID, amount, date, reason, upload.RelativePath, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	refundID, _ := result.LastInsertId()
	_, err = postJournalTx(r.Context(), tx, JournalInput{OwnerID: session.User.ID, PropertyID: sql.NullInt64{Int64: propertyID, Valid: true}, Date: date, Description: "Refund uang booking " + number, SourceType: "booking_refund", SourceID: refundID, EventType: "posted", ActorID: session.User.ID, Lines: []JournalLineInput{{AccountID: liability, Debit: amount}, {AccountID: bankLedger, Credit: amount}}})
	if err != nil {
		if errors.Is(err, ErrClosedPeriod) {
			http.Error(w, "Tanggal refund berada pada periode tertutup.", http.StatusConflict)
			return
		}
		a.serverError(w, r, err)
		return
	}
	newRemaining := int64(0)
	newFundStatus := "refunded"
	newBookingStatus := "refunded"
	_, err = tx.ExecContext(r.Context(), `UPDATE booking_funds SET remaining_liability=?,status=? WHERE id=?`, newRemaining, newFundStatus, fundID)
	if err == nil {
		_, err = tx.ExecContext(r.Context(), `UPDATE booking_requests SET status=?,owner_reason=? WHERE id=?`, newBookingStatus, reason, bookingID)
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `INSERT INTO notification_deliveries (user_id,channel,event_key,recipient,subject,body) VALUES (NULL,'email',?,?,?,?)`, fmt.Sprintf("booking.refunded.%d", refundID), applicantEmail, "Refund booking Alam Segar diproses", "Refund penuh untuk booking "+bookingReference+" sebesar "+formatRupiah(amount)+" telah dicatat. Referensi refund: "+number+". Alasan: "+reason); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "booking.refund", "booking_refund", refundID, nil, map[string]any{"booking_id": bookingID, "amount": amount, "reason": reason}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/booking?berhasil=refund", http.StatusSeeOther)
}

func (a *App) ownerInvoiceList(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	page := PaymentPage{}
	rows, err := a.db.QueryContext(r.Context(), `SELECT i.id,i.invoice_number,u.name,p.name,COALESCE(rm.number,''),i.invoice_type,i.status,i.issue_date,i.due_date,i.total_amount,i.paid_amount,COALESCE((SELECT SUM(tc.remaining_amount) FROM tenant_credits tc WHERE tc.tenant_id=i.tenant_id AND tc.property_id=i.property_id AND tc.status='available'),0) FROM invoices i JOIN users u ON u.id=i.tenant_id JOIN properties p ON p.id=i.property_id LEFT JOIN rooms rm ON rm.id=i.room_id WHERE p.owner_id=? ORDER BY i.due_date DESC,i.id DESC`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var v InvoiceView
		if err := rows.Scan(&v.ID, &v.Number, &v.TenantName, &v.PropertyName, &v.RoomNumber, &v.Type, &v.Status, &v.IssueDate, &v.DueDate, &v.Total, &v.Paid, &v.CreditAvailable); err != nil {
			a.serverError(w, r, err)
			return
		}
		page.Invoices = append(page.Invoices, v)
	}
	data := PageData{Title: "Tagihan — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page}
	if r.URL.Query().Get("berhasil") != "" {
		data.Success = "Tagihan berhasil diperbarui."
	}
	a.render(w, http.StatusOK, "owner_invoices", data)
}

func (a *App) ownerApplyTenantCredit(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	invoiceID, idErr := parseID(r.PathValue("id"))
	requested, amountErr := strconv.ParseInt(r.FormValue("amount"), 10, 64)
	if idErr != nil || amountErr != nil || requested <= 0 {
		http.Error(w, "Nominal penggunaan kredit tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var tenantID, propertyID, total, paid int64
	var status string
	err = tx.QueryRowContext(r.Context(), `SELECT i.tenant_id,i.property_id,i.total_amount,i.paid_amount,i.status FROM invoices i JOIN properties p ON p.id=i.property_id WHERE i.id=? AND p.owner_id=? FOR UPDATE`, invoiceID, session.User.ID).Scan(&tenantID, &propertyID, &total, &paid, &status)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if status == "paid" || status == "void" || status == "refunded" || total-paid <= 0 {
		http.Error(w, "Tagihan tidak dapat menerima kredit.", http.StatusConflict)
		return
	}
	type creditItem struct{ id, remaining int64 }
	rows, err := tx.QueryContext(r.Context(), `SELECT id,remaining_amount FROM tenant_credits WHERE tenant_id=? AND property_id=? AND status='available' AND remaining_amount>0 ORDER BY created_at,id FOR UPDATE`, tenantID, propertyID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	var credits []creditItem
	var available int64
	for rows.Next() {
		var item creditItem
		if err := rows.Scan(&item.id, &item.remaining); err != nil {
			rows.Close()
			a.serverError(w, r, err)
			return
		}
		credits = append(credits, item)
		available += item.remaining
	}
	rows.Close()
	outstanding := total - paid
	if requested > outstanding || requested > available {
		http.Error(w, "Nominal melebihi sisa tagihan atau kredit tenant yang tersedia.", http.StatusConflict)
		return
	}
	number, _ := referenceNumber("KRD")
	today := businessToday(a.location)
	result, err := tx.ExecContext(r.Context(), `INSERT INTO tenant_credit_applications (application_number,owner_id,tenant_id,property_id,invoice_id,amount,application_date,created_by) VALUES (?,?,?,?,?,?,?,?)`, number, session.User.ID, tenantID, propertyID, invoiceID, requested, today, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	applicationID, _ := result.LastInsertId()
	remaining := requested
	for _, item := range credits {
		if remaining == 0 {
			break
		}
		use := item.remaining
		if use > remaining {
			use = remaining
		}
		if _, err = tx.ExecContext(r.Context(), `INSERT INTO tenant_credit_application_lines (application_id,tenant_credit_id,amount) VALUES (?,?,?)`, applicationID, item.id, use); err != nil {
			a.serverError(w, r, err)
			return
		}
		newRemaining := item.remaining - use
		newStatus := "available"
		if newRemaining == 0 {
			newStatus = "applied"
		}
		if _, err = tx.ExecContext(r.Context(), `UPDATE tenant_credits SET remaining_amount=?,status=? WHERE id=?`, newRemaining, newStatus, item.id); err != nil {
			a.serverError(w, r, err)
			return
		}
		remaining -= use
	}
	creditLiability, err := systemAccountID(r.Context(), tx, session.User.ID, "tenant_credit")
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	arID, err := systemAccountID(r.Context(), tx, session.User.ID, "accounts_receivable")
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	_, err = postJournalTx(r.Context(), tx, JournalInput{OwnerID: session.User.ID, PropertyID: sql.NullInt64{Int64: propertyID, Valid: true}, Date: today, Description: "Pemakaian kredit tenant " + number, SourceType: "tenant_credit_application", SourceID: applicationID, EventType: "applied", ActorID: session.User.ID, Lines: []JournalLineInput{{AccountID: creditLiability, Debit: requested}, {AccountID: arID, Credit: requested}}})
	if err != nil {
		if errors.Is(err, ErrClosedPeriod) {
			http.Error(w, "Periode saat ini sudah ditutup.", http.StatusConflict)
			return
		}
		a.serverError(w, r, err)
		return
	}
	newPaid := paid + requested
	newStatus := "partially_paid"
	if newPaid == total {
		newStatus = "paid"
	}
	if _, err = tx.ExecContext(r.Context(), `UPDATE invoices SET paid_amount=?,status=? WHERE id=?`, newPaid, newStatus, invoiceID); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := createNotificationTx(r.Context(), tx, tenantID, fmt.Sprintf("tenant_credit.applied.%d", applicationID), "Kredit diterapkan ke tagihan", "Kredit "+number+" telah diterapkan ke tagihan.", "/penghuni/tagihan"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "tenant_credit.apply", "tenant_credit_application", applicationID, nil, map[string]any{"invoice_id": invoiceID, "amount": requested}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/tagihan?berhasil=kredit", http.StatusSeeOther)
}

func (a *App) ownerVoidInvoice(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	id, err := parseID(r.PathValue("id"))
	reason := strings.TrimSpace(r.FormValue("reason"))
	if err != nil || len(reason) < 5 {
		http.Error(w, "Alasan pembatalan minimal 5 karakter.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var paid, total, leaseID, propertyID int64
	var status, invoiceType string
	err = tx.QueryRowContext(r.Context(), `SELECT i.paid_amount,i.total_amount,COALESCE(i.lease_id,0),i.property_id,i.status,i.invoice_type FROM invoices i JOIN properties p ON p.id=i.property_id WHERE i.id=? AND p.owner_id=? FOR UPDATE`, id, session.User.ID).Scan(&paid, &total, &leaseID, &propertyID, &status, &invoiceType)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if paid != 0 || status == "void" {
		http.Error(w, "Tagihan yang sudah dibayar atau dibatalkan tidak dapat dibatalkan langsung.", http.StatusConflict)
		return
	}
	var journalID int64
	if invoiceType == "rent" {
		err = tx.QueryRowContext(r.Context(), `SELECT id FROM journal_entries WHERE owner_id=? AND source_type='invoice' AND source_id=? AND event_type='issued'`, session.User.ID, id).Scan(&journalID)
	} else if invoiceType == "damage" {
		err = tx.QueryRowContext(r.Context(), `SELECT je.id FROM damage_charges dc JOIN journal_entries je ON je.source_type='damage_charge' AND je.source_id=dc.id AND je.event_type='issued' WHERE dc.invoice_id=? AND je.owner_id=?`, id, session.User.ID).Scan(&journalID)
	} else {
		http.Error(w, "Jenis tagihan ini belum dapat dibatalkan melalui alur ini.", http.StatusConflict)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	voidDate := businessToday(a.location)
	if invoiceType == "rent" {
		unearnedAmount, reduceErr := a.reduceRevenueScheduleTx(r.Context(), tx, leaseID, total)
		if reduceErr != nil {
			a.serverError(w, r, reduceErr)
			return
		}
		recognizedAmount := total - unearnedAmount
		if recognizedAmount > 0 {
			unearnedID, accountErr := systemAccountID(r.Context(), tx, session.User.ID, "unearned_rent")
			if accountErr != nil {
				a.serverError(w, r, accountErr)
				return
			}
			revenueID, accountErr := systemAccountID(r.Context(), tx, session.User.ID, "rent_revenue")
			if accountErr != nil {
				a.serverError(w, r, accountErr)
				return
			}
			_, err = postJournalTx(r.Context(), tx, JournalInput{OwnerID: session.User.ID, PropertyID: sql.NullInt64{Int64: propertyID, Valid: true}, Date: voidDate, Description: "Koreksi pendapatan atas pembatalan tagihan", SourceType: "invoice", SourceID: id, EventType: "void_revenue_adjustment", ActorID: session.User.ID, Lines: []JournalLineInput{{AccountID: revenueID, Debit: recognizedAmount}, {AccountID: unearnedID, Credit: recognizedAmount}}})
			if err != nil {
				if errors.Is(err, ErrClosedPeriod) {
					http.Error(w, "Periode saat ini tertutup.", http.StatusConflict)
					return
				}
				a.serverError(w, r, err)
				return
			}
		}
	}
	if _, err := reverseJournalTx(r.Context(), tx, session.User.ID, journalID, session.User.ID, voidDate, reason); err != nil {
		if errors.Is(err, ErrClosedPeriod) {
			http.Error(w, "Periode saat ini tertutup.", http.StatusConflict)
			return
		}
		a.serverError(w, r, err)
		return
	}
	_, err = tx.ExecContext(r.Context(), `UPDATE invoices SET status='void',voided_at=UTC_TIMESTAMP(6),voided_by=?,void_reason=? WHERE id=?`, session.User.ID, reason, id)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/tagihan?berhasil=void", http.StatusSeeOther)
}

func (a *App) tenantInvoices(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "tenant")
	if session == nil {
		return
	}
	page := PaymentPage{}
	rows, err := a.db.QueryContext(r.Context(), `SELECT i.id,i.invoice_number,'',p.name,COALESCE(rm.number,''),i.invoice_type,CASE WHEN i.status IN ('issued','partially_paid') AND i.due_date<CURRENT_DATE THEN 'overdue' ELSE i.status END,i.issue_date,i.due_date,i.total_amount,i.paid_amount FROM invoices i JOIN properties p ON p.id=i.property_id LEFT JOIN rooms rm ON rm.id=i.room_id WHERE i.tenant_id=? ORDER BY i.due_date DESC`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var v InvoiceView
		if err := rows.Scan(&v.ID, &v.Number, &v.TenantName, &v.PropertyName, &v.RoomNumber, &v.Type, &v.Status, &v.IssueDate, &v.DueDate, &v.Total, &v.Paid); err != nil {
			rows.Close()
			a.serverError(w, r, err)
			return
		}
		page.Invoices = append(page.Invoices, v)
	}
	rows.Close()
	page.Banks, err = a.paymentBankOptions(r.Context(), session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	rows, err = a.db.QueryContext(r.Context(), `SELECT pr.id,pr.refund_number,i.invoice_number,pr.amount,pr.refund_date,pr.reason FROM payment_refunds pr JOIN payments p ON p.id=pr.payment_id JOIN invoices i ON i.id=pr.invoice_id WHERE p.tenant_id=? AND pr.status='posted' ORDER BY pr.refund_date DESC,pr.id DESC`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var v RefundView
		if err := rows.Scan(&v.ID, &v.Number, &v.Invoice, &v.Amount, &v.Date, &v.Reason); err != nil {
			rows.Close()
			a.serverError(w, r, err)
			return
		}
		page.Refunds = append(page.Refunds, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(r.Context(), `SELECT tc.id,tc.amount,tc.remaining_amount,'',p.name,tc.source_type,tc.status,tc.created_at FROM tenant_credits tc JOIN properties p ON p.id=tc.property_id WHERE tc.tenant_id=? ORDER BY tc.created_at DESC,tc.id DESC`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var v TenantCreditView
		if err := rows.Scan(&v.ID, &v.Amount, &v.Remaining, &v.Tenant, &v.Property, &v.Source, &v.Status, &v.CreatedAt); err != nil {
			rows.Close()
			a.serverError(w, r, err)
			return
		}
		page.Credits = append(page.Credits, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(r.Context(), `SELECT id,refund_number,'', '',amount,refund_date,reason FROM tenant_credit_refunds WHERE tenant_id=? AND status='posted' ORDER BY refund_date DESC,id DESC`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var v TenantCreditRefundView
		if err := rows.Scan(&v.ID, &v.Number, &v.Tenant, &v.Property, &v.Amount, &v.Date, &v.Reason); err != nil {
			rows.Close()
			a.serverError(w, r, err)
			return
		}
		page.CreditRefunds = append(page.CreditRefunds, v)
	}
	rows.Close()
	data := PageData{Title: "My Bills — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page}
	if r.URL.Query().Get("berhasil") != "" {
		data.Success = "Bukti transfer berhasil dikirim dan menunggu verifikasi pemilik."
	}
	a.render(w, http.StatusOK, "tenant_invoices", data)
}

func (a *App) tenantSubmitPayment(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "tenant")
	if session == nil {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20)
	if err := r.ParseMultipartForm(8 << 20); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	invoiceID, err := parseID(r.PathValue("id"))
	bankID, e1 := parseID(r.FormValue("bank_account_id"))
	date, e2 := parseBusinessDate(r.FormValue("transfer_date"))
	amount, _ := strconv.ParseInt(r.FormValue("amount"), 10, 64)
	if err != nil || e1 != nil || e2 != nil || amount <= 0 {
		http.Error(w, "Data transfer tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var propertyID, total, paid int64
	var status string
	var partial bool
	err = tx.QueryRowContext(r.Context(), `SELECT i.property_id,i.total_amount,i.paid_amount,i.status,COALESCE(l.partial_payment_allowed,0) FROM invoices i LEFT JOIN leases l ON l.id=i.lease_id WHERE i.id=? AND i.tenant_id=? FOR UPDATE`, invoiceID, session.User.ID).Scan(&propertyID, &total, &paid, &status, &partial)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	outstanding := total - paid
	if status == "paid" || status == "void" || outstanding <= 0 {
		http.Error(w, "Tagihan tidak dapat dibayar.", http.StatusConflict)
		return
	}
	if amount < outstanding && !partial {
		http.Error(w, "Kontrak ini tidak mengizinkan pembayaran sebagian.", http.StatusUnprocessableEntity)
		return
	}
	var bankValid int
	_ = tx.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM bank_accounts WHERE id=? AND status='active' AND is_accepting_payments=1 AND (property_id IS NULL OR property_id=?)`, bankID, propertyID).Scan(&bankValid)
	if bankValid != 1 {
		http.Error(w, "Rekening tujuan tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	var pending int
	_ = tx.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM payment_submission_allocations psa JOIN payment_submissions ps ON ps.id=psa.payment_submission_id WHERE psa.invoice_id=? AND ps.status='pending_verification'`, invoiceID).Scan(&pending)
	if pending > 0 {
		http.Error(w, "Tagihan ini sudah memiliki bukti transfer yang menunggu verifikasi.", http.StatusConflict)
		return
	}
	upload, err := a.savePrivateUpload(r, "proof", "tenant-payments", 5<<20, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	key, _ := randomKey(32)
	number, _ := referenceNumber("PAY-SUB")
	result, err := tx.ExecContext(r.Context(), `INSERT INTO payment_submissions (submission_number,idempotency_key,tenant_id,bank_account_id,amount,transfer_date,bank_reference,proof_storage_path,proof_sha256) VALUES (?,?,?,?,?,?,?,?,?)`, number, key, session.User.ID, bankID, amount, date, nullableString(strings.TrimSpace(r.FormValue("bank_reference"))), upload.RelativePath, upload.SHA256)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	submissionID, _ := result.LastInsertId()
	_, err = tx.ExecContext(r.Context(), `INSERT INTO payment_submission_allocations (payment_submission_id,invoice_id,requested_amount) VALUES (?,?,?)`, submissionID, invoiceID, amount)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := createNotificationForOwnerTx(r.Context(), tx, propertyID, fmt.Sprintf("payment.submitted.%d", submissionID), "Bukti transfer baru", session.User.Name+" mengirim bukti transfer.", "/pemilik/pembayaran"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/penghuni/tagihan?berhasil=dikirim", http.StatusSeeOther)
}

func (a *App) ownerPaymentList(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	page := PaymentPage{}
	finance, _ := a.loadFinancePage(r.Context(), session.User.ID)
	page.Accounts = finance.Accounts
	rows, err := a.db.QueryContext(r.Context(), `SELECT ps.id,COALESCE(pmt.id,0),ps.submission_number,u.name,i.invoice_number,ba.bank_name,COALESCE(ps.bank_reference,''),ps.status,COALESCE(ps.rejection_reason,''),ps.transfer_date,ps.created_at,ps.amount,CASE WHEN pmt.id IS NULL THEN 0 ELSE COALESCE((SELECT SUM(pa.amount) FROM payment_allocations pa WHERE pa.payment_id=pmt.id),0)-COALESCE((SELECT SUM(pr.amount) FROM payment_refunds pr WHERE pr.payment_id=pmt.id AND pr.status='posted'),0) END FROM payment_submissions ps JOIN users u ON u.id=ps.tenant_id JOIN bank_accounts ba ON ba.id=ps.bank_account_id JOIN payment_submission_allocations psa ON psa.payment_submission_id=ps.id JOIN invoices i ON i.id=psa.invoice_id LEFT JOIN payments pmt ON pmt.submission_id=ps.id WHERE ba.owner_id=? ORDER BY FIELD(ps.status,'pending_verification','rejected','verified'),ps.created_at DESC`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var v PaymentSubmissionView
		if err := rows.Scan(&v.ID, &v.PaymentID, &v.Number, &v.Tenant, &v.Invoice, &v.Bank, &v.Reference, &v.Status, &v.Reason, &v.Date, &v.CreatedAt, &v.Amount, &v.Refundable); err != nil {
			a.serverError(w, r, err)
			return
		}
		page.Submissions = append(page.Submissions, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(r.Context(), `SELECT pr.id,pr.refund_number,i.invoice_number,pr.amount,pr.refund_date,pr.reason FROM payment_refunds pr JOIN invoices i ON i.id=pr.invoice_id JOIN properties p ON p.id=i.property_id WHERE p.owner_id=? AND pr.status='posted' ORDER BY pr.refund_date DESC,pr.id DESC`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var v RefundView
		if err := rows.Scan(&v.ID, &v.Number, &v.Invoice, &v.Amount, &v.Date, &v.Reason); err != nil {
			rows.Close()
			a.serverError(w, r, err)
			return
		}
		page.Refunds = append(page.Refunds, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(r.Context(), `SELECT tc.id,tc.amount,tc.remaining_amount,u.name,p.name,tc.source_type,tc.status,tc.created_at FROM tenant_credits tc JOIN users u ON u.id=tc.tenant_id JOIN properties p ON p.id=tc.property_id WHERE p.owner_id=? ORDER BY tc.created_at DESC,tc.id DESC`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var v TenantCreditView
		if err := rows.Scan(&v.ID, &v.Amount, &v.Remaining, &v.Tenant, &v.Property, &v.Source, &v.Status, &v.CreatedAt); err != nil {
			rows.Close()
			a.serverError(w, r, err)
			return
		}
		page.Credits = append(page.Credits, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(r.Context(), `SELECT tcr.id,tcr.refund_number,u.name,p.name,tcr.amount,tcr.refund_date,tcr.reason FROM tenant_credit_refunds tcr JOIN users u ON u.id=tcr.tenant_id JOIN properties p ON p.id=tcr.property_id WHERE tcr.owner_id=? AND tcr.status='posted' ORDER BY tcr.refund_date DESC,tcr.id DESC`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var v TenantCreditRefundView
		if err := rows.Scan(&v.ID, &v.Number, &v.Tenant, &v.Property, &v.Amount, &v.Date, &v.Reason); err != nil {
			rows.Close()
			a.serverError(w, r, err)
			return
		}
		page.CreditRefunds = append(page.CreditRefunds, v)
	}
	rows.Close()
	data := PageData{Title: "Verifikasi Pembayaran — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page}
	if r.URL.Query().Get("berhasil") != "" {
		data.Success = "Bukti pembayaran berhasil diproses."
	}
	a.render(w, http.StatusOK, "owner_payments", data)
}

func (a *App) ownerRefundTenantCredit(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20)
	if err := r.ParseMultipartForm(8 << 20); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	creditID, err := parseID(r.PathValue("id"))
	bankID, bankErr := parseID(r.FormValue("bank_account_id"))
	date, dateErr := parseBusinessDate(r.FormValue("refund_date"))
	amount, _ := strconv.ParseInt(r.FormValue("amount"), 10, 64)
	reason := strings.TrimSpace(r.FormValue("reason"))
	if err != nil || bankErr != nil || dateErr != nil || amount <= 0 || len(reason) < 5 {
		http.Error(w, "Data refund kredit tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	upload, err := a.savePrivateUpload(r, "proof", "tenant-credit-refunds", 5<<20, true)
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
	var tenantID, propertyID, remaining int64
	var status string
	err = tx.QueryRowContext(r.Context(), `SELECT tc.tenant_id,tc.property_id,tc.remaining_amount,tc.status FROM tenant_credits tc JOIN properties p ON p.id=tc.property_id WHERE tc.id=? AND p.owner_id=? FOR UPDATE`, creditID, session.User.ID).Scan(&tenantID, &propertyID, &remaining, &status)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if status != "available" || remaining <= 0 || amount != remaining {
		http.Error(w, "Refund harus sebesar seluruh sisa kredit yang masih tersedia.", http.StatusConflict)
		return
	}
	bankLedger, err := financialAccountLedgerID(r.Context(), tx, session.User.ID, "bank", bankID, true)
	if err != nil {
		http.Error(w, "Rekening refund tidak ditemukan.", http.StatusNotFound)
		return
	}
	balance, err := ledgerBalance(r.Context(), tx, bankLedger)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if balance < amount {
		http.Error(w, "Saldo pembukuan rekening tidak mencukupi.", http.StatusConflict)
		return
	}
	creditLedger, err := systemAccountID(r.Context(), tx, session.User.ID, "tenant_credit")
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	number, _ := referenceNumber("RFD-KRD")
	result, err := tx.ExecContext(r.Context(), `INSERT INTO tenant_credit_refunds (refund_number,owner_id,tenant_id,property_id,tenant_credit_id,bank_account_id,amount,refund_date,reason,proof_storage_path,created_by) VALUES (?,?,?,?,?,?,?,?,?,?,?)`, number, session.User.ID, tenantID, propertyID, creditID, bankID, amount, date, reason, upload.RelativePath, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	refundID, _ := result.LastInsertId()
	_, err = postJournalTx(r.Context(), tx, JournalInput{OwnerID: session.User.ID, PropertyID: sql.NullInt64{Int64: propertyID, Valid: true}, Date: date, Description: "Refund kredit penghuni " + number, SourceType: "tenant_credit_refund", SourceID: refundID, EventType: "posted", ActorID: session.User.ID, Lines: []JournalLineInput{{AccountID: creditLedger, Debit: amount}, {AccountID: bankLedger, Credit: amount}}})
	if errors.Is(err, ErrClosedPeriod) {
		http.Error(w, "Tanggal refund berada pada periode tertutup.", http.StatusConflict)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if _, err = tx.ExecContext(r.Context(), `UPDATE tenant_credits SET remaining_amount=0,status='refunded' WHERE id=?`, creditID); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := createNotificationTx(r.Context(), tx, tenantID, fmt.Sprintf("tenant_credit.refunded.%d", refundID), "Kredit dikembalikan", "Refund kredit "+number+" telah diposting: "+reason, "/penghuni/tagihan"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "tenant_credit.refund", "tenant_credit_refund", refundID, nil, map[string]any{"tenant_credit_id": creditID, "amount": amount, "reason": reason}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/pembayaran?berhasil=refund-kredit", http.StatusSeeOther)
}

func (a *App) ownerVerifyPayment(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	submissionID, err := parseID(r.PathValue("id"))
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
	var tenantID, bankID, amount int64
	var date time.Time
	var status string
	err = tx.QueryRowContext(r.Context(), `SELECT ps.tenant_id,ps.bank_account_id,ps.amount,ps.transfer_date,ps.status FROM payment_submissions ps JOIN bank_accounts ba ON ba.id=ps.bank_account_id WHERE ps.id=? AND ba.owner_id=? FOR UPDATE`, submissionID, session.User.ID).Scan(&tenantID, &bankID, &amount, &date, &status)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if status != "pending_verification" {
		http.Error(w, "Bukti sudah diproses.", http.StatusConflict)
		return
	}
	var invoiceID, requested, propertyID, total, paid int64
	var invoiceStatus string
	var partial bool
	err = tx.QueryRowContext(r.Context(), `SELECT i.id,psa.requested_amount,i.property_id,i.total_amount,i.paid_amount,i.status,COALESCE(l.partial_payment_allowed,0) FROM payment_submission_allocations psa JOIN invoices i ON i.id=psa.invoice_id LEFT JOIN leases l ON l.id=i.lease_id WHERE psa.payment_submission_id=? AND i.tenant_id=? FOR UPDATE`, submissionID, tenantID).Scan(&invoiceID, &requested, &propertyID, &total, &paid, &invoiceStatus, &partial)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	outstanding := total - paid
	allocation := amount
	if allocation > outstanding {
		allocation = outstanding
	}
	if allocation < outstanding && !partial {
		http.Error(w, "Nominal kurang dan kontrak tidak mengizinkan pembayaran sebagian. Tolak bukti atau koreksi kebijakan secara terdokumentasi.", http.StatusConflict)
		return
	}
	bankLedger, err := financialAccountLedgerID(r.Context(), tx, session.User.ID, "bank", bankID, true)
	if err != nil {
		http.Error(w, "Rekening belum terhubung ke buku besar.", http.StatusConflict)
		return
	}
	arID, err := systemAccountID(r.Context(), tx, session.User.ID, "accounts_receivable")
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	excess := amount - allocation
	lines := []JournalLineInput{{AccountID: bankLedger, Debit: amount}, {AccountID: arID, Credit: allocation}}
	if excess > 0 {
		creditID, err := systemAccountID(r.Context(), tx, session.User.ID, "tenant_credit")
		if err != nil {
			a.serverError(w, r, err)
			return
		}
		lines = append(lines, JournalLineInput{AccountID: creditID, Credit: excess})
	}
	paymentNumber, _ := referenceNumber("PAY")
	receipt, _ := referenceNumber("KWT")
	result, err := tx.ExecContext(r.Context(), `INSERT INTO payments (payment_number,submission_id,tenant_id,bank_account_id,amount,payment_date,verified_by,verified_at,receipt_number) VALUES (?,?,?,?,?,?,?,UTC_TIMESTAMP(6),?)`, paymentNumber, submissionID, tenantID, bankID, amount, date, session.User.ID, receipt)
	if err != nil {
		if isDuplicate(err) {
			http.Error(w, "Pembayaran sudah pernah diverifikasi.", http.StatusConflict)
			return
		}
		a.serverError(w, r, err)
		return
	}
	paymentID, _ := result.LastInsertId()
	_, err = tx.ExecContext(r.Context(), `INSERT INTO payment_allocations (payment_id,invoice_id,amount) VALUES (?,?,?)`, paymentID, invoiceID, allocation)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	newStatus := "partially_paid"
	if paid+allocation == total {
		newStatus = "paid"
	}
	_, err = tx.ExecContext(r.Context(), `UPDATE invoices SET paid_amount=paid_amount+?,status=? WHERE id=?`, allocation, newStatus, invoiceID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if excess > 0 {
		_, err = tx.ExecContext(r.Context(), `INSERT INTO tenant_credits (tenant_id,property_id,source_type,source_id,amount,remaining_amount) VALUES (?,?,'payment',?,?,?)`, tenantID, propertyID, paymentID, excess, excess)
		if err != nil {
			a.serverError(w, r, err)
			return
		}
	}
	_, err = postJournalTx(r.Context(), tx, JournalInput{OwnerID: session.User.ID, PropertyID: sql.NullInt64{Int64: propertyID, Valid: true}, Date: date, Description: "Pembayaran terverifikasi " + paymentNumber, SourceType: "payment", SourceID: paymentID, EventType: "verified", ActorID: session.User.ID, Lines: lines})
	if err != nil {
		if errors.Is(err, ErrClosedPeriod) {
			http.Error(w, "Tanggal transfer berada di periode tertutup. Tolak dan minta bukti dengan tanggal koreksi pada periode terbuka.", http.StatusConflict)
			return
		}
		a.serverError(w, r, err)
		return
	}
	_, err = tx.ExecContext(r.Context(), `UPDATE payment_submissions SET status='verified',reviewed_by=?,reviewed_at=UTC_TIMESTAMP(6) WHERE id=?`, session.User.ID, submissionID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := createNotificationTx(r.Context(), tx, tenantID, fmt.Sprintf("payment.verified.%d", paymentID), "Pembayaran terverifikasi", "Pembayaran "+paymentNumber+" telah diverifikasi. Kuitansi tersedia.", "/penghuni/kuitansi/"+strconv.FormatInt(paymentID, 10)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "payment.verify", "payment", paymentID, map[string]any{"submission_status": status}, map[string]any{"amount": amount, "invoice_id": invoiceID}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/pembayaran?berhasil=verifikasi", http.StatusSeeOther)
}

func (a *App) ownerRejectPayment(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	id, err := parseID(r.PathValue("id"))
	reason := strings.TrimSpace(r.FormValue("reason"))
	if err != nil || len(reason) < 5 {
		http.Error(w, "Alasan minimal 5 karakter.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var tenantID int64
	err = tx.QueryRowContext(r.Context(), `SELECT ps.tenant_id FROM payment_submissions ps JOIN bank_accounts ba ON ba.id=ps.bank_account_id WHERE ps.id=? AND ba.owner_id=? AND ps.status='pending_verification'`, id, session.User.ID).Scan(&tenantID)
	if err != nil {
		http.Error(w, "Bukti tidak ditemukan atau sudah diproses.", http.StatusConflict)
		return
	}
	_, err = tx.ExecContext(r.Context(), `UPDATE payment_submissions SET status='rejected',reviewed_by=?,reviewed_at=UTC_TIMESTAMP(6),rejection_reason=? WHERE id=?`, session.User.ID, reason, id)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := createNotificationTx(r.Context(), tx, tenantID, fmt.Sprintf("payment.rejected.%d", id), "Bukti pembayaran ditolak", reason, "/penghuni/tagihan"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/pembayaran?berhasil=tolak", http.StatusSeeOther)
}

func (a *App) ownerRefundPayment(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20)
	if err := r.ParseMultipartForm(8 << 20); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	paymentID, err := parseID(r.PathValue("id"))
	bankID, e1 := parseID(r.FormValue("bank_account_id"))
	date, e2 := parseBusinessDate(r.FormValue("refund_date"))
	amount, _ := strconv.ParseInt(r.FormValue("amount"), 10, 64)
	reason := strings.TrimSpace(r.FormValue("reason"))
	if err != nil || e1 != nil || e2 != nil || amount <= 0 || len(reason) < 5 {
		http.Error(w, "Data refund tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	upload, err := a.savePrivateUpload(r, "proof", "payment-refunds", 5<<20, true)
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
	var tenantID, invoiceID, leaseID, propertyID, allocated int64
	var invoiceType string
	err = tx.QueryRowContext(r.Context(), `SELECT p.tenant_id,pa.invoice_id,COALESCE(i.lease_id,0),i.property_id,pa.amount,i.invoice_type FROM payments p JOIN payment_allocations pa ON pa.payment_id=p.id JOIN invoices i ON i.id=pa.invoice_id JOIN properties pr ON pr.id=i.property_id WHERE p.id=? AND pr.owner_id=? FOR UPDATE`, paymentID, session.User.ID).Scan(&tenantID, &invoiceID, &leaseID, &propertyID, &allocated, &invoiceType)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	var already int64
	_ = tx.QueryRowContext(r.Context(), `SELECT COALESCE(SUM(amount),0) FROM payment_refunds WHERE payment_id=? AND status='posted'`, paymentID).Scan(&already)
	if amount != allocated-already {
		http.Error(w, "Refund pembayaran harus sebesar seluruh alokasi yang belum direfund agar tagihan dan jurnal tetap konsisten.", http.StatusConflict)
		return
	}
	bankLedger, err := financialAccountLedgerID(r.Context(), tx, session.User.ID, "bank", bankID, true)
	if err != nil {
		http.Error(w, "Rekening refund tidak ditemukan.", http.StatusNotFound)
		return
	}
	balance, _ := ledgerBalance(r.Context(), tx, bankLedger)
	if balance < amount {
		http.Error(w, "Saldo pembukuan rekening tidak mencukupi.", http.StatusConflict)
		return
	}
	lines := []JournalLineInput{}
	if invoiceType == "rent" {
		unearnedAmount, reduceErr := a.reduceRevenueScheduleTx(r.Context(), tx, leaseID, amount)
		if reduceErr != nil {
			a.serverError(w, r, reduceErr)
			return
		}
		recognizedAmount := amount - unearnedAmount
		if unearnedAmount > 0 {
			account, err := systemAccountID(r.Context(), tx, session.User.ID, "unearned_rent")
			if err != nil {
				a.serverError(w, r, err)
				return
			}
			lines = append(lines, JournalLineInput{AccountID: account, Debit: unearnedAmount})
		}
		if recognizedAmount > 0 {
			account, err := systemAccountID(r.Context(), tx, session.User.ID, "rent_revenue")
			if err != nil {
				a.serverError(w, r, err)
				return
			}
			lines = append(lines, JournalLineInput{AccountID: account, Debit: recognizedAmount})
		}
	} else if invoiceType == "damage" {
		account, err := systemAccountID(r.Context(), tx, session.User.ID, "damage_recovery")
		if err != nil {
			a.serverError(w, r, err)
			return
		}
		lines = append(lines, JournalLineInput{AccountID: account, Debit: amount})
	} else {
		http.Error(w, "Jenis tagihan ini belum mendukung refund.", http.StatusConflict)
		return
	}
	lines = append(lines, JournalLineInput{AccountID: bankLedger, Credit: amount})
	number, _ := referenceNumber("RFD")
	result, err := tx.ExecContext(r.Context(), `INSERT INTO payment_refunds (refund_number,payment_id,invoice_id,bank_account_id,amount,refund_date,reason,proof_storage_path,created_by) VALUES (?,?,?,?,?,?,?,?,?)`, number, paymentID, invoiceID, bankID, amount, date, reason, upload.RelativePath, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	refundID, _ := result.LastInsertId()
	_, err = postJournalTx(r.Context(), tx, JournalInput{OwnerID: session.User.ID, PropertyID: sql.NullInt64{Int64: propertyID, Valid: true}, Date: date, Description: "Refund pembayaran " + number, SourceType: "payment_refund", SourceID: refundID, EventType: "posted", ActorID: session.User.ID, Lines: lines})
	if err != nil {
		if errors.Is(err, ErrClosedPeriod) {
			http.Error(w, "Tanggal refund berada pada periode tertutup.", http.StatusConflict)
			return
		}
		a.serverError(w, r, err)
		return
	}
	_, err = tx.ExecContext(r.Context(), `UPDATE invoices SET status='refunded' WHERE id=?`, invoiceID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := createNotificationTx(r.Context(), tx, tenantID, fmt.Sprintf("payment.refund.%d", refundID), "Refund diproses", "Refund "+number+" telah diposting: "+reason, "/penghuni/tagihan"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "payment.refund", "payment_refund", refundID, nil, map[string]any{"payment_id": paymentID, "amount": amount, "reason": reason}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/pembayaran?berhasil=refund", http.StatusSeeOther)
}

func (a *App) reduceRevenueScheduleTx(ctx context.Context, tx *sql.Tx, leaseID, amount int64) (int64, error) {
	if leaseID <= 0 {
		return 0, nil
	}
	rows, err := tx.QueryContext(ctx, `SELECT id,amount FROM revenue_recognition_schedules WHERE lease_id=? AND status='scheduled' ORDER BY recognition_date DESC,id DESC FOR UPDATE`, leaseID)
	if err != nil {
		return 0, err
	}
	type item struct{ id, amount int64 }
	var items []item
	for rows.Next() {
		var v item
		if err := rows.Scan(&v.id, &v.amount); err != nil {
			rows.Close()
			return 0, err
		}
		items = append(items, v)
	}
	rows.Close()
	remaining := amount
	var reduced int64
	for _, v := range items {
		if remaining <= 0 {
			break
		}
		take := v.amount
		if take > remaining {
			take = remaining
		}
		if take == v.amount {
			_, err = tx.ExecContext(ctx, `UPDATE revenue_recognition_schedules SET status='cancelled' WHERE id=?`, v.id)
		} else {
			_, err = tx.ExecContext(ctx, `UPDATE revenue_recognition_schedules SET amount=amount-? WHERE id=?`, take, v.id)
		}
		if err != nil {
			return reduced, err
		}
		reduced += take
		remaining -= take
	}
	return reduced, nil
}

func (a *App) tenantReceipt(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "tenant")
	if session == nil {
		return
	}
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	var paymentNumber, receipt, invoice, bank string
	var amount int64
	var date time.Time
	err = a.db.QueryRowContext(r.Context(), `SELECT p.payment_number,p.receipt_number,i.invoice_number,ba.bank_name,p.amount,p.payment_date FROM payments p JOIN payment_allocations pa ON pa.payment_id=p.id JOIN invoices i ON i.id=pa.invoice_id JOIN bank_accounts ba ON ba.id=p.bank_account_id WHERE p.id=? AND p.tenant_id=?`, id, session.User.ID).Scan(&paymentNumber, &receipt, &invoice, &bank, &amount, &date)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	a.render(w, http.StatusOK, "tenant_receipt", PageData{Title: "Kuitansi " + receipt, User: session.User, CSRFToken: session.CSRFToken, Payload: map[string]any{"PaymentNumber": paymentNumber, "Receipt": receipt, "Invoice": invoice, "Bank": bank, "Amount": amount, "Date": date}})
}

func (a *App) paymentBankOptions(ctx context.Context, tenantID int64) ([]BankPaymentOption, error) {
	rows, err := a.db.QueryContext(ctx, `SELECT DISTINCT ba.id,ba.bank_name,ba.account_name,ba.account_number_encrypted FROM bank_accounts ba JOIN properties p ON p.owner_id=ba.owner_id JOIN leases l ON l.property_id=p.id WHERE l.tenant_id=? AND ba.status='active' AND ba.is_accepting_payments=1 AND (ba.property_id IS NULL OR ba.property_id=l.property_id) ORDER BY ba.bank_name`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []BankPaymentOption
	for rows.Next() {
		var item BankPaymentOption
		var encrypted string
		if err := rows.Scan(&item.ID, &item.Bank, &item.Name, &encrypted); err != nil {
			return nil, err
		}
		number, err := decryptAccountNumber(a.cfg.HashKey, encrypted)
		if err != nil {
			continue
		}
		item.Number = number
		result = append(result, item)
	}
	return result, rows.Err()
}

func (a *App) paymentBankOptionsForProperty(ctx context.Context, propertyID int64) ([]BankPaymentOption, error) {
	rows, err := a.db.QueryContext(ctx, `SELECT ba.id,ba.bank_name,ba.account_name,ba.account_number_encrypted FROM bank_accounts ba JOIN properties p ON p.owner_id=ba.owner_id WHERE p.id=? AND ba.status='active' AND ba.is_accepting_payments=1 AND (ba.property_id IS NULL OR ba.property_id=p.id) ORDER BY ba.bank_name`, propertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []BankPaymentOption
	for rows.Next() {
		var item BankPaymentOption
		var encrypted string
		if err := rows.Scan(&item.ID, &item.Bank, &item.Name, &encrypted); err != nil {
			return nil, err
		}
		number, err := decryptAccountNumber(a.cfg.HashKey, encrypted)
		if err != nil {
			continue
		}
		item.Number = number
		result = append(result, item)
	}
	return result, rows.Err()
}

func createNotificationForOwnerTx(ctx context.Context, tx *sql.Tx, propertyID int64, eventKey, title, body, url string) error {
	var ownerID int64
	if err := tx.QueryRowContext(ctx, `SELECT owner_id FROM properties WHERE id=?`, propertyID).Scan(&ownerID); err != nil {
		return err
	}
	return createNotificationTx(ctx, tx, ownerID, eventKey, title, body, url)
}
func auditActor(session any) any {
	typed, ok := session.(*auth.Session)
	if !ok || typed == nil || typed.User == nil {
		return nil
	}
	return typed.User.ID
}
