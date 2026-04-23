CREATE TABLE password_reset_tokens (
    id         BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES accounts(id),
    token      TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used       BOOLEAN NOT NULL DEFAULT FALSE
);

COMMENT ON TABLE  password_reset_tokens            IS 'Single-use tokens for the admin-initiated password reset flow. At most one active (unused, non-expired) token per account.';
COMMENT ON COLUMN password_reset_tokens.token      IS 'CSPRNG, 32 bytes, base64url-encoded.';
COMMENT ON COLUMN password_reset_tokens.expires_at IS 'UTC. 7 days from generation.';
COMMENT ON COLUMN password_reset_tokens.used       IS 'Set true when the password reset completes or when a new token is generated for the same account.';
