CREATE TABLE tenant_credit_applications (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    application_number VARCHAR(50) NOT NULL,
    owner_id BIGINT UNSIGNED NOT NULL,
    tenant_id BIGINT UNSIGNED NOT NULL,
    property_id BIGINT UNSIGNED NOT NULL,
    invoice_id BIGINT UNSIGNED NOT NULL,
    amount BIGINT UNSIGNED NOT NULL,
    application_date DATE NOT NULL,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_tenant_credit_application_number (application_number),
    KEY idx_tenant_credit_application_invoice (invoice_id, application_date),
    CONSTRAINT chk_tenant_credit_application_amount CHECK (amount > 0),
    CONSTRAINT fk_tenant_credit_application_owner FOREIGN KEY (owner_id) REFERENCES users (id),
    CONSTRAINT fk_tenant_credit_application_tenant FOREIGN KEY (tenant_id) REFERENCES users (id),
    CONSTRAINT fk_tenant_credit_application_property FOREIGN KEY (property_id) REFERENCES properties (id),
    CONSTRAINT fk_tenant_credit_application_invoice FOREIGN KEY (invoice_id) REFERENCES invoices (id),
    CONSTRAINT fk_tenant_credit_application_actor FOREIGN KEY (created_by) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE tenant_credit_application_lines (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    application_id BIGINT UNSIGNED NOT NULL,
    tenant_credit_id BIGINT UNSIGNED NOT NULL,
    amount BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_tenant_credit_application_line (application_id, tenant_credit_id),
    CONSTRAINT chk_tenant_credit_application_line_amount CHECK (amount > 0),
    CONSTRAINT fk_tenant_credit_application_line_header FOREIGN KEY (application_id) REFERENCES tenant_credit_applications (id) ON DELETE CASCADE,
    CONSTRAINT fk_tenant_credit_application_line_credit FOREIGN KEY (tenant_credit_id) REFERENCES tenant_credits (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
