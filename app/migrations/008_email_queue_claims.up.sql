ALTER TABLE notification_deliveries
    MODIFY COLUMN status ENUM('queued','processing','sent','failed') NOT NULL DEFAULT 'queued',
    ADD COLUMN claimed_at DATETIME(6) NULL AFTER attempts;
