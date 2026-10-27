ALTER TABLE revenue_recognition_schedules
    MODIFY COLUMN status ENUM('scheduled','posted','cancelled') NOT NULL DEFAULT 'scheduled';

CREATE TABLE payment_refunds (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    refund_number VARCHAR(50) NOT NULL,
    payment_id BIGINT UNSIGNED NOT NULL,
    invoice_id BIGINT UNSIGNED NOT NULL,
    bank_account_id BIGINT UNSIGNED NOT NULL,
    amount BIGINT UNSIGNED NOT NULL,
    refund_date DATE NOT NULL,
    reason TEXT NOT NULL,
    proof_storage_path VARCHAR(500) NOT NULL,
    status ENUM('posted','reversed') NOT NULL DEFAULT 'posted',
    created_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_payment_refund_number (refund_number),
    KEY idx_payment_refund_payment (payment_id, status),
    CONSTRAINT chk_payment_refund_amount CHECK (amount > 0),
    CONSTRAINT fk_payment_refund_payment FOREIGN KEY (payment_id) REFERENCES payments (id),
    CONSTRAINT fk_payment_refund_invoice FOREIGN KEY (invoice_id) REFERENCES invoices (id),
    CONSTRAINT fk_payment_refund_bank FOREIGN KEY (bank_account_id) REFERENCES bank_accounts (id),
    CONSTRAINT fk_payment_refund_actor FOREIGN KEY (created_by) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE reconciliations (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    owner_id BIGINT UNSIGNED NOT NULL,
    account_type ENUM('bank','cash') NOT NULL,
    account_id BIGINT UNSIGNED NOT NULL,
    reconciliation_date DATE NOT NULL,
    book_balance BIGINT NOT NULL,
    actual_balance BIGINT NOT NULL,
    difference_amount BIGINT NOT NULL,
    notes TEXT NOT NULL,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_reconciliation_account_date (owner_id, account_type, account_id, reconciliation_date),
    CONSTRAINT fk_reconciliation_owner FOREIGN KEY (owner_id) REFERENCES users (id),
    CONSTRAINT fk_reconciliation_actor FOREIGN KEY (created_by) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
