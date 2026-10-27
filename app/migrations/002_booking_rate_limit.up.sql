CREATE TABLE booking_request_attempts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    email_normalized VARCHAR(190) NOT NULL,
    ip_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    was_accepted TINYINT(1) NOT NULL DEFAULT 0,
    attempted_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_booking_attempts_lookup (email_normalized, ip_hash, attempted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

