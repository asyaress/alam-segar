package httpapp

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

var ErrClosedPeriod = errors.New("periode akuntansi sudah ditutup")

type JournalLineInput struct {
	AccountID   int64
	Description string
	Debit       int64
	Credit      int64
}

type JournalInput struct {
	OwnerID     int64
	PropertyID  sql.NullInt64
	Date        time.Time
	Description string
	SourceType  string
	SourceID    int64
	EventType   string
	ActorID     int64
	Lines       []JournalLineInput
}

var defaultAccounts = []struct {
	Code, Name, Category, Normal, Key string
}{
	{"1100", "Piutang Usaha", "asset", "debit", "accounts_receivable"},
	{"1200", "Aset Tetap", "asset", "debit", "fixed_assets"},
	{"2100", "Kewajiban Uang Booking", "liability", "credit", "booking_liability"},
	{"2110", "Pendapatan Sewa Diterima di Muka", "liability", "credit", "unearned_rent"},
	{"2120", "Kredit Penghuni", "liability", "credit", "tenant_credit"},
	{"2130", "Utang Refund", "liability", "credit", "refund_payable"},
	{"3100", "Modal dan Saldo Awal", "equity", "credit", "opening_equity"},
	{"3200", "Prive Pemilik", "equity", "debit", "owner_draw"},
	{"4100", "Pendapatan Sewa", "revenue", "credit", "rent_revenue"},
	{"4200", "Pendapatan Booking", "revenue", "credit", "booking_revenue"},
	{"4300", "Pemulihan Kerusakan", "revenue", "credit", "damage_recovery"},
	{"5100", "Beban Listrik Bersama", "expense", "debit", "electricity_expense"},
	{"5110", "Beban Air Bersama", "expense", "debit", "water_expense"},
	{"5120", "Beban Internet", "expense", "debit", "internet_expense"},
	{"5130", "Beban Kebersihan", "expense", "debit", "cleaning_expense"},
	{"5140", "Beban Pemeliharaan", "expense", "debit", "maintenance_expense"},
	{"5150", "Beban Perlengkapan", "expense", "debit", "supplies_expense"},
	{"5160", "Beban Gaji", "expense", "debit", "salary_expense"},
	{"5190", "Beban Lain-lain", "expense", "debit", "other_expense"},
}

func ensureDefaultAccounts(ctx context.Context, tx *sql.Tx, ownerID int64) error {
	for _, account := range defaultAccounts {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO chart_of_accounts (owner_id, code, name, category, normal_balance, system_key)
			VALUES (?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE name=VALUES(name), category=VALUES(category), normal_balance=VALUES(normal_balance)`,
			ownerID, account.Code, account.Name, account.Category, account.Normal, account.Key)
		if err != nil {
			return err
		}
	}
	return nil
}

func systemAccountID(ctx context.Context, tx *sql.Tx, ownerID int64, key string) (int64, error) {
	if err := ensureDefaultAccounts(ctx, tx, ownerID); err != nil {
		return 0, err
	}
	var id int64
	err := tx.QueryRowContext(ctx, `SELECT id FROM chart_of_accounts WHERE owner_id=? AND system_key=? AND is_active=1`, ownerID, key).Scan(&id)
	return id, err
}

func postJournalTx(ctx context.Context, tx *sql.Tx, input JournalInput) (int64, error) {
	if input.OwnerID <= 0 || input.SourceID <= 0 || input.ActorID <= 0 || input.SourceType == "" || input.EventType == "" || len(input.Lines) < 2 {
		return 0, errors.New("data jurnal tidak lengkap")
	}
	var debit, credit int64
	for _, line := range input.Lines {
		if line.AccountID <= 0 || (line.Debit <= 0 && line.Credit <= 0) || (line.Debit > 0 && line.Credit > 0) {
			return 0, errors.New("baris jurnal tidak valid")
		}
		debit += line.Debit
		credit += line.Credit
	}
	if debit != credit || debit <= 0 {
		return 0, fmt.Errorf("jurnal tidak seimbang: debit=%d kredit=%d", debit, credit)
	}
	var closed int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM accounting_periods
		WHERE owner_id=? AND status='closed' AND ? BETWEEN period_start AND period_end`, input.OwnerID, input.Date.Format("2006-01-02")).Scan(&closed); err != nil {
		return 0, err
	}
	if closed > 0 {
		return 0, ErrClosedPeriod
	}
	var accountCount int
	ids := make([]any, 0, len(input.Lines)+1)
	ids = append(ids, input.OwnerID)
	placeholders := ""
	for index, line := range input.Lines {
		if index > 0 {
			placeholders += ","
		}
		placeholders += "?"
		ids = append(ids, line.AccountID)
	}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM chart_of_accounts WHERE owner_id=? AND id IN (`+placeholders+`)`, ids...).Scan(&accountCount); err != nil {
		return 0, err
	}
	if accountCount != len(input.Lines) {
		return 0, errors.New("akun jurnal bukan milik owner")
	}
	entryNumber, err := referenceNumber("JRN")
	if err != nil {
		return 0, err
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO journal_entries
			(owner_id, property_id, entry_number, entry_date, description, source_type, source_id, event_type, posted_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, input.OwnerID, nullableInt64(input.PropertyID), entryNumber,
		input.Date.Format("2006-01-02"), input.Description, input.SourceType, input.SourceID, input.EventType, input.ActorID)
	if err != nil {
		if isDuplicate(err) {
			var existingID int64
			lookupErr := tx.QueryRowContext(ctx, `
				SELECT id FROM journal_entries WHERE owner_id=? AND source_type=? AND source_id=? AND event_type=?`,
				input.OwnerID, input.SourceType, input.SourceID, input.EventType).Scan(&existingID)
			return existingID, lookupErr
		}
		return 0, err
	}
	entryID, _ := result.LastInsertId()
	for _, line := range input.Lines {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO journal_lines (journal_entry_id, account_id, description, debit, credit)
			VALUES (?, ?, ?, ?, ?)`, entryID, line.AccountID, nullableString(line.Description), line.Debit, line.Credit); err != nil {
			return 0, err
		}
	}
	return entryID, nil
}

