CREATE TABLE booking_fund_allocations (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    booking_fund_id BIGINT UNSIGNED NOT NULL,
    lease_id BIGINT UNSIGNED NOT NULL,
    invoice_id BIGINT UNSIGNED NULL,
    amount BIGINT UNSIGNED NOT NULL,
    allocation_type ENUM('invoice_credit','tenant_credit','booking_fee') NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_booking_fund_allocation_type (booking_fund_id, allocation_type),
    CONSTRAINT chk_booking_fund_allocation_amount CHECK (amount > 0),
    CONSTRAINT fk_booking_fund_allocation_fund FOREIGN KEY (booking_fund_id) REFERENCES booking_funds (id),
    CONSTRAINT fk_booking_fund_allocation_lease FOREIGN KEY (lease_id) REFERENCES leases (id),
    CONSTRAINT fk_booking_fund_allocation_invoice FOREIGN KEY (invoice_id) REFERENCES invoices (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
