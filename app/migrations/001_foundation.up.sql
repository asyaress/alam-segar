CREATE TABLE users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    role ENUM('owner', 'tenant') NOT NULL,
    status ENUM('pending', 'active', 'inactive', 'rejected') NOT NULL DEFAULT 'pending',
    name VARCHAR(150) NOT NULL,
    email VARCHAR(190) NOT NULL,
    phone VARCHAR(40) NULL,
    password_hash VARCHAR(255) NOT NULL,
    email_verified_at DATETIME(6) NULL,
    approved_at DATETIME(6) NULL,
    approved_by BIGINT UNSIGNED NULL,
    last_login_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    deleted_at DATETIME(6) NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_users_email (email),
    KEY idx_users_role_status (role, status),
    CONSTRAINT fk_users_approved_by FOREIGN KEY (approved_by) REFERENCES users (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE user_sessions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    token_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    csrf_token CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    user_id BIGINT UNSIGNED NULL,
    expires_at DATETIME(6) NOT NULL,
    last_seen_at DATETIME(6) NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_user_sessions_token (token_hash),
    KEY idx_user_sessions_user (user_id),
    KEY idx_user_sessions_expires (expires_at),
    CONSTRAINT fk_user_sessions_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE login_attempts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    email_normalized VARCHAR(190) NOT NULL,
    ip_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    was_successful TINYINT(1) NOT NULL DEFAULT 0,
    attempted_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_login_attempts_lookup (email_normalized, ip_hash, attempted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE tenant_profiles (
    user_id BIGINT UNSIGNED NOT NULL,
    emergency_contact_name VARCHAR(150) NULL,
    emergency_contact_phone VARCHAR(40) NULL,
    notes TEXT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (user_id),
    CONSTRAINT fk_tenant_profiles_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE registration_requests (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    status ENUM('pending_email', 'pending_review', 'approved', 'rejected', 'cancelled') NOT NULL DEFAULT 'pending_email',
    reviewed_by BIGINT UNSIGNED NULL,
    reviewed_at DATETIME(6) NULL,
    review_reason TEXT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_registration_user_open (user_id, status),
    KEY idx_registration_status_created (status, created_at),
    CONSTRAINT fk_registration_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_registration_reviewer FOREIGN KEY (reviewed_by) REFERENCES users (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE properties (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    owner_id BIGINT UNSIGNED NOT NULL,
    code VARCHAR(40) NOT NULL,
    slug VARCHAR(120) NOT NULL,
    name VARCHAR(150) NOT NULL,
    address TEXT NOT NULL,
    description TEXT NULL,
    rules_text TEXT NULL,
    contact_phone VARCHAR(40) NULL,
    cover_image_path VARCHAR(500) NULL,
    is_published TINYINT(1) NOT NULL DEFAULT 0,
    status ENUM('active', 'inactive') NOT NULL DEFAULT 'active',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    deleted_at DATETIME(6) NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_properties_owner_code (owner_id, code),
    UNIQUE KEY uq_properties_slug (slug),
    KEY idx_properties_owner_status (owner_id, status),
    CONSTRAINT fk_properties_owner FOREIGN KEY (owner_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE room_types (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    property_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(120) NOT NULL,
    description TEXT NULL,
    base_monthly_rate BIGINT UNSIGNED NOT NULL,
    base_annual_rate BIGINT UNSIGNED NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_room_types_property_name (property_id, name),
    CONSTRAINT chk_room_types_monthly_rate CHECK (base_monthly_rate > 0),
    CONSTRAINT chk_room_types_annual_rate CHECK (base_annual_rate IS NULL OR base_annual_rate > 0),
    CONSTRAINT fk_room_types_property FOREIGN KEY (property_id) REFERENCES properties (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE rooms (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    property_id BIGINT UNSIGNED NOT NULL,
    room_type_id BIGINT UNSIGNED NULL,
    number VARCHAR(40) NOT NULL,
    slug VARCHAR(140) NOT NULL,
    block_name VARCHAR(80) NULL,
    floor_name VARCHAR(80) NULL,
    area_m2 DECIMAL(8,2) NULL,
    monthly_rate BIGINT UNSIGNED NOT NULL,
    annual_rate BIGINT UNSIGNED NULL,
    capacity TINYINT UNSIGNED NOT NULL DEFAULT 1,
    electricity_customer_number VARCHAR(100) NULL,
    water_customer_number VARCHAR(100) NULL,
    condition_status ENUM('ready', 'maintenance', 'inactive') NOT NULL DEFAULT 'ready',
    is_published TINYINT(1) NOT NULL DEFAULT 0,
    description TEXT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    deleted_at DATETIME(6) NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_rooms_property_number (property_id, number),
    UNIQUE KEY uq_rooms_property_slug (property_id, slug),
    KEY idx_rooms_property_condition (property_id, condition_status, is_published),
    CONSTRAINT chk_rooms_monthly_rate CHECK (monthly_rate > 0),
    CONSTRAINT chk_rooms_annual_rate CHECK (annual_rate IS NULL OR annual_rate > 0),
    CONSTRAINT chk_rooms_capacity_one CHECK (capacity = 1),
    CONSTRAINT fk_rooms_property FOREIGN KEY (property_id) REFERENCES properties (id),
    CONSTRAINT fk_rooms_type FOREIGN KEY (room_type_id) REFERENCES room_types (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE room_price_histories (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    room_id BIGINT UNSIGNED NOT NULL,
    monthly_rate BIGINT UNSIGNED NOT NULL,
    annual_rate BIGINT UNSIGNED NULL,
    effective_from DATE NOT NULL,
    reason TEXT NOT NULL,
    changed_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_room_price_effective (room_id, effective_from),
    CONSTRAINT fk_room_price_room FOREIGN KEY (room_id) REFERENCES rooms (id),
    CONSTRAINT fk_room_price_actor FOREIGN KEY (changed_by) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE room_photos (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    room_id BIGINT UNSIGNED NOT NULL,
    storage_path VARCHAR(500) NOT NULL,
    alt_text VARCHAR(255) NOT NULL,
    sort_order INT UNSIGNED NOT NULL DEFAULT 0,
    is_public TINYINT(1) NOT NULL DEFAULT 0,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_room_photos_room_sort (room_id, sort_order),
    CONSTRAINT fk_room_photos_room FOREIGN KEY (room_id) REFERENCES rooms (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE amenities (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    property_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(120) NOT NULL,
    icon_class VARCHAR(120) NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_amenities_property_name (property_id, name),
    CONSTRAINT fk_amenities_property FOREIGN KEY (property_id) REFERENCES properties (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE room_amenities (
    room_id BIGINT UNSIGNED NOT NULL,
    amenity_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (room_id, amenity_id),
    CONSTRAINT fk_room_amenities_room FOREIGN KEY (room_id) REFERENCES rooms (id) ON DELETE CASCADE,
    CONSTRAINT fk_room_amenities_amenity FOREIGN KEY (amenity_id) REFERENCES amenities (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE bank_accounts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    owner_id BIGINT UNSIGNED NOT NULL,
    property_id BIGINT UNSIGNED NULL,
    bank_name VARCHAR(120) NOT NULL,
    account_name VARCHAR(150) NOT NULL,
    account_number_encrypted TEXT NOT NULL,
    account_number_last4 CHAR(4) NOT NULL,
    is_accepting_payments TINYINT(1) NOT NULL DEFAULT 1,
    status ENUM('active', 'inactive') NOT NULL DEFAULT 'active',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_bank_accounts_owner_status (owner_id, status),
    CONSTRAINT fk_bank_accounts_owner FOREIGN KEY (owner_id) REFERENCES users (id),
    CONSTRAINT fk_bank_accounts_property FOREIGN KEY (property_id) REFERENCES properties (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE booking_requests (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    reference_number VARCHAR(50) NOT NULL,
    property_id BIGINT UNSIGNED NOT NULL,
    room_id BIGINT UNSIGNED NOT NULL,
    applicant_user_id BIGINT UNSIGNED NULL,
    applicant_name VARCHAR(150) NOT NULL,
    applicant_email VARCHAR(190) NOT NULL,
    applicant_phone VARCHAR(40) NOT NULL,
    planned_start_date DATE NOT NULL,
    planned_end_date DATE NOT NULL,
    booking_amount BIGINT UNSIGNED NULL,
    booking_usage ENUM('undecided', 'credit_first_rent', 'additional_booking_fee') NOT NULL DEFAULT 'undecided',
    status ENUM('submitted', 'under_review', 'approved_waiting_payment', 'proof_submitted', 'confirmed', 'converted_to_contract', 'rejected', 'cancelled', 'released', 'refund_pending', 'refunded') NOT NULL DEFAULT 'submitted',
    hold_started_at DATETIME(6) NULL,
    due_by DATETIME(6) NULL,
    owner_reason TEXT NULL,
    reviewed_by BIGINT UNSIGNED NULL,
    reviewed_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_booking_reference (reference_number),
    KEY idx_booking_room_dates (room_id, planned_start_date, planned_end_date, status),
    KEY idx_booking_property_status (property_id, status, created_at),
    CONSTRAINT chk_booking_dates CHECK (planned_end_date > planned_start_date),
    CONSTRAINT chk_booking_amount CHECK (booking_amount IS NULL OR booking_amount > 0),
    CONSTRAINT fk_booking_property FOREIGN KEY (property_id) REFERENCES properties (id),
    CONSTRAINT fk_booking_room FOREIGN KEY (room_id) REFERENCES rooms (id),
    CONSTRAINT fk_booking_applicant FOREIGN KEY (applicant_user_id) REFERENCES users (id) ON DELETE SET NULL,
    CONSTRAINT fk_booking_reviewer FOREIGN KEY (reviewed_by) REFERENCES users (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE leases (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    lease_number VARCHAR(50) NOT NULL,
    property_id BIGINT UNSIGNED NOT NULL,
    room_id BIGINT UNSIGNED NOT NULL,
    tenant_id BIGINT UNSIGNED NOT NULL,
    booking_request_id BIGINT UNSIGNED NULL,
    billing_cycle ENUM('monthly', 'annual') NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    original_anchor_day TINYINT UNSIGNED NOT NULL,
    due_day TINYINT UNSIGNED NOT NULL,
    agreed_rent BIGINT UNSIGNED NOT NULL,
    partial_payment_allowed TINYINT(1) NOT NULL DEFAULT 0,
    status ENUM('draft', 'active', 'expiring', 'ended', 'terminated') NOT NULL DEFAULT 'draft',
    created_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_leases_number (lease_number),
    UNIQUE KEY uq_leases_booking (booking_request_id),
    KEY idx_leases_room_dates (room_id, start_date, end_date, status),
    KEY idx_leases_tenant_dates (tenant_id, start_date, end_date, status),
    CONSTRAINT chk_lease_dates CHECK (end_date > start_date),
    CONSTRAINT chk_lease_anchor CHECK (original_anchor_day BETWEEN 1 AND 31),
    CONSTRAINT chk_lease_due_day CHECK (due_day BETWEEN 1 AND 31),
    CONSTRAINT chk_lease_rent CHECK (agreed_rent > 0),
    CONSTRAINT fk_leases_property FOREIGN KEY (property_id) REFERENCES properties (id),
    CONSTRAINT fk_leases_room FOREIGN KEY (room_id) REFERENCES rooms (id),
    CONSTRAINT fk_leases_tenant FOREIGN KEY (tenant_id) REFERENCES users (id),
    CONSTRAINT fk_leases_booking FOREIGN KEY (booking_request_id) REFERENCES booking_requests (id) ON DELETE SET NULL,
    CONSTRAINT fk_leases_creator FOREIGN KEY (created_by) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE invoices (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    invoice_number VARCHAR(50) NOT NULL,
    property_id BIGINT UNSIGNED NOT NULL,
    room_id BIGINT UNSIGNED NULL,
    tenant_id BIGINT UNSIGNED NOT NULL,
    lease_id BIGINT UNSIGNED NULL,
    invoice_type ENUM('rent', 'booking', 'damage', 'adjustment') NOT NULL,
    service_start_date DATE NULL,
    service_end_date DATE NULL,
    issue_date DATE NOT NULL,
    due_date DATE NOT NULL,
    total_amount BIGINT UNSIGNED NOT NULL,
    paid_amount BIGINT UNSIGNED NOT NULL DEFAULT 0,
    status ENUM('draft', 'issued', 'partially_paid', 'paid', 'overdue', 'void', 'refunded') NOT NULL DEFAULT 'draft',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_invoices_number (invoice_number),
    KEY idx_invoices_tenant_status (tenant_id, status, due_date),
    KEY idx_invoices_property_status (property_id, status, due_date),
    CONSTRAINT chk_invoice_total CHECK (total_amount > 0),
    CONSTRAINT chk_invoice_paid CHECK (paid_amount <= total_amount),
    CONSTRAINT fk_invoices_property FOREIGN KEY (property_id) REFERENCES properties (id),
    CONSTRAINT fk_invoices_room FOREIGN KEY (room_id) REFERENCES rooms (id) ON DELETE SET NULL,
    CONSTRAINT fk_invoices_tenant FOREIGN KEY (tenant_id) REFERENCES users (id),
    CONSTRAINT fk_invoices_lease FOREIGN KEY (lease_id) REFERENCES leases (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE payment_submissions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    submission_number VARCHAR(50) NOT NULL,
    tenant_id BIGINT UNSIGNED NOT NULL,
    bank_account_id BIGINT UNSIGNED NOT NULL,
    amount BIGINT UNSIGNED NOT NULL,
    transfer_date DATE NOT NULL,
    bank_reference VARCHAR(190) NULL,
    proof_storage_path VARCHAR(500) NOT NULL,
    status ENUM('pending_verification', 'verified', 'rejected') NOT NULL DEFAULT 'pending_verification',
    reviewed_by BIGINT UNSIGNED NULL,
    reviewed_at DATETIME(6) NULL,
    rejection_reason TEXT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_payment_submission_number (submission_number),
    KEY idx_payment_submission_status (status, created_at),
    KEY idx_payment_submission_tenant (tenant_id, created_at),
    CONSTRAINT chk_payment_submission_amount CHECK (amount > 0),
    CONSTRAINT fk_payment_submission_tenant FOREIGN KEY (tenant_id) REFERENCES users (id),
    CONSTRAINT fk_payment_submission_bank FOREIGN KEY (bank_account_id) REFERENCES bank_accounts (id),
    CONSTRAINT fk_payment_submission_reviewer FOREIGN KEY (reviewed_by) REFERENCES users (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE audit_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    actor_user_id BIGINT UNSIGNED NULL,
    action VARCHAR(120) NOT NULL,
    entity_type VARCHAR(120) NOT NULL,
    entity_id VARCHAR(120) NULL,
    before_json LONGTEXT NULL,
    after_json LONGTEXT NULL,
    ip_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_audit_entity (entity_type, entity_id, created_at),
    KEY idx_audit_actor (actor_user_id, created_at),
    CONSTRAINT fk_audit_actor FOREIGN KEY (actor_user_id) REFERENCES users (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE app_settings (
    setting_key VARCHAR(190) NOT NULL,
    setting_value LONGTEXT NOT NULL,
    updated_by BIGINT UNSIGNED NULL,
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (setting_key),
    CONSTRAINT fk_app_settings_actor FOREIGN KEY (updated_by) REFERENCES users (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

