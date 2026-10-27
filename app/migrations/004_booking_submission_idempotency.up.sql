ALTER TABLE booking_requests
    ADD COLUMN public_submission_key CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NULL AFTER reference_number,
    ADD UNIQUE KEY uq_booking_public_submission (public_submission_key);

