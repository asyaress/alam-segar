package httpapp

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type FinanceAccountView struct {
	ID, LedgerID               int64
	Type, Name, Detail, Status string
	Balance                    int64
	Accepting                  bool
}
type COAView struct {
	ID                   int64
	Code, Name, Category string
	Balance              int64
}
type ExpenseView struct {
	ID                                                             int64
	Number, Property, Category, Vendor, Description, Ticket, Asset string
	Date                                                           time.Time
	Amount                                                         int64
}
type JournalView struct {
	ID                                  int64
	Number, Description, Source, Status string
	Date                                time.Time
	Debit, Credit                       int64
}
type ReconciliationView struct {
	ID                                     int64
	AccountType, AccountName, Notes        string
	Date                                   time.Time
	BookBalance, ActualBalance, Difference int64
}
type FinancePage struct {
	Accounts        []FinanceAccountView
	ExpenseAccounts []COAView
	Expenses        []ExpenseView
	COA             []COAView
	Journals        []JournalView
	Periods         []PeriodView
	Reconciliations []ReconciliationView
	Summary         ReportSummary
	Aging           ReceivableAging
	CashFlow        CashFlowSummary
	ReportFilters   ReportFilters
	Properties      []PropertyOption
	Tickets, Assets []LeaseOption
}
type PeriodView struct {
	ID         int64
	Start, End time.Time
	Status     string
	ClosedAt   sql.NullTime
}
type ReportSummary struct{ Assets, Liabilities, Equity, Revenue, Expenses, Profit, Receivables, Cash int64 }
type ReceivableAging struct{ Current, Days1To30, Days31To60, Days61To90, Over90 int64 }
type CashFlowSummary struct{ Operating, Investing, Financing, Net int64 }
type ReportFilters struct {
	PropertyID int64
	Start      string
	End        string
	ExportURL  string
}

