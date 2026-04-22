CREATE TABLE invite_tokens (
    id         BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES accounts(id),
    token      TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used       BOOLEAN NOT NULL DEFAULT FALSE
);

COMMENT ON TABLE  invite_tokens            IS 'Single-use tokens for the musician invite flow. At most one active (unused, non-expired) token per account.';
COMMENT ON COLUMN invite_tokens.token      IS 'CSPRNG, 32 bytes, base64url-encoded.';
COMMENT ON COLUMN invite_tokens.expires_at IS 'UTC. 7 days from generation.';
COMMENT ON COLUMN invite_tokens.used       IS 'Set true when the invite flow completes or when a new token is generated for the same account.';
