-- Versi aplikasi sebelum migrasi ini mengirim time.Time WITA ke koneksi MySQL
-- berlokasi UTC. Nilai yang masuk ke kolom DATE bergeser satu hari ke belakang.
-- Perbaikan ini hanya menyentuh kolom tanggal yang berasal dari input/plan Go;
-- tanggal jurnal dikirim sebagai string YYYY-MM-DD dan tidak ikut bergeser.
UPDATE booking_requests
SET planned_start_date=DATE_ADD(planned_start_date, INTERVAL 1 DAY),
    planned_end_date=DATE_ADD(planned_end_date, INTERVAL 1 DAY);

UPDATE booking_payment_submissions SET transfer_date=DATE_ADD(transfer_date, INTERVAL 1 DAY);
UPDATE booking_refunds SET refund_date=DATE_ADD(refund_date, INTERVAL 1 DAY);

UPDATE leases
SET start_date=DATE_ADD(start_date, INTERVAL 1 DAY),
    end_date=DATE_ADD(end_date, INTERVAL 1 DAY);
UPDATE lease_versions SET effective_date=DATE_ADD(effective_date, INTERVAL 1 DAY);
UPDATE lease_installments SET due_date=DATE_ADD(due_date, INTERVAL 1 DAY);
UPDATE lease_terminations SET effective_date=DATE_ADD(effective_date, INTERVAL 1 DAY);

UPDATE invoices
SET service_start_date=DATE_ADD(service_start_date, INTERVAL 1 DAY),
    service_end_date=DATE_ADD(service_end_date, INTERVAL 1 DAY),
    issue_date=DATE_ADD(issue_date, INTERVAL 1 DAY),
    due_date=DATE_ADD(due_date, INTERVAL 1 DAY)
WHERE invoice_type='rent';
UPDATE invoices
SET due_date=DATE_ADD(due_date, INTERVAL 1 DAY)
WHERE invoice_type='damage';

UPDATE revenue_recognition_schedules
SET recognition_date=DATE_ADD(recognition_date, INTERVAL 1 DAY),
    service_start_date=DATE_ADD(service_start_date, INTERVAL 1 DAY),
    service_end_date=DATE_ADD(service_end_date, INTERVAL 1 DAY);

UPDATE payment_submissions SET transfer_date=DATE_ADD(transfer_date, INTERVAL 1 DAY);
UPDATE payments SET payment_date=DATE_ADD(payment_date, INTERVAL 1 DAY);
UPDATE payment_refunds SET refund_date=DATE_ADD(refund_date, INTERVAL 1 DAY);
UPDATE tenant_credit_refunds SET refund_date=DATE_ADD(refund_date, INTERVAL 1 DAY);

UPDATE opening_balances SET balance_date=DATE_ADD(balance_date, INTERVAL 1 DAY);
UPDATE internal_transfers SET transfer_date=DATE_ADD(transfer_date, INTERVAL 1 DAY);
UPDATE expenses SET expense_date=DATE_ADD(expense_date, INTERVAL 1 DAY);
UPDATE reconciliations SET reconciliation_date=DATE_ADD(reconciliation_date, INTERVAL 1 DAY);
UPDATE accounting_periods
SET period_start=DATE_ADD(period_start, INTERVAL 1 DAY),
    period_end=DATE_ADD(period_end, INTERVAL 1 DAY);
