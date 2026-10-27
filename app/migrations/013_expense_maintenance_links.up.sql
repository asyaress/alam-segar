ALTER TABLE expenses
    ADD COLUMN maintenance_ticket_id BIGINT UNSIGNED NULL AFTER property_id,
    ADD COLUMN asset_id BIGINT UNSIGNED NULL AFTER maintenance_ticket_id,
    ADD KEY idx_expense_ticket (maintenance_ticket_id),
    ADD KEY idx_expense_asset (asset_id),
    ADD CONSTRAINT fk_expense_ticket FOREIGN KEY (maintenance_ticket_id) REFERENCES maintenance_tickets (id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_expense_asset FOREIGN KEY (asset_id) REFERENCES assets (id) ON DELETE SET NULL;
