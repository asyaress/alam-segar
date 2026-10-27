-- Calon penghuni belum menjadi user internal, tetapi tetap harus menerima
-- email transaksional booking. Foreign key tetap melindungi user_id saat ada.
ALTER TABLE notification_deliveries
    MODIFY COLUMN user_id BIGINT UNSIGNED NULL;