func reverseJournalTx(ctx context.Context, tx *sql.Tx, ownerID, originalID, actorID int64, date time.Time, reason string) (int64, error) {
	var propertyID sql.NullInt64
	var originalNumber, status string
	if err := tx.QueryRowContext(ctx, `SELECT property_id,entry_number,status FROM journal_entries WHERE id=? AND owner_id=? FOR UPDATE`, originalID, ownerID).Scan(&propertyID, &originalNumber, &status); err != nil {
		return 0, err
	}
	if status != "posted" {
		return 0, errors.New("jurnal sudah dibalik")
	}
	rows, err := tx.QueryContext(ctx, `SELECT account_id,COALESCE(description,''),debit,credit FROM journal_lines WHERE journal_entry_id=? ORDER BY id`, originalID)
	if err != nil {
		return 0, err
	}
	var lines []JournalLineInput
	for rows.Next() {
		var line JournalLineInput
		if err := rows.Scan(&line.AccountID, &line.Description, &line.Credit, &line.Debit); err != nil {
			rows.Close()
			return 0, err
		}
		lines = append(lines, line)
	}
	rows.Close()
	entryID, err := postJournalTx(ctx, tx, JournalInput{OwnerID: ownerID, PropertyID: propertyID, Date: date, Description: "Pembalikan " + originalNumber + ": " + reason, SourceType: "journal_entry", SourceID: originalID, EventType: "reversal", ActorID: actorID, Lines: lines})
	if err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE journal_entries SET status='reversed',reversed_entry_id=? WHERE id=?`, entryID, originalID); err != nil {
		return 0, err
	}
	return entryID, nil
}

func financialAccountLedgerID(ctx context.Context, tx *sql.Tx, ownerID int64, accountType string, accountID int64, lock bool) (int64, error) {
	query := ""
	switch accountType {
	case "bank":
		query = `SELECT ledger_account_id FROM bank_accounts WHERE id=? AND owner_id=? AND status='active'`
	case "cash":
		query = `SELECT ledger_account_id FROM cash_accounts WHERE id=? AND owner_id=? AND status='active'`
	default:
		return 0, errors.New("jenis akun keuangan tidak valid")
	}
	if lock {
		query += " FOR UPDATE"
	}
	var ledgerID sql.NullInt64
	if err := tx.QueryRowContext(ctx, query, accountID, ownerID).Scan(&ledgerID); err != nil {
		return 0, err
	}
	if !ledgerID.Valid {
		return 0, errors.New("akun keuangan belum terhubung ke COA")
	}
	return ledgerID.Int64, nil
}

func ledgerBalance(ctx context.Context, tx *sql.Tx, accountID int64) (int64, error) {
	var category string
	var debits, credits int64
	err := tx.QueryRowContext(ctx, `
		SELECT coa.category, COALESCE(SUM(CASE WHEN je.status='posted' THEN jl.debit ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN je.status='posted' THEN jl.credit ELSE 0 END),0)
		FROM chart_of_accounts coa
		LEFT JOIN journal_lines jl ON jl.account_id=coa.id
		LEFT JOIN journal_entries je ON je.id=jl.journal_entry_id
		WHERE coa.id=? GROUP BY coa.id, coa.category`, accountID).Scan(&category, &debits, &credits)
	if err != nil {
		return 0, err
	}
	if category == "asset" || category == "expense" {
		return debits - credits, nil
	}
	return credits - debits, nil
}

func referenceNumber(prefix string) (string, error) {
	buffer := make([]byte, 5)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s-%x", prefix, time.Now().UTC().Format("20060102"), buffer), nil
}

func nullableInt64(value sql.NullInt64) any {
	if !value.Valid {
		return nil
	}
	return value.Int64
}

func encryptAccountNumber(secret, plain string) (string, error) {
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.RawURLEncoding.EncodeToString(sealed), nil
}

func decryptAccountNumber(secret, encoded string) (string, error) {
	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(payload) < gcm.NonceSize() {
		return "", errors.New("nomor rekening terenkripsi rusak")
	}
	plain, err := gcm.Open(nil, payload[:gcm.NonceSize()], payload[gcm.NonceSize():], nil)
	return string(plain), err
}

func parseID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("ID tidak valid")
	}
	return id, nil
}

func parseFinancialAccountRef(value string) (string, int64, error) {
	typ, rawID, ok := strings.Cut(strings.TrimSpace(value), ":")
	if !ok || (typ != "bank" && typ != "cash") {
		return "", 0, errors.New("referensi akun keuangan tidak valid")
	}
	id, err := parseID(rawID)
	if err != nil {
		return "", 0, err
	}
	return typ, id, nil
}
