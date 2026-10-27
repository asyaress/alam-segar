CREATE TABLE request_rate_limits (
    action VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    ip_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    window_started_at DATETIME(6) NOT NULL,
    attempts INT UNSIGNED NOT NULL DEFAULT 0,
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (action, ip_hash),
    KEY idx_request_rate_limits_cleanup (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
