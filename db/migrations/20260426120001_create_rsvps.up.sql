CREATE TABLE rsvps (
    id            BIGSERIAL PRIMARY KEY,
    account_id    BIGINT NOT NULL REFERENCES accounts(id),
    event_id      BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    state         TEXT NOT NULL DEFAULT 'unanswered',
    instrument_id BIGINT REFERENCES instruments(id),
    UNIQUE(account_id, event_id)
);

CREATE INDEX ON rsvps (account_id);
CREATE INDEX ON rsvps (event_id);

COMMENT ON TABLE  rsvps               IS 'RSVP record per (account, event). Created proactively; no implicit unanswered logic.';
COMMENT ON COLUMN rsvps.state         IS 'One of: unanswered, yes, no, maybe';
COMMENT ON COLUMN rsvps.instrument_id IS 'Non-null only when state=yes and event_type=concert';
