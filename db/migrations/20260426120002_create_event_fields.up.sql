CREATE TABLE event_fields (
    id         BIGSERIAL PRIMARY KEY,
    event_id   BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    label      TEXT NOT NULL,
    field_type TEXT NOT NULL,
    required   BOOLEAN NOT NULL DEFAULT false,
    position   INTEGER NOT NULL
);

COMMENT ON TABLE  event_fields            IS 'Custom fields defined by admin on other-type events.';
COMMENT ON COLUMN event_fields.field_type IS 'One of: choice, integer, text';
COMMENT ON COLUMN event_fields.position   IS 'Display order, ascending';