func (a *App) ownerFinance(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	page, err := a.loadFinancePage(r.Context(), session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	data := PageData{Title: "Kas & Bank — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page}
	if r.URL.Query().Get("berhasil") != "" {
		data.Success = "Transaksi keuangan berhasil diposting."
	}
	a.render(w, http.StatusOK, "owner_finance", data)
}

func (a *App) loadFinancePage(ctx context.Context, ownerID int64) (FinancePage, error) {
	page := FinancePage{}
	rows, err := a.db.QueryContext(ctx, `SELECT ba.id,COALESCE(ba.ledger_account_id,0),'bank',CONCAT(ba.bank_name,' — ',ba.account_name),CONCAT('•••• ',ba.account_number_last4),ba.status,ba.is_accepting_payments,COALESCE(SUM(CASE WHEN je.status='posted' THEN CAST(jl.debit AS SIGNED)-CAST(jl.credit AS SIGNED) ELSE 0 END),0) FROM bank_accounts ba LEFT JOIN journal_lines jl ON jl.account_id=ba.ledger_account_id LEFT JOIN journal_entries je ON je.id=jl.journal_entry_id WHERE ba.owner_id=? GROUP BY ba.id ORDER BY ba.bank_name`, ownerID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v FinanceAccountView
		if err := rows.Scan(&v.ID, &v.LedgerID, &v.Type, &v.Name, &v.Detail, &v.Status, &v.Accepting, &v.Balance); err != nil {
			rows.Close()
			return page, err
		}
		page.Accounts = append(page.Accounts, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(ctx, `SELECT ca.id,ca.ledger_account_id,'cash',ca.name,CONCAT('Penanggung jawab: ',ca.custodian_name),ca.status,0,COALESCE(SUM(CASE WHEN je.status='posted' THEN CAST(jl.debit AS SIGNED)-CAST(jl.credit AS SIGNED) ELSE 0 END),0) FROM cash_accounts ca LEFT JOIN journal_lines jl ON jl.account_id=ca.ledger_account_id LEFT JOIN journal_entries je ON je.id=jl.journal_entry_id WHERE ca.owner_id=? GROUP BY ca.id ORDER BY ca.name`, ownerID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v FinanceAccountView
		if err := rows.Scan(&v.ID, &v.LedgerID, &v.Type, &v.Name, &v.Detail, &v.Status, &v.Accepting, &v.Balance); err != nil {
			rows.Close()
			return page, err
		}
		page.Accounts = append(page.Accounts, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(ctx, `
		SELECT r.id,r.account_type,
			CASE WHEN r.account_type='bank'
				THEN COALESCE((SELECT CONCAT(ba.bank_name,' — ',ba.account_name) FROM bank_accounts ba WHERE ba.id=r.account_id AND ba.owner_id=r.owner_id),'Rekening dihapus')
				ELSE COALESCE((SELECT ca.name FROM cash_accounts ca WHERE ca.id=r.account_id AND ca.owner_id=r.owner_id),'Kas dihapus') END,
			r.reconciliation_date,r.book_balance,r.actual_balance,r.difference_amount,r.notes
		FROM reconciliations r WHERE r.owner_id=?
		ORDER BY r.reconciliation_date DESC,r.id DESC LIMIT 20`, ownerID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v ReconciliationView
		if err := rows.Scan(&v.ID, &v.AccountType, &v.AccountName, &v.Date, &v.BookBalance, &v.ActualBalance, &v.Difference, &v.Notes); err != nil {
			rows.Close()
			return page, err
		}
		page.Reconciliations = append(page.Reconciliations, v)
	}
	if err := rows.Close(); err != nil {
		return page, err
	}
	return page, nil
}

func (a *App) ownerCreateBankAccount(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	bank := strings.TrimSpace(r.FormValue("bank_name"))
	name := strings.TrimSpace(r.FormValue("account_name"))
	number := strings.ReplaceAll(strings.TrimSpace(r.FormValue("account_number")), " ", "")
	propertyID := optionalID(r.FormValue("property_id"))
	if len(bank) < 2 || len(name) < 3 || len(number) < 6 || len(number) > 40 {
		http.Error(w, "Data rekening tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	encrypted, err := encryptAccountNumber(a.cfg.HashKey, number)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	if propertyID.Valid {
		var count int
		if err := tx.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM properties WHERE id=? AND owner_id=?`, propertyID.Int64, session.User.ID).Scan(&count); err != nil || count != 1 {
			http.Error(w, "Properti tidak valid.", http.StatusUnprocessableEntity)
			return
		}
	}
	code, _ := referenceNumber("BANK")
	result, err := tx.ExecContext(r.Context(), `INSERT INTO chart_of_accounts (owner_id,code,name,category,normal_balance) VALUES (?,?,?,'asset','debit')`, session.User.ID, code, "Bank "+bank+" "+name)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	ledgerID, _ := result.LastInsertId()
	result, err = tx.ExecContext(r.Context(), `INSERT INTO bank_accounts (owner_id,property_id,bank_name,account_name,account_number_encrypted,account_number_last4,ledger_account_id,opened_at,is_accepting_payments,status) VALUES (?,?,?,?,?,?,?,?,?,'active')`, session.User.ID, nullableInt64(propertyID), bank, name, encrypted, number[len(number)-4:], ledgerID, businessToday(a.location), r.FormValue("accepting") == "1")
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	id, _ := result.LastInsertId()
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "bank_account.create", "bank_account", id, nil, map[string]any{"bank": bank, "last4": number[len(number)-4:]}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/keuangan?berhasil=rekening", http.StatusSeeOther)
}

func (a *App) ownerCreateCashAccount(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	custodian := strings.TrimSpace(r.FormValue("custodian"))
	if len(name) < 3 || len(custodian) < 3 {
		http.Error(w, "Nama kas dan penanggung jawab wajib diisi.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	code, _ := referenceNumber("KAS")
	result, err := tx.ExecContext(r.Context(), `INSERT INTO chart_of_accounts (owner_id,code,name,category,normal_balance) VALUES (?,?,?,'asset','debit')`, session.User.ID, code, name)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	ledgerID, _ := result.LastInsertId()
	result, err = tx.ExecContext(r.Context(), `INSERT INTO cash_accounts (owner_id,name,custodian_name,ledger_account_id,opened_at) VALUES (?,?,?,?,?)`, session.User.ID, name, custodian, ledgerID, businessToday(a.location))
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	id, _ := result.LastInsertId()
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "cash_account.create", "cash_account", id, nil, map[string]any{"name": name, "custodian": custodian}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/keuangan?berhasil=kas", http.StatusSeeOther)
}

func (a *App) ownerPostOpeningBalance(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	typ, id, err := parseFinancialAccountRef(r.FormValue("account"))
	date, dateErr := parseBusinessDate(r.FormValue("balance_date"))
	amount, _ := strconv.ParseInt(r.FormValue("amount"), 10, 64)
	if err != nil || dateErr != nil || amount <= 0 {
		http.Error(w, "Data saldo awal tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	ledgerID, err := financialAccountLedgerID(r.Context(), tx, session.User.ID, typ, id, true)
	if err != nil {
		http.Error(w, "Akun keuangan tidak ditemukan.", http.StatusNotFound)
		return
	}
	equityID, err := systemAccountID(r.Context(), tx, session.User.ID, "opening_equity")
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	entryID, err := postJournalTx(r.Context(), tx, JournalInput{OwnerID: session.User.ID, Date: date, Description: "Saldo awal akun keuangan", SourceType: "opening_balance", SourceID: id, EventType: typ, ActorID: session.User.ID, Lines: []JournalLineInput{{AccountID: ledgerID, Debit: amount}, {AccountID: equityID, Credit: amount}}})
	if errors.Is(err, ErrClosedPeriod) {
		http.Error(w, "Tanggal berada pada periode tertutup.", http.StatusConflict)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	_, err = tx.ExecContext(r.Context(), `INSERT INTO opening_balances (owner_id,account_type,account_id,balance_date,amount,journal_entry_id,created_by) VALUES (?,?,?,?,?,?,?)`, session.User.ID, typ, id, date, amount, entryID, session.User.ID)
	if err != nil {
		if isDuplicate(err) {
			http.Error(w, "Saldo awal akun ini sudah pernah diposting.", http.StatusConflict)
			return
		}
		a.serverError(w, r, err)
		return
	}
	table := "bank_accounts"
	if typ == "cash" {
		table = "cash_accounts"
	}
	_, err = tx.ExecContext(r.Context(), `UPDATE `+table+` SET opening_balance_posted=1,opened_at=? WHERE id=?`, date, id)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/keuangan?berhasil=saldo", http.StatusSeeOther)
}

func (a *App) ownerInternalTransfer(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	fromType, fromID, e1 := parseFinancialAccountRef(r.FormValue("from_account"))
	toType, toID, e2 := parseFinancialAccountRef(r.FormValue("to_account"))
	date, e3 := parseBusinessDate(r.FormValue("transfer_date"))
	amount, _ := strconv.ParseInt(r.FormValue("amount"), 10, 64)
	if e1 != nil || e2 != nil || e3 != nil || amount <= 0 || (fromType == toType && fromID == toID) {
		http.Error(w, "Data transfer tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	fromLedger, err := financialAccountLedgerID(r.Context(), tx, session.User.ID, fromType, fromID, true)
	if err != nil {
		http.Error(w, "Akun asal tidak ditemukan.", http.StatusNotFound)
		return
	}
	toLedger, err := financialAccountLedgerID(r.Context(), tx, session.User.ID, toType, toID, true)
	if err != nil {
		http.Error(w, "Akun tujuan tidak ditemukan.", http.StatusNotFound)
		return
	}
	balance, err := ledgerBalance(r.Context(), tx, fromLedger)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if balance < amount {
		http.Error(w, "Saldo pembukuan akun asal tidak mencukupi.", http.StatusConflict)
		return
	}
	number, _ := referenceNumber("TRF")
	result, err := tx.ExecContext(r.Context(), `INSERT INTO internal_transfers (transfer_number,owner_id,transfer_date,from_account_type,from_account_id,to_account_type,to_account_id,amount,notes,created_by) VALUES (?,?,?,?,?,?,?,?,?,?)`, number, session.User.ID, date, fromType, fromID, toType, toID, amount, nullableString(strings.TrimSpace(r.FormValue("notes"))), session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	id, _ := result.LastInsertId()
	_, err = postJournalTx(r.Context(), tx, JournalInput{OwnerID: session.User.ID, Date: date, Description: "Transfer internal " + number, SourceType: "internal_transfer", SourceID: id, EventType: "posted", ActorID: session.User.ID, Lines: []JournalLineInput{{AccountID: toLedger, Debit: amount}, {AccountID: fromLedger, Credit: amount}}})
	if err != nil {
		if errors.Is(err, ErrClosedPeriod) {
			http.Error(w, "Tanggal berada pada periode tertutup.", http.StatusConflict)
			return
		}
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/keuangan?berhasil=transfer", http.StatusSeeOther)
}

func (a *App) ownerReconcileAccount(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	typ, accountID, idErr := parseFinancialAccountRef(r.FormValue("account"))
	date, dateErr := parseBusinessDate(r.FormValue("reconciliation_date"))
	actual, amountErr := strconv.ParseInt(r.FormValue("actual_balance"), 10, 64)
	notes := strings.TrimSpace(r.FormValue("notes"))
	if (typ != "bank" && typ != "cash") || idErr != nil || dateErr != nil || amountErr != nil || actual < 0 || len(notes) < 5 {
		http.Error(w, "Data rekonsiliasi tidak valid. Catatan minimal 5 karakter.", http.StatusUnprocessableEntity)
		return
	}
	if date.After(businessToday(a.location)) {
		http.Error(w, "Tanggal rekonsiliasi tidak boleh di masa depan.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	ledgerID, err := financialAccountLedgerID(r.Context(), tx, session.User.ID, typ, accountID, true)
	if err != nil {
		http.Error(w, "Akun keuangan tidak ditemukan.", http.StatusNotFound)
		return
	}
	book, err := ledgerBalance(r.Context(), tx, ledgerID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	result, err := tx.ExecContext(r.Context(), `INSERT INTO reconciliations
		(owner_id,account_type,account_id,reconciliation_date,book_balance,actual_balance,difference_amount,notes,created_by)
		VALUES (?,?,?,?,?,?,?,?,?)`, session.User.ID, typ, accountID, date, book, actual, actual-book, notes, session.User.ID)
	if err != nil {
		if isDuplicate(err) {
			http.Error(w, "Akun ini sudah direkonsiliasi pada tanggal tersebut.", http.StatusConflict)
			return
		}
		a.serverError(w, r, err)
		return
	}
	id, _ := result.LastInsertId()
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "finance.reconcile", "reconciliation", id, nil, map[string]any{"account_type": typ, "account_id": accountID, "book_balance": book, "actual_balance": actual, "difference": actual - book}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/keuangan?berhasil=rekonsiliasi", http.StatusSeeOther)
}

func (a *App) ownerExpenseList(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	page, err := a.loadExpensePage(r.Context(), session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	data := PageData{Title: "Pengeluaran — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page}
	if r.URL.Query().Get("berhasil") != "" {
		data.Success = "Pengeluaran dan jurnal berhasil diposting."
	}
	a.render(w, http.StatusOK, "owner_expenses", data)
}

func (a *App) loadExpensePage(ctx context.Context, ownerID int64) (FinancePage, error) {
	page, err := a.loadFinancePage(ctx, ownerID)
	if err != nil {
		return page, err
	}
	rows, err := a.db.QueryContext(ctx, `SELECT id,code,name,category,0 FROM chart_of_accounts WHERE owner_id=? AND (category='expense' OR system_key='fixed_assets') AND is_active=1 ORDER BY code`, ownerID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v COAView
		if err := rows.Scan(&v.ID, &v.Code, &v.Name, &v.Category, &v.Balance); err != nil {
			rows.Close()
			return page, err
		}
		page.ExpenseAccounts = append(page.ExpenseAccounts, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(ctx, `SELECT e.id,e.expense_number,COALESCE(p.name,'Semua properti'),coa.name,COALESCE(e.vendor_name,''),e.description,e.expense_date,e.amount,COALESCE(mt.ticket_number,''),COALESCE(a.asset_code,'') FROM expenses e LEFT JOIN properties p ON p.id=e.property_id JOIN chart_of_accounts coa ON coa.id=e.expense_account_id LEFT JOIN maintenance_tickets mt ON mt.id=e.maintenance_ticket_id LEFT JOIN assets a ON a.id=e.asset_id WHERE e.owner_id=? ORDER BY e.expense_date DESC,e.id DESC`, ownerID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v ExpenseView
		if err := rows.Scan(&v.ID, &v.Number, &v.Property, &v.Category, &v.Vendor, &v.Description, &v.Date, &v.Amount, &v.Ticket, &v.Asset); err != nil {
			rows.Close()
			return page, err
		}
		page.Expenses = append(page.Expenses, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(ctx, `SELECT id,name FROM properties WHERE owner_id=? AND deleted_at IS NULL ORDER BY name`, ownerID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v PropertyOption
		if err := rows.Scan(&v.ID, &v.Name); err != nil {
			rows.Close()
			return page, err
		}
		page.Properties = append(page.Properties, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(ctx, `SELECT mt.id,CONCAT(mt.ticket_number,' — ',mt.subject) FROM maintenance_tickets mt JOIN properties p ON p.id=mt.property_id WHERE p.owner_id=? ORDER BY FIELD(mt.status,'open','in_progress','waiting_for_tenant','resolved','closed'),mt.created_at DESC`, ownerID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v LeaseOption
		if err := rows.Scan(&v.ID, &v.Label); err != nil {
			rows.Close()
			return page, err
		}
		page.Tickets = append(page.Tickets, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(ctx, `SELECT a.id,CONCAT(a.asset_code,' — ',a.name) FROM assets a WHERE a.owner_id=? ORDER BY a.asset_code`, ownerID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v LeaseOption
		if err := rows.Scan(&v.ID, &v.Label); err != nil {
			rows.Close()
			return page, err
		}
		page.Assets = append(page.Assets, v)
	}
	rows.Close()
	return page, nil
}

func (a *App) ownerCreateExpense(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20)
	if err := r.ParseMultipartForm(8 << 20); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	date, e1 := parseBusinessDate(r.FormValue("expense_date"))
	amount, _ := strconv.ParseInt(r.FormValue("amount"), 10, 64)
	expenseAccount, e2 := parseID(r.FormValue("expense_account_id"))
	accountType, accountID, e3 := parseFinancialAccountRef(r.FormValue("payment_account"))
	description := strings.TrimSpace(r.FormValue("description"))
	propertyID := optionalID(r.FormValue("property_id"))
	ticketID := optionalID(r.FormValue("maintenance_ticket_id"))
	assetID := optionalID(r.FormValue("asset_id"))
	if e1 != nil || e2 != nil || e3 != nil || amount <= 0 || len(description) < 5 {
		http.Error(w, "Data pengeluaran tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	upload, err := a.savePrivateUpload(r, "proof", "expenses", 5<<20, true)
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
	if ticketID.Valid {
		var ticketProperty int64
		if err := tx.QueryRowContext(r.Context(), `SELECT mt.property_id FROM maintenance_tickets mt JOIN properties p ON p.id=mt.property_id WHERE mt.id=? AND p.owner_id=?`, ticketID.Int64, session.User.ID).Scan(&ticketProperty); err != nil {
			http.Error(w, "Tiket pemeliharaan tidak valid.", http.StatusUnprocessableEntity)
			return
		}
		if propertyID.Valid && propertyID.Int64 != ticketProperty {
			http.Error(w, "Properti pengeluaran tidak sesuai dengan tiket.", http.StatusUnprocessableEntity)
			return
		}
		propertyID = sql.NullInt64{Int64: ticketProperty, Valid: true}
	}
	if assetID.Valid {
		var assetProperty int64
		if err := tx.QueryRowContext(r.Context(), `SELECT property_id FROM assets WHERE id=? AND owner_id=?`, assetID.Int64, session.User.ID).Scan(&assetProperty); err != nil {
			http.Error(w, "Aset terdampak tidak valid.", http.StatusUnprocessableEntity)
			return
		}
		if propertyID.Valid && propertyID.Int64 != assetProperty {
			http.Error(w, "Properti pengeluaran tidak sesuai dengan aset.", http.StatusUnprocessableEntity)
			return
		}
		propertyID = sql.NullInt64{Int64: assetProperty, Valid: true}
	}
	var count int
	if err := tx.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM chart_of_accounts WHERE id=? AND owner_id=? AND (category='expense' OR system_key='fixed_assets')`, expenseAccount, session.User.ID).Scan(&count); err != nil || count != 1 {
		http.Error(w, "Akun beban atau pembelian aset tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	paymentLedger, err := financialAccountLedgerID(r.Context(), tx, session.User.ID, accountType, accountID, true)
	if err != nil {
		http.Error(w, "Akun pembayaran tidak ditemukan.", http.StatusNotFound)
		return
	}
	if accountType == "cash" {
		balance, _ := ledgerBalance(r.Context(), tx, paymentLedger)
		if balance < amount {
			http.Error(w, "Saldo kas tidak mencukupi.", http.StatusConflict)
			return
		}
	}
	number, _ := referenceNumber("EXP")
	result, err := tx.ExecContext(r.Context(), `INSERT INTO expenses (expense_number,owner_id,property_id,maintenance_ticket_id,asset_id,expense_account_id,payment_account_type,payment_account_id,expense_date,amount,vendor_name,description,proof_storage_path,proof_sha256,created_by) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, number, session.User.ID, nullableInt64(propertyID), nullableInt64(ticketID), nullableInt64(assetID), expenseAccount, accountType, accountID, date, amount, nullableString(strings.TrimSpace(r.FormValue("vendor"))), description, upload.RelativePath, upload.SHA256, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	expenseID, _ := result.LastInsertId()
	_, err = postJournalTx(r.Context(), tx, JournalInput{OwnerID: session.User.ID, PropertyID: propertyID, Date: date, Description: "Pengeluaran " + number + " — " + description, SourceType: "expense", SourceID: expenseID, EventType: "posted", ActorID: session.User.ID, Lines: []JournalLineInput{{AccountID: expenseAccount, Debit: amount}, {AccountID: paymentLedger, Credit: amount}}})
	if err != nil {
		if errors.Is(err, ErrClosedPeriod) {
			http.Error(w, "Tanggal berada pada periode tertutup.", http.StatusConflict)
			return
		}
		a.serverError(w, r, err)
		return
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "expense.post", "expense", expenseID, nil, map[string]any{"number": number, "amount": amount}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/pengeluaran?berhasil=posted", http.StatusSeeOther)
}

func (a *App) ownerAccounting(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	page, err := a.loadAccountingPage(r.Context(), session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	data := PageData{Title: "Akuntansi — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page}
	if r.URL.Query().Get("berhasil") != "" {
		data.Success = "Periode berhasil ditutup dan dikunci."
	}
	a.render(w, http.StatusOK, "owner_accounting", data)
}

func (a *App) loadAccountingPage(ctx context.Context, ownerID int64) (FinancePage, error) {
	page := FinancePage{}
	rows, err := a.db.QueryContext(ctx, `SELECT coa.id,coa.code,coa.name,coa.category,CASE WHEN coa.category IN ('asset','expense') THEN COALESCE(SUM(CASE WHEN je.status='posted' THEN CAST(jl.debit AS SIGNED)-CAST(jl.credit AS SIGNED) ELSE 0 END),0) ELSE COALESCE(SUM(CASE WHEN je.status='posted' THEN CAST(jl.credit AS SIGNED)-CAST(jl.debit AS SIGNED) ELSE 0 END),0) END balance FROM chart_of_accounts coa LEFT JOIN journal_lines jl ON jl.account_id=coa.id LEFT JOIN journal_entries je ON je.id=jl.journal_entry_id WHERE coa.owner_id=? GROUP BY coa.id ORDER BY coa.code`, ownerID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v COAView
		if err := rows.Scan(&v.ID, &v.Code, &v.Name, &v.Category, &v.Balance); err != nil {
			rows.Close()
			return page, err
		}
		page.COA = append(page.COA, v)
		switch v.Category {
		case "asset":
			page.Summary.Assets += v.Balance
		case "liability":
			page.Summary.Liabilities += v.Balance
		case "equity":
			page.Summary.Equity += v.Balance
		case "revenue":
			page.Summary.Revenue += v.Balance
		case "expense":
			page.Summary.Expenses += v.Balance
		}
	}
	rows.Close()
	page.Summary.Profit = page.Summary.Revenue - page.Summary.Expenses
	rows, err = a.db.QueryContext(ctx, `SELECT je.id,je.entry_number,je.description,CONCAT(je.source_type,'/',je.event_type),je.status,je.entry_date,SUM(jl.debit),SUM(jl.credit) FROM journal_entries je JOIN journal_lines jl ON jl.journal_entry_id=je.id WHERE je.owner_id=? GROUP BY je.id ORDER BY je.entry_date DESC,je.id DESC LIMIT 100`, ownerID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v JournalView
		if err := rows.Scan(&v.ID, &v.Number, &v.Description, &v.Source, &v.Status, &v.Date, &v.Debit, &v.Credit); err != nil {
			rows.Close()
			return page, err
		}
		page.Journals = append(page.Journals, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(ctx, `SELECT id,period_start,period_end,status,closed_at FROM accounting_periods WHERE owner_id=? ORDER BY period_start DESC`, ownerID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v PeriodView
		if err := rows.Scan(&v.ID, &v.Start, &v.End, &v.Status, &v.ClosedAt); err != nil {
			rows.Close()
			return page, err
		}
		page.Periods = append(page.Periods, v)
	}
	rows.Close()
	return page, nil
}

func (a *App) ownerClosePeriod(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	month, err := time.ParseInLocation("2006-01", r.FormValue("month"), time.UTC)
	notes := strings.TrimSpace(r.FormValue("notes"))
	if err != nil || len(notes) < 5 {
		http.Error(w, "Bulan dan catatan pemeriksaan wajib diisi.", http.StatusUnprocessableEntity)
		return
	}
	start := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, -1)
	today := businessToday(a.location)
	if !end.Before(today) {
		http.Error(w, "Periode hanya dapat ditutup setelah bulan tersebut berakhir.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	if err := a.recognizeRevenueThroughTx(r.Context(), tx, session.User.ID, end, session.User.ID); err != nil {
		a.serverError(w, r, err)
		return
	}
	var pending int
	if err := tx.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM payment_submissions ps JOIN bank_accounts ba ON ba.id=ps.bank_account_id WHERE ba.owner_id=? AND ps.status='pending_verification' AND ps.transfer_date<=?`, session.User.ID, end).Scan(&pending); err != nil {
		a.serverError(w, r, err)
		return
	}
	if pending > 0 {
		http.Error(w, "Masih ada bukti transfer periode ini yang belum diverifikasi.", http.StatusConflict)
		return
	}
	var unbalanced int
	if err := tx.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM (SELECT je.id FROM journal_entries je JOIN journal_lines jl ON jl.journal_entry_id=je.id WHERE je.owner_id=? AND je.entry_date BETWEEN ? AND ? GROUP BY je.id HAVING SUM(jl.debit)<>SUM(jl.credit)) x`, session.User.ID, start, end).Scan(&unbalanced); err != nil {
		a.serverError(w, r, err)
		return
	}
	if unbalanced > 0 {
		http.Error(w, "Ada jurnal tidak seimbang.", http.StatusConflict)
		return
	}
	_, err = tx.ExecContext(r.Context(), `INSERT INTO accounting_periods (owner_id,period_start,period_end,status,closed_by,closed_at,close_notes) VALUES (?,?,?,'closed',?,UTC_TIMESTAMP(6),?)`, session.User.ID, start, end, session.User.ID, notes)
	if err != nil {
		if isDuplicate(err) {
			http.Error(w, "Periode sudah pernah ditutup.", http.StatusConflict)
			return
		}
		a.serverError(w, r, err)
		return
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "accounting.close_period", "accounting_period", int64(start.Year()*100+int(start.Month())), nil, map[string]any{"start": start, "end": end, "notes": notes}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/akuntansi?berhasil=tutup", http.StatusSeeOther)
}

func (a *App) recognizeRevenueThroughTx(ctx context.Context, tx *sql.Tx, ownerID int64, through time.Time, actorID int64) error {
	rows, err := tx.QueryContext(ctx, `SELECT rrs.id,rrs.property_id,rrs.recognition_date,rrs.amount FROM revenue_recognition_schedules rrs JOIN leases l ON l.id=rrs.lease_id JOIN properties p ON p.id=l.property_id WHERE p.owner_id=? AND rrs.status='scheduled' AND rrs.recognition_date<=? ORDER BY rrs.recognition_date,rrs.id FOR UPDATE`, ownerID, through)
	if err != nil {
		return err
	}
	type item struct {
		id, property, amount int64
		date                 time.Time
	}
	var items []item
	for rows.Next() {
		var v item
		if err := rows.Scan(&v.id, &v.property, &v.date, &v.amount); err != nil {
			rows.Close()
			return err
		}
		items = append(items, v)
	}
	rows.Close()
	unearned, err := systemAccountID(ctx, tx, ownerID, "unearned_rent")
	if err != nil {
		return err
	}
	revenue, err := systemAccountID(ctx, tx, ownerID, "rent_revenue")
	if err != nil {
		return err
	}
	for _, v := range items {
		entryID, err := postJournalTx(ctx, tx, JournalInput{OwnerID: ownerID, PropertyID: sql.NullInt64{Int64: v.property, Valid: true}, Date: v.date, Description: "Pengakuan pendapatan sewa", SourceType: "revenue_schedule", SourceID: v.id, EventType: "recognized", ActorID: actorID, Lines: []JournalLineInput{{AccountID: unearned, Debit: v.amount}, {AccountID: revenue, Credit: v.amount}}})
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE revenue_recognition_schedules SET status='posted',journal_entry_id=? WHERE id=? AND status='scheduled'`, entryID, v.id); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) ownerReports(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	localNow := time.Now().In(a.location)
	start := time.Date(localNow.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	end := businessToday(a.location)
	if value := strings.TrimSpace(r.URL.Query().Get("start")); value != "" {
		parsed, err := parseBusinessDate(value)
		if err != nil {
			http.Error(w, "Tanggal awal tidak valid.", http.StatusUnprocessableEntity)
			return
		}
		start = parsed
	}
	if value := strings.TrimSpace(r.URL.Query().Get("end")); value != "" {
		parsed, err := parseBusinessDate(value)
		if err != nil {
			http.Error(w, "Tanggal akhir tidak valid.", http.StatusUnprocessableEntity)
			return
		}
		end = parsed
	}
	propertyID, _ := strconv.ParseInt(r.URL.Query().Get("property_id"), 10, 64)
	if end.Before(start) {
		http.Error(w, "Tanggal akhir harus sama atau setelah tanggal awal.", http.StatusUnprocessableEntity)
		return
	}
	page, err := a.loadReportPage(r.Context(), session.User.ID, propertyID, start, end)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	a.render(w, http.StatusOK, "owner_reports", PageData{Title: "Laporan — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page})
}

func (a *App) loadReportPage(ctx context.Context, ownerID, propertyID int64, start, end time.Time) (FinancePage, error) {
	page := FinancePage{ReportFilters: ReportFilters{PropertyID: propertyID, Start: start.Format("2006-01-02"), End: end.Format("2006-01-02")}}
	rows, err := a.db.QueryContext(ctx, `SELECT id,name FROM properties WHERE owner_id=? AND deleted_at IS NULL ORDER BY name`, ownerID)
	if err != nil {
		return page, err
	}
	foundProperty := propertyID == 0
	for rows.Next() {
		var v PropertyOption
		if err := rows.Scan(&v.ID, &v.Name); err != nil {
			rows.Close()
			return page, err
		}
		page.Properties = append(page.Properties, v)
		if v.ID == propertyID {
			foundProperty = true
		}
	}
	rows.Close()
	if !foundProperty {
		return page, errors.New("properti laporan bukan milik owner")
	}
	page.ReportFilters.ExportURL = fmt.Sprintf("/pemilik/laporan/jurnal.csv?property_id=%d&start=%s&end=%s", propertyID, page.ReportFilters.Start, page.ReportFilters.End)
	rows, err = a.db.QueryContext(ctx, `
		SELECT coa.category,
			CASE WHEN coa.category IN ('asset','expense')
				THEN COALESCE(SUM(CASE WHEN je.status='posted' AND je.entry_date<=? AND (?=0 OR je.property_id=?) THEN CAST(jl.debit AS SIGNED)-CAST(jl.credit AS SIGNED) ELSE 0 END),0)
				ELSE COALESCE(SUM(CASE WHEN je.status='posted' AND je.entry_date<=? AND (?=0 OR je.property_id=?) THEN CAST(jl.credit AS SIGNED)-CAST(jl.debit AS SIGNED) ELSE 0 END),0) END
		FROM chart_of_accounts coa
		LEFT JOIN journal_lines jl ON jl.account_id=coa.id LEFT JOIN journal_entries je ON je.id=jl.journal_entry_id
		WHERE coa.owner_id=? GROUP BY coa.category`, end, propertyID, propertyID, end, propertyID, propertyID, ownerID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var category string
		var balance int64
		if err := rows.Scan(&category, &balance); err != nil {
			rows.Close()
			return page, err
		}
		switch category {
		case "asset":
			page.Summary.Assets = balance
		case "liability":
			page.Summary.Liabilities = balance
		case "equity":
			page.Summary.Equity = balance
		}
	}
	rows.Close()
	rows, err = a.db.QueryContext(ctx, `
		SELECT coa.category,CASE WHEN coa.category='expense'
			THEN COALESCE(SUM(CAST(jl.debit AS SIGNED)-CAST(jl.credit AS SIGNED)),0) ELSE COALESCE(SUM(CAST(jl.credit AS SIGNED)-CAST(jl.debit AS SIGNED)),0) END
		FROM chart_of_accounts coa JOIN journal_lines jl ON jl.account_id=coa.id JOIN journal_entries je ON je.id=jl.journal_entry_id
		WHERE coa.owner_id=? AND coa.category IN ('revenue','expense') AND je.status='posted'
		  AND je.entry_date BETWEEN ? AND ? AND (?=0 OR je.property_id=?) GROUP BY coa.category`, ownerID, start, end, propertyID, propertyID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var category string
		var balance int64
		if err := rows.Scan(&category, &balance); err != nil {
			rows.Close()
			return page, err
		}
		if category == "revenue" {
			page.Summary.Revenue = balance
		} else {
			page.Summary.Expenses = balance
		}
	}
	rows.Close()
	page.Summary.Profit = page.Summary.Revenue - page.Summary.Expenses
	_ = a.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(CAST(jl.debit AS SIGNED)-CAST(jl.credit AS SIGNED)),0) FROM chart_of_accounts coa JOIN journal_lines jl ON jl.account_id=coa.id JOIN journal_entries je ON je.id=jl.journal_entry_id WHERE coa.owner_id=? AND coa.system_key='accounts_receivable' AND je.status='posted' AND je.entry_date<=? AND (?=0 OR je.property_id=?)`, ownerID, end, propertyID, propertyID).Scan(&page.Summary.Receivables)
	_ = a.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(CAST(jl.debit AS SIGNED)-CAST(jl.credit AS SIGNED)),0) FROM chart_of_accounts coa JOIN journal_lines jl ON jl.account_id=coa.id JOIN journal_entries je ON je.id=jl.journal_entry_id WHERE coa.owner_id=? AND je.status='posted' AND je.entry_date<=? AND (?=0 OR je.property_id=?) AND coa.id IN (SELECT ledger_account_id FROM bank_accounts WHERE owner_id=? UNION SELECT ledger_account_id FROM cash_accounts WHERE owner_id=?)`, ownerID, end, propertyID, propertyID, ownerID, ownerID).Scan(&page.Summary.Cash)
	_ = a.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN DATEDIFF(?,i.due_date)<=0 THEN CAST(i.total_amount AS SIGNED)-CAST(i.paid_amount AS SIGNED) ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN DATEDIFF(?,i.due_date) BETWEEN 1 AND 30 THEN CAST(i.total_amount AS SIGNED)-CAST(i.paid_amount AS SIGNED) ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN DATEDIFF(?,i.due_date) BETWEEN 31 AND 60 THEN CAST(i.total_amount AS SIGNED)-CAST(i.paid_amount AS SIGNED) ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN DATEDIFF(?,i.due_date) BETWEEN 61 AND 90 THEN CAST(i.total_amount AS SIGNED)-CAST(i.paid_amount AS SIGNED) ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN DATEDIFF(?,i.due_date)>90 THEN CAST(i.total_amount AS SIGNED)-CAST(i.paid_amount AS SIGNED) ELSE 0 END),0)
		FROM invoices i JOIN properties p ON p.id=i.property_id
		WHERE p.owner_id=? AND i.issue_date<=? AND i.status NOT IN ('void','refunded') AND i.total_amount>i.paid_amount AND (?=0 OR i.property_id=?)`, end, end, end, end, end, ownerID, end, propertyID, propertyID).Scan(&page.Aging.Current, &page.Aging.Days1To30, &page.Aging.Days31To60, &page.Aging.Days61To90, &page.Aging.Over90)
	rows, err = a.db.QueryContext(ctx, `
		SELECT je.source_type,
			MAX(CASE WHEN coa.system_key='fixed_assets' THEN 1 ELSE 0 END) AS has_fixed_asset,
			COALESCE(SUM(CASE WHEN jl.account_id IN (
				SELECT ledger_account_id FROM bank_accounts WHERE owner_id=?
				UNION SELECT ledger_account_id FROM cash_accounts WHERE owner_id=?
			) THEN CAST(jl.debit AS SIGNED)-CAST(jl.credit AS SIGNED) ELSE 0 END),0) AS cash_change
		FROM journal_entries je
		JOIN journal_lines jl ON jl.journal_entry_id=je.id
		JOIN chart_of_accounts coa ON coa.id=jl.account_id
		WHERE je.owner_id=? AND je.status='posted' AND je.entry_date BETWEEN ? AND ? AND (?=0 OR je.property_id=?)
		GROUP BY je.id,je.source_type`, ownerID, ownerID, ownerID, start, end, propertyID, propertyID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var source string
		var hasFixed bool
		var change int64
		if err := rows.Scan(&source, &hasFixed, &change); err != nil {
			rows.Close()
			return page, err
		}
		switch {
		case source == "internal_transfer":
			// Transfer antar akun keuangan tidak mengubah arus kas bersih usaha.
		case source == "opening_balance":
			page.CashFlow.Financing += change
		case source == "expense" && hasFixed:
			page.CashFlow.Investing += change
		default:
			page.CashFlow.Operating += change
		}
	}
	rows.Close()
	page.CashFlow.Net = page.CashFlow.Operating + page.CashFlow.Investing + page.CashFlow.Financing
	rows, err = a.db.QueryContext(ctx, `SELECT je.id,je.entry_number,je.description,CONCAT(je.source_type,'/',je.event_type),je.status,je.entry_date,SUM(jl.debit),SUM(jl.credit) FROM journal_entries je JOIN journal_lines jl ON jl.journal_entry_id=je.id WHERE je.owner_id=? AND je.entry_date BETWEEN ? AND ? AND (?=0 OR je.property_id=?) GROUP BY je.id ORDER BY je.entry_date DESC,je.id DESC LIMIT 200`, ownerID, start, end, propertyID, propertyID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v JournalView
		if err := rows.Scan(&v.ID, &v.Number, &v.Description, &v.Source, &v.Status, &v.Date, &v.Debit, &v.Credit); err != nil {
			rows.Close()
			return page, err
		}
		page.Journals = append(page.Journals, v)
	}
	rows.Close()
	return page, nil
}

func (a *App) ownerJournalCSV(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	propertyID, _ := strconv.ParseInt(r.URL.Query().Get("property_id"), 10, 64)
	start, e1 := parseBusinessDate(r.URL.Query().Get("start"))
	end, e2 := parseBusinessDate(r.URL.Query().Get("end"))
	if e1 != nil {
		start = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	if e2 != nil {
		end = businessToday(a.location)
	}
	if propertyID > 0 {
		var count int
		_ = a.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM properties WHERE id=? AND owner_id=?`, propertyID, session.User.ID).Scan(&count)
		if count != 1 {
			http.NotFound(w, r)
			return
		}
	}
	rows, err := a.db.QueryContext(r.Context(), `SELECT je.entry_date,je.entry_number,je.description,coa.code,coa.name,jl.debit,jl.credit FROM journal_entries je JOIN journal_lines jl ON jl.journal_entry_id=je.id JOIN chart_of_accounts coa ON coa.id=jl.account_id WHERE je.owner_id=? AND je.entry_date BETWEEN ? AND ? AND (?=0 OR je.property_id=?) ORDER BY je.entry_date,je.id,jl.id`, session.User.ID, start, end, propertyID, propertyID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer rows.Close()
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="jurnal-alam-segar.csv"`)
	writer := csv.NewWriter(w)
	defer writer.Flush()
	_ = writer.Write([]string{"Tanggal", "Nomor", "Keterangan", "Kode Akun", "Nama Akun", "Debit", "Kredit"})
	for rows.Next() {
		var date time.Time
		var number, description, code, name string
		var debit, credit int64
		if err := rows.Scan(&date, &number, &description, &code, &name, &debit, &credit); err != nil {
			return
		}
		_ = writer.Write([]string{date.Format("2006-01-02"), number, description, code, name, strconv.FormatInt(debit, 10), strconv.FormatInt(credit, 10)})
	}
}

func optionalID(value string) sql.NullInt64 {
	id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	return sql.NullInt64{Int64: id, Valid: err == nil && id > 0}
}

var _ = fmt.Sprint
