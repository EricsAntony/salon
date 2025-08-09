-- +migrate Down
DROP TABLE IF EXISTS otps;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS gender;
