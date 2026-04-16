CREATE TABLE roles (
    id   BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    CONSTRAINT roles_name_unique UNIQUE (name)
);

INSERT INTO roles (name) VALUES ('admin');

COMMENT ON TABLE  roles      IS 'Application roles (e.g. admin). Membership tracked in account_roles.';
COMMENT ON COLUMN roles.id   IS 'Surrogate primary key.';
COMMENT ON COLUMN roles.name IS 'Role name — unique slug (e.g. ''admin'').';
