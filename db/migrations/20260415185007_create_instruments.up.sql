CREATE TABLE instruments (
    id   BIGSERIAL PRIMARY KEY,
    name TEXT      NOT NULL,
    CONSTRAINT instruments_name_unique UNIQUE (name)
);
