ALTER TABLE room_price_histories
    DROP INDEX uq_room_price_effective,
    ADD INDEX idx_room_price_effective (room_id, effective_from, created_at);

