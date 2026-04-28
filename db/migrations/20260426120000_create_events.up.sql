CREATE TABLE events (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    datetime    TIMESTAMPTZ NOT NULL,
    event_type  TEXT NOT NULL,
    description TEXT
);

COMMENT ON TABLE  events            IS 'Concerts, répétitions et autres rendez-vous de l''orchestre.';
COMMENT ON COLUMN events.event_type IS 'One of: concert, rehearsal, other';
COMMENT ON COLUMN events.datetime   IS 'Stored as UTC. Display in Europe/Paris timezone.';
