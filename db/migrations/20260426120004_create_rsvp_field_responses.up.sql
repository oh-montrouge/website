CREATE TABLE rsvp_field_responses (
    id             BIGSERIAL PRIMARY KEY,
    rsvp_id        BIGINT NOT NULL REFERENCES rsvps(id) ON DELETE CASCADE,
    event_field_id BIGINT NOT NULL REFERENCES event_fields(id) ON DELETE CASCADE,
    value          TEXT NOT NULL,
    UNIQUE(rsvp_id, event_field_id)
);

COMMENT ON TABLE  rsvp_field_responses       IS 'A musician''s response to a single event field, captured as part of a yes RSVP.';
COMMENT ON COLUMN rsvp_field_responses.value IS 'Stored as text. For choice fields, the value is the event_field_choice ID.';
