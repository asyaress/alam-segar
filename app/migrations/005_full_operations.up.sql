CREATE TABLE email_verification_tokens (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    token_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    expires_at DATETIME(6) NOT NULL,
    used_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_email_verification_token (token_hash),
    KEY idx_email_verification_user (user_id, expires_at),
    CONSTRAINT fk_email_verification_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE password_reset_tokens (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    token_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    expires_at DATETIME(6) NOT NULL,
    used_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_password_reset_token (token_hash),
    KEY idx_password_reset_user (user_id, expires_at),
    CONSTRAINT fk_password_reset_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

ALTER TABLE leases
    ADD COLUMN installment_count TINYINT UNSIGNED NOT NULL DEFAULT 1 AFTER partial_payment_allowed,
    ADD COLUMN special_terms TEXT NULL AFTER installment_count,
    ADD COLUMN contract_storage_path VARCHAR(500) NULL AFTER special_terms,
    ADD COLUMN actual_check_in_at DATETIME(6) NULL AFTER contract_storage_path,
    ADD COLUMN actual_check_out_at DATETIME(6) NULL AFTER actual_check_in_at,
    ADD COLUMN checked_in_by BIGINT UNSIGNED NULL AFTER actual_check_out_at,
    ADD COLUMN checked_out_by BIGINT UNSIGNED NULL AFTER checked_in_by,
    ADD CONSTRAINT chk_lease_installment_count CHECK (installment_count IN (1,2,4,6,12)),
    ADD CONSTRAINT fk_lease_checkin_actor FOREIGN KEY (checked_in_by) REFERENCES users (id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_lease_checkout_actor FOREIGN KEY (checked_out_by) REFERENCES users (id) ON DELETE SET NULL;

CREATE TABLE lease_versions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    lease_id BIGINT UNSIGNED NOT NULL,
    version_number INT UNSIGNED NOT NULL,
    effective_date DATE NOT NULL,
    change_type ENUM('created','extended','amended','terminated') NOT NULL,
    before_json LONGTEXT NULL,
    after_json LONGTEXT NOT NULL,
    reason TEXT NOT NULL,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_lease_version (lease_id, version_number),
    CONSTRAINT fk_lease_version_lease FOREIGN KEY (lease_id) REFERENCES leases (id),
    CONSTRAINT fk_lease_version_actor FOREIGN KEY (created_by) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE lease_installments (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    lease_id BIGINT UNSIGNED NOT NULL,
    installment_number TINYINT UNSIGNED NOT NULL,
    due_date DATE NOT NULL,
    amount BIGINT UNSIGNED NOT NULL,
    invoice_id BIGINT UNSIGNED NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_lease_installment_number (lease_id, installment_number),
    CONSTRAINT chk_lease_installment_amount CHECK (amount > 0),
    CONSTRAINT fk_lease_installment_lease FOREIGN KEY (lease_id) REFERENCES leases (id),
    CONSTRAINT fk_lease_installment_invoice FOREIGN KEY (invoice_id) REFERENCES invoices (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

ALTER TABLE invoices
    ADD COLUMN notes TEXT NULL AFTER status,
    ADD COLUMN issued_by BIGINT UNSIGNED NULL AFTER notes,
    ADD COLUMN voided_at DATETIME(6) NULL AFTER issued_by,
    ADD COLUMN voided_by BIGINT UNSIGNED NULL AFTER voided_at,
    ADD COLUMN void_reason TEXT NULL AFTER voided_by,
    ADD CONSTRAINT fk_invoice_issuer FOREIGN KEY (issued_by) REFERENCES users (id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_invoice_void_actor FOREIGN KEY (voided_by) REFERENCES users (id) ON DELETE SET NULL;

CREATE TABLE invoice_lines (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    invoice_id BIGINT UNSIGNED NOT NULL,
    description VARCHAR(255) NOT NULL,
    quantity DECIMAL(12,2) NOT NULL DEFAULT 1,
    unit_amount BIGINT UNSIGNED NOT NULL,
    line_total BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_invoice_lines_invoice (invoice_id),
    CONSTRAINT chk_invoice_line_positive CHECK (quantity > 0 AND unit_amount > 0 AND line_total > 0),
    CONSTRAINT fk_invoice_line_invoice FOREIGN KEY (invoice_id) REFERENCES invoices (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

ALTER TABLE payment_submissions
    ADD COLUMN idempotency_key CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NULL AFTER submission_number,
    ADD COLUMN proof_sha256 CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NULL AFTER proof_storage_path,
    ADD UNIQUE KEY uq_payment_submission_idempotency (idempotency_key),
    ADD KEY idx_payment_submission_proof_hash (tenant_id, proof_sha256);

CREATE TABLE payment_submission_allocations (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    payment_submission_id BIGINT UNSIGNED NOT NULL,
    invoice_id BIGINT UNSIGNED NOT NULL,
    requested_amount BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_submission_invoice (payment_submission_id, invoice_id),
    CONSTRAINT chk_submission_allocation_amount CHECK (requested_amount > 0),
    CONSTRAINT fk_submission_allocation_submission FOREIGN KEY (payment_submission_id) REFERENCES payment_submissions (id) ON DELETE CASCADE,
    CONSTRAINT fk_submission_allocation_invoice FOREIGN KEY (invoice_id) REFERENCES invoices (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE payments (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    payment_number VARCHAR(50) NOT NULL,
    submission_id BIGINT UNSIGNED NOT NULL,
    tenant_id BIGINT UNSIGNED NOT NULL,
    bank_account_id BIGINT UNSIGNED NOT NULL,
    amount BIGINT UNSIGNED NOT NULL,
    payment_date DATE NOT NULL,
    verified_by BIGINT UNSIGNED NOT NULL,
    verified_at DATETIME(6) NOT NULL,
    receipt_number VARCHAR(50) NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_payment_number (payment_number),
    UNIQUE KEY uq_payment_submission (submission_id),
    UNIQUE KEY uq_payment_receipt (receipt_number),
    CONSTRAINT fk_payment_submission FOREIGN KEY (submission_id) REFERENCES payment_submissions (id),
    CONSTRAINT fk_payment_tenant FOREIGN KEY (tenant_id) REFERENCES users (id),
    CONSTRAINT fk_payment_bank FOREIGN KEY (bank_account_id) REFERENCES bank_accounts (id),
    CONSTRAINT fk_payment_verifier FOREIGN KEY (verified_by) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE payment_allocations (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    payment_id BIGINT UNSIGNED NOT NULL,
    invoice_id BIGINT UNSIGNED NOT NULL,
    amount BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_payment_allocation_invoice (payment_id, invoice_id),
    CONSTRAINT chk_payment_allocation_amount CHECK (amount > 0),
    CONSTRAINT fk_payment_allocation_payment FOREIGN KEY (payment_id) REFERENCES payments (id),
    CONSTRAINT fk_payment_allocation_invoice FOREIGN KEY (invoice_id) REFERENCES invoices (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE tenant_credits (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    tenant_id BIGINT UNSIGNED NOT NULL,
    property_id BIGINT UNSIGNED NOT NULL,
    source_type VARCHAR(50) NOT NULL,
    source_id BIGINT UNSIGNED NOT NULL,
    amount BIGINT UNSIGNED NOT NULL,
    remaining_amount BIGINT UNSIGNED NOT NULL,
    status ENUM('available','applied','refunded') NOT NULL DEFAULT 'available',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_tenant_credit_source (source_type, source_id),
    CONSTRAINT chk_tenant_credit_amount CHECK (amount > 0 AND remaining_amount <= amount),
    CONSTRAINT fk_tenant_credit_tenant FOREIGN KEY (tenant_id) REFERENCES users (id),
    CONSTRAINT fk_tenant_credit_property FOREIGN KEY (property_id) REFERENCES properties (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE booking_payment_submissions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    submission_number VARCHAR(50) NOT NULL,
    booking_request_id BIGINT UNSIGNED NOT NULL,
    bank_account_id BIGINT UNSIGNED NOT NULL,
    amount BIGINT UNSIGNED NOT NULL,
    transfer_date DATE NOT NULL,
    bank_reference VARCHAR(190) NULL,
    proof_storage_path VARCHAR(500) NOT NULL,
    proof_sha256 CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    status ENUM('pending_verification','verified','rejected') NOT NULL DEFAULT 'pending_verification',
    reviewed_by BIGINT UNSIGNED NULL,
    reviewed_at DATETIME(6) NULL,
    rejection_reason TEXT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_booking_submission_number (submission_number),
    KEY idx_booking_submission_booking (booking_request_id, status),
    KEY idx_booking_submission_status (status, created_at),
    CONSTRAINT chk_booking_submission_amount CHECK (amount > 0),
    CONSTRAINT fk_booking_submission_booking FOREIGN KEY (booking_request_id) REFERENCES booking_requests (id),
    CONSTRAINT fk_booking_submission_bank FOREIGN KEY (bank_account_id) REFERENCES bank_accounts (id),
    CONSTRAINT fk_booking_submission_reviewer FOREIGN KEY (reviewed_by) REFERENCES users (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE booking_funds (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    booking_request_id BIGINT UNSIGNED NOT NULL,
    booking_payment_submission_id BIGINT UNSIGNED NOT NULL,
    received_amount BIGINT UNSIGNED NOT NULL,
    remaining_liability BIGINT UNSIGNED NOT NULL,
    status ENUM('held','allocated','fee_recognized','refund_pending','refunded') NOT NULL DEFAULT 'held',
    verified_at DATETIME(6) NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_booking_fund_booking (booking_request_id),
    UNIQUE KEY uq_booking_fund_submission (booking_payment_submission_id),
    CONSTRAINT chk_booking_fund_amount CHECK (received_amount > 0 AND remaining_liability <= received_amount),
    CONSTRAINT fk_booking_fund_booking FOREIGN KEY (booking_request_id) REFERENCES booking_requests (id),
    CONSTRAINT fk_booking_fund_submission FOREIGN KEY (booking_payment_submission_id) REFERENCES booking_payment_submissions (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE booking_refunds (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    refund_number VARCHAR(50) NOT NULL,
    booking_fund_id BIGINT UNSIGNED NOT NULL,
    bank_account_id BIGINT UNSIGNED NOT NULL,
    amount BIGINT UNSIGNED NOT NULL,
    refund_date DATE NOT NULL,
    reason TEXT NOT NULL,
    proof_storage_path VARCHAR(500) NOT NULL,
    status ENUM('posted','reversed') NOT NULL DEFAULT 'posted',
    created_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_booking_refund_number (refund_number),
    CONSTRAINT chk_booking_refund_amount CHECK (amount > 0),
    CONSTRAINT fk_booking_refund_fund FOREIGN KEY (booking_fund_id) REFERENCES booking_funds (id),
    CONSTRAINT fk_booking_refund_bank FOREIGN KEY (bank_account_id) REFERENCES bank_accounts (id),
    CONSTRAINT fk_booking_refund_actor FOREIGN KEY (created_by) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE chart_of_accounts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    owner_id BIGINT UNSIGNED NOT NULL,
    code VARCHAR(30) NOT NULL,
    name VARCHAR(150) NOT NULL,
    category ENUM('asset','liability','equity','revenue','expense') NOT NULL,
    normal_balance ENUM('debit','credit') NOT NULL,
    system_key VARCHAR(80) NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_coa_owner_code (owner_id, code),
    UNIQUE KEY uq_coa_owner_system (owner_id, system_key),
    CONSTRAINT fk_coa_owner FOREIGN KEY (owner_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

ALTER TABLE bank_accounts
    ADD COLUMN ledger_account_id BIGINT UNSIGNED NULL AFTER account_number_last4,
    ADD COLUMN opened_at DATE NULL AFTER ledger_account_id,
    ADD COLUMN opening_balance_posted TINYINT(1) NOT NULL DEFAULT 0 AFTER opened_at,
    ADD CONSTRAINT fk_bank_ledger_account FOREIGN KEY (ledger_account_id) REFERENCES chart_of_accounts (id) ON DELETE RESTRICT;

CREATE TABLE cash_accounts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    owner_id BIGINT UNSIGNED NOT NULL,
    property_id BIGINT UNSIGNED NULL,
    name VARCHAR(120) NOT NULL,
    custodian_name VARCHAR(150) NOT NULL,
    ledger_account_id BIGINT UNSIGNED NOT NULL,
    opened_at DATE NULL,
    opening_balance_posted TINYINT(1) NOT NULL DEFAULT 0,
    status ENUM('active','inactive') NOT NULL DEFAULT 'active',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_cash_owner_name (owner_id, name),
    CONSTRAINT fk_cash_owner FOREIGN KEY (owner_id) REFERENCES users (id),
    CONSTRAINT fk_cash_property FOREIGN KEY (property_id) REFERENCES properties (id) ON DELETE SET NULL,
    CONSTRAINT fk_cash_ledger_account FOREIGN KEY (ledger_account_id) REFERENCES chart_of_accounts (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE accounting_periods (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    owner_id BIGINT UNSIGNED NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    status ENUM('open','closed') NOT NULL DEFAULT 'open',
    closed_by BIGINT UNSIGNED NULL,
    closed_at DATETIME(6) NULL,
    close_notes TEXT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_accounting_period (owner_id, period_start, period_end),
    KEY idx_accounting_period_lookup (owner_id, status, period_start, period_end),
    CONSTRAINT chk_accounting_period_dates CHECK (period_end >= period_start),
    CONSTRAINT fk_accounting_period_owner FOREIGN KEY (owner_id) REFERENCES users (id),
    CONSTRAINT fk_accounting_period_closer FOREIGN KEY (closed_by) REFERENCES users (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE journal_entries (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    owner_id BIGINT UNSIGNED NOT NULL,
    property_id BIGINT UNSIGNED NULL,
    entry_number VARCHAR(50) NOT NULL,
    entry_date DATE NOT NULL,
    description VARCHAR(255) NOT NULL,
    source_type VARCHAR(80) NOT NULL,
    source_id BIGINT UNSIGNED NOT NULL,
    event_type VARCHAR(80) NOT NULL,
    status ENUM('posted','reversed') NOT NULL DEFAULT 'posted',
    reversed_entry_id BIGINT UNSIGNED NULL,
    posted_by BIGINT UNSIGNED NOT NULL,
    posted_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_journal_entry_number (entry_number),
    UNIQUE KEY uq_journal_source_event (owner_id, source_type, source_id, event_type),
    KEY idx_journal_owner_date (owner_id, entry_date),
    CONSTRAINT fk_journal_owner FOREIGN KEY (owner_id) REFERENCES users (id),
    CONSTRAINT fk_journal_property FOREIGN KEY (property_id) REFERENCES properties (id) ON DELETE SET NULL,
    CONSTRAINT fk_journal_reversed FOREIGN KEY (reversed_entry_id) REFERENCES journal_entries (id) ON DELETE SET NULL,
    CONSTRAINT fk_journal_actor FOREIGN KEY (posted_by) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE journal_lines (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    journal_entry_id BIGINT UNSIGNED NOT NULL,
    account_id BIGINT UNSIGNED NOT NULL,
    description VARCHAR(255) NULL,
    debit BIGINT UNSIGNED NOT NULL DEFAULT 0,
    credit BIGINT UNSIGNED NOT NULL DEFAULT 0,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_journal_lines_entry (journal_entry_id),
    KEY idx_journal_lines_account (account_id, journal_entry_id),
    CONSTRAINT chk_journal_line_one_side CHECK ((debit > 0 AND credit = 0) OR (credit > 0 AND debit = 0)),
    CONSTRAINT fk_journal_line_entry FOREIGN KEY (journal_entry_id) REFERENCES journal_entries (id),
    CONSTRAINT fk_journal_line_account FOREIGN KEY (account_id) REFERENCES chart_of_accounts (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE revenue_recognition_schedules (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    lease_id BIGINT UNSIGNED NOT NULL,
    property_id BIGINT UNSIGNED NOT NULL,
    recognition_date DATE NOT NULL,
    service_start_date DATE NOT NULL,
    service_end_date DATE NOT NULL,
    amount BIGINT UNSIGNED NOT NULL,
    status ENUM('scheduled','posted') NOT NULL DEFAULT 'scheduled',
    journal_entry_id BIGINT UNSIGNED NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_revenue_schedule_period (lease_id, service_start_date, service_end_date),
    KEY idx_revenue_schedule_due (status, recognition_date),
    CONSTRAINT fk_revenue_schedule_lease FOREIGN KEY (lease_id) REFERENCES leases (id),
    CONSTRAINT fk_revenue_schedule_property FOREIGN KEY (property_id) REFERENCES properties (id),
    CONSTRAINT fk_revenue_schedule_journal FOREIGN KEY (journal_entry_id) REFERENCES journal_entries (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE expenses (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    expense_number VARCHAR(50) NOT NULL,
    owner_id BIGINT UNSIGNED NOT NULL,
    property_id BIGINT UNSIGNED NULL,
    expense_account_id BIGINT UNSIGNED NOT NULL,
    payment_account_type ENUM('bank','cash') NOT NULL,
    payment_account_id BIGINT UNSIGNED NOT NULL,
    expense_date DATE NOT NULL,
    amount BIGINT UNSIGNED NOT NULL,
    vendor_name VARCHAR(150) NULL,
    description TEXT NOT NULL,
    proof_storage_path VARCHAR(500) NOT NULL,
    proof_sha256 CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    status ENUM('posted','reversed') NOT NULL DEFAULT 'posted',
    created_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_expense_number (expense_number),
    KEY idx_expense_owner_date (owner_id, expense_date),
    CONSTRAINT chk_expense_amount CHECK (amount > 0),
    CONSTRAINT fk_expense_owner FOREIGN KEY (owner_id) REFERENCES users (id),
    CONSTRAINT fk_expense_property FOREIGN KEY (property_id) REFERENCES properties (id) ON DELETE SET NULL,
    CONSTRAINT fk_expense_account FOREIGN KEY (expense_account_id) REFERENCES chart_of_accounts (id),
    CONSTRAINT fk_expense_actor FOREIGN KEY (created_by) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE internal_transfers (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    transfer_number VARCHAR(50) NOT NULL,
    owner_id BIGINT UNSIGNED NOT NULL,
    transfer_date DATE NOT NULL,
    from_account_type ENUM('bank','cash') NOT NULL,
    from_account_id BIGINT UNSIGNED NOT NULL,
    to_account_type ENUM('bank','cash') NOT NULL,
    to_account_id BIGINT UNSIGNED NOT NULL,
    amount BIGINT UNSIGNED NOT NULL,
    notes TEXT NULL,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_internal_transfer_number (transfer_number),
    CONSTRAINT chk_internal_transfer_amount CHECK (amount > 0),
    CONSTRAINT fk_internal_transfer_owner FOREIGN KEY (owner_id) REFERENCES users (id),
    CONSTRAINT fk_internal_transfer_actor FOREIGN KEY (created_by) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE opening_balances (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    owner_id BIGINT UNSIGNED NOT NULL,
    account_type ENUM('bank','cash') NOT NULL,
    account_id BIGINT UNSIGNED NOT NULL,
    balance_date DATE NOT NULL,
    amount BIGINT UNSIGNED NOT NULL,
    journal_entry_id BIGINT UNSIGNED NOT NULL,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_opening_balance_account (owner_id, account_type, account_id),
    CONSTRAINT fk_opening_balance_owner FOREIGN KEY (owner_id) REFERENCES users (id),
    CONSTRAINT fk_opening_balance_journal FOREIGN KEY (journal_entry_id) REFERENCES journal_entries (id),
    CONSTRAINT fk_opening_balance_actor FOREIGN KEY (created_by) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE assets (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    owner_id BIGINT UNSIGNED NOT NULL,
    property_id BIGINT UNSIGNED NOT NULL,
    room_id BIGINT UNSIGNED NULL,
    asset_code VARCHAR(50) NOT NULL,
    name VARCHAR(150) NOT NULL,
    category VARCHAR(100) NOT NULL,
    serial_number VARCHAR(120) NULL,
    acquired_date DATE NULL,
    acquired_cost BIGINT UNSIGNED NULL,
    condition_status ENUM('good','fair','damaged','lost','disposed') NOT NULL DEFAULT 'good',
    is_fixed_asset TINYINT(1) NOT NULL DEFAULT 0,
    notes TEXT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_asset_owner_code (owner_id, asset_code),
    KEY idx_asset_location (property_id, room_id, condition_status),
    CONSTRAINT fk_asset_owner FOREIGN KEY (owner_id) REFERENCES users (id),
    CONSTRAINT fk_asset_property FOREIGN KEY (property_id) REFERENCES properties (id),
    CONSTRAINT fk_asset_room FOREIGN KEY (room_id) REFERENCES rooms (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE inspections (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    lease_id BIGINT UNSIGNED NOT NULL,
    inspection_type ENUM('check_in','check_out') NOT NULL,
    inspected_at DATETIME(6) NOT NULL,
    overall_notes TEXT NULL,
    photo_storage_path VARCHAR(500) NOT NULL,
    recorded_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_inspection_lease_type (lease_id, inspection_type),
    CONSTRAINT fk_inspection_lease FOREIGN KEY (lease_id) REFERENCES leases (id),
    CONSTRAINT fk_inspection_actor FOREIGN KEY (recorded_by) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE inspection_items (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    inspection_id BIGINT UNSIGNED NOT NULL,
    asset_id BIGINT UNSIGNED NULL,
    item_name_snapshot VARCHAR(150) NOT NULL,
    quantity INT UNSIGNED NOT NULL DEFAULT 1,
    condition_snapshot VARCHAR(100) NOT NULL,
    notes TEXT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_inspection_item_inspection FOREIGN KEY (inspection_id) REFERENCES inspections (id) ON DELETE CASCADE,
    CONSTRAINT fk_inspection_item_asset FOREIGN KEY (asset_id) REFERENCES assets (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE damage_charges (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    lease_id BIGINT UNSIGNED NOT NULL,
    inspection_id BIGINT UNSIGNED NULL,
    invoice_id BIGINT UNSIGNED NOT NULL,
    amount BIGINT UNSIGNED NOT NULL,
    basis ENUM('actual_cost','agreement') NOT NULL,
    reason TEXT NOT NULL,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_damage_charge_invoice (invoice_id),
    CONSTRAINT fk_damage_charge_lease FOREIGN KEY (lease_id) REFERENCES leases (id),
    CONSTRAINT fk_damage_charge_inspection FOREIGN KEY (inspection_id) REFERENCES inspections (id) ON DELETE SET NULL,
    CONSTRAINT fk_damage_charge_invoice FOREIGN KEY (invoice_id) REFERENCES invoices (id),
    CONSTRAINT fk_damage_charge_actor FOREIGN KEY (created_by) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE maintenance_tickets (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    ticket_number VARCHAR(50) NOT NULL,
    tenant_id BIGINT UNSIGNED NOT NULL,
    lease_id BIGINT UNSIGNED NOT NULL,
    property_id BIGINT UNSIGNED NOT NULL,
    room_id BIGINT UNSIGNED NOT NULL,
    category ENUM('electricity','water','ac','furniture','security','cleanliness','other') NOT NULL,
    urgency ENUM('low','normal','high','emergency') NOT NULL DEFAULT 'normal',
    subject VARCHAR(180) NOT NULL,
    description TEXT NOT NULL,
    photo_storage_path VARCHAR(500) NULL,
    status ENUM('open','in_progress','waiting_for_tenant','resolved','closed') NOT NULL DEFAULT 'open',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    closed_at DATETIME(6) NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_maintenance_ticket_number (ticket_number),
    KEY idx_maintenance_owner_status (property_id, status, created_at),
    CONSTRAINT fk_maintenance_tenant FOREIGN KEY (tenant_id) REFERENCES users (id),
    CONSTRAINT fk_maintenance_lease FOREIGN KEY (lease_id) REFERENCES leases (id),
    CONSTRAINT fk_maintenance_property FOREIGN KEY (property_id) REFERENCES properties (id),
    CONSTRAINT fk_maintenance_room FOREIGN KEY (room_id) REFERENCES rooms (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE maintenance_updates (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    ticket_id BIGINT UNSIGNED NOT NULL,
    actor_user_id BIGINT UNSIGNED NOT NULL,
    status_from VARCHAR(40) NULL,
    status_to VARCHAR(40) NOT NULL,
    message TEXT NOT NULL,
    vendor_name VARCHAR(150) NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_maintenance_updates_ticket (ticket_id, created_at),
    CONSTRAINT fk_maintenance_update_ticket FOREIGN KEY (ticket_id) REFERENCES maintenance_tickets (id) ON DELETE CASCADE,
    CONSTRAINT fk_maintenance_update_actor FOREIGN KEY (actor_user_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE announcements (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    owner_id BIGINT UNSIGNED NOT NULL,
    property_id BIGINT UNSIGNED NULL,
    title VARCHAR(180) NOT NULL,
    body TEXT NOT NULL,
    published_at DATETIME(6) NOT NULL,
    expires_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_announcements_property_date (property_id, published_at),
    CONSTRAINT fk_announcement_owner FOREIGN KEY (owner_id) REFERENCES users (id),
    CONSTRAINT fk_announcement_property FOREIGN KEY (property_id) REFERENCES properties (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE notifications (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    event_key VARCHAR(120) NOT NULL,
    title VARCHAR(180) NOT NULL,
    body TEXT NOT NULL,
    action_url VARCHAR(500) NULL,
    read_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_notifications_user_read (user_id, read_at, created_at),
    CONSTRAINT fk_notification_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE notification_deliveries (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    channel ENUM('email') NOT NULL,
    event_key VARCHAR(120) NOT NULL,
    recipient VARCHAR(190) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    status ENUM('queued','sent','failed') NOT NULL DEFAULT 'queued',
    attempts INT UNSIGNED NOT NULL DEFAULT 0,
    last_error TEXT NULL,
    sent_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_notification_delivery_event (user_id, channel, event_key),
    KEY idx_notification_delivery_queue (status, created_at),
    CONSTRAINT fk_notification_delivery_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT INTO chart_of_accounts (owner_id, code, name, category, normal_balance, system_key)
SELECT id, '1100', 'Piutang Usaha', 'asset', 'debit', 'accounts_receivable' FROM users WHERE role='owner' AND deleted_at IS NULL;
INSERT INTO chart_of_accounts (owner_id, code, name, category, normal_balance, system_key)
SELECT id, '2100', 'Kewajiban Uang Booking', 'liability', 'credit', 'booking_liability' FROM users WHERE role='owner' AND deleted_at IS NULL;
INSERT INTO chart_of_accounts (owner_id, code, name, category, normal_balance, system_key)
SELECT id, '2110', 'Pendapatan Sewa Diterima di Muka', 'liability', 'credit', 'unearned_rent' FROM users WHERE role='owner' AND deleted_at IS NULL;
INSERT INTO chart_of_accounts (owner_id, code, name, category, normal_balance, system_key)
SELECT id, '2120', 'Kredit Penghuni', 'liability', 'credit', 'tenant_credit' FROM users WHERE role='owner' AND deleted_at IS NULL;
INSERT INTO chart_of_accounts (owner_id, code, name, category, normal_balance, system_key)
SELECT id, '2130', 'Utang Refund', 'liability', 'credit', 'refund_payable' FROM users WHERE role='owner' AND deleted_at IS NULL;
INSERT INTO chart_of_accounts (owner_id, code, name, category, normal_balance, system_key)
SELECT id, '3100', 'Modal dan Saldo Awal', 'equity', 'credit', 'opening_equity' FROM users WHERE role='owner' AND deleted_at IS NULL;
INSERT INTO chart_of_accounts (owner_id, code, name, category, normal_balance, system_key)
SELECT id, '3200', 'Prive Pemilik', 'equity', 'debit', 'owner_draw' FROM users WHERE role='owner' AND deleted_at IS NULL;
INSERT INTO chart_of_accounts (owner_id, code, name, category, normal_balance, system_key)
SELECT id, '4100', 'Pendapatan Sewa', 'revenue', 'credit', 'rent_revenue' FROM users WHERE role='owner' AND deleted_at IS NULL;
INSERT INTO chart_of_accounts (owner_id, code, name, category, normal_balance, system_key)
SELECT id, '4200', 'Pendapatan Booking', 'revenue', 'credit', 'booking_revenue' FROM users WHERE role='owner' AND deleted_at IS NULL;
INSERT INTO chart_of_accounts (owner_id, code, name, category, normal_balance, system_key)
SELECT id, '4300', 'Pemulihan Kerusakan', 'revenue', 'credit', 'damage_recovery' FROM users WHERE role='owner' AND deleted_at IS NULL;
INSERT INTO chart_of_accounts (owner_id, code, name, category, normal_balance, system_key)
SELECT id, '5100', 'Beban Listrik Bersama', 'expense', 'debit', 'electricity_expense' FROM users WHERE role='owner' AND deleted_at IS NULL;
INSERT INTO chart_of_accounts (owner_id, code, name, category, normal_balance, system_key)
SELECT id, '5110', 'Beban Air Bersama', 'expense', 'debit', 'water_expense' FROM users WHERE role='owner' AND deleted_at IS NULL;
INSERT INTO chart_of_accounts (owner_id, code, name, category, normal_balance, system_key)
SELECT id, '5120', 'Beban Internet', 'expense', 'debit', 'internet_expense' FROM users WHERE role='owner' AND deleted_at IS NULL;
INSERT INTO chart_of_accounts (owner_id, code, name, category, normal_balance, system_key)
SELECT id, '5130', 'Beban Kebersihan', 'expense', 'debit', 'cleaning_expense' FROM users WHERE role='owner' AND deleted_at IS NULL;
INSERT INTO chart_of_accounts (owner_id, code, name, category, normal_balance, system_key)
SELECT id, '5140', 'Beban Pemeliharaan', 'expense', 'debit', 'maintenance_expense' FROM users WHERE role='owner' AND deleted_at IS NULL;
INSERT INTO chart_of_accounts (owner_id, code, name, category, normal_balance, system_key)
SELECT id, '5150', 'Beban Perlengkapan', 'expense', 'debit', 'supplies_expense' FROM users WHERE role='owner' AND deleted_at IS NULL;
INSERT INTO chart_of_accounts (owner_id, code, name, category, normal_balance, system_key)
SELECT id, '5160', 'Beban Gaji', 'expense', 'debit', 'salary_expense' FROM users WHERE role='owner' AND deleted_at IS NULL;
INSERT INTO chart_of_accounts (owner_id, code, name, category, normal_balance, system_key)
SELECT id, '5190', 'Beban Lain-lain', 'expense', 'debit', 'other_expense' FROM users WHERE role='owner' AND deleted_at IS NULL;
