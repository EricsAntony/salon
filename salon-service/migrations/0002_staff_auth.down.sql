DROP INDEX IF EXISTS idx_staff_refresh_tokens_hash;
DROP INDEX IF EXISTS idx_staff_refresh_tokens_staff;
DROP TABLE IF EXISTS staff_refresh_tokens;

DROP INDEX IF EXISTS idx_staff_otps_phone;
DROP TABLE IF EXISTS staff_otps;

DROP INDEX IF EXISTS idx_staff_phone_number;
ALTER TABLE staff
    DROP COLUMN IF EXISTS email,
    DROP COLUMN IF EXISTS phone_number;
