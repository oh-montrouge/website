CREATE TABLE event_field_choices (
    id             BIGSERIAL PRIMARY KEY,
    event_field_id BIGINT NOT NULL REFERENCES event_fields(id) ON DELETE CASCADE,
    label          TEXT NOT NULL,
    position       INTEGER NOT NULL
);

COMMENT ON TABLE  event_field_choices          IS 'Selectable options for choice-type event fields.';
COMMENT ON COLUMN event_field_choices.position IS 'Display order, ascending';
