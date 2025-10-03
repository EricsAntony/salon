ALTER TABLE staff
    ADD COLUMN phone_number TEXT,
    ADD COLUMN email TEXT;

CREATE UNIQUE INDEX idx_staff_phone_number ON staff (phone_number) WHERE phone_number IS NOT NULL;

CREATE TABLE staff_otps (
    id BIGSERIAL PRIMARY KEY,
    phone_number TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    attempts INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_staff_otps_phone ON staff_otps (phone_number);

CREATE TABLE staff_refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    staff_id UUID NOT NULL REFERENCES staff(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_staff_refresh_tokens_staff ON staff_refresh_tokens (staff_id);
CREATE UNIQUE INDEX idx_staff_refresh_tokens_hash ON staff_refresh_tokens (token_hash);
