CREATE TABLE http_sessions (
    id          BIGSERIAL   PRIMARY KEY,
    key         TEXT        NOT NULL,
    data        BYTEA       NOT NULL,
    created_on  TIMESTAMPTZ NOT NULL,
    modified_on TIMESTAMPTZ NOT NULL,
    expires_on  TIMESTAMPTZ NOT NULL,
    account_id  BIGINT      REFERENCES accounts(id) ON DELETE CASCADE,
    CONSTRAINT http_sessions_key_unique UNIQUE (key)
);

CREATE INDEX http_sessions_account_id_idx ON http_sessions (account_id);

COMMENT ON TABLE  http_sessions            IS 'Server-side session store (pgstore). Each row is one active browser session.';
COMMENT ON COLUMN http_sessions.id         IS 'Surrogate primary key.';
COMMENT ON COLUMN http_sessions.key        IS 'Opaque session cookie value — unique per session.';
COMMENT ON COLUMN http_sessions.data       IS 'Serialised session payload (pgstore binary format).';
COMMENT ON COLUMN http_sessions.created_on IS 'Timestamp when the session was created.';
COMMENT ON COLUMN http_sessions.modified_on IS 'Timestamp of the last session write.';
COMMENT ON COLUMN http_sessions.expires_on IS 'Timestamp after which the session is expired and may be purged.';
COMMENT ON COLUMN http_sessions.account_id IS 'FK → accounts. Null until the user logs in. Cascades on account deletion.';
