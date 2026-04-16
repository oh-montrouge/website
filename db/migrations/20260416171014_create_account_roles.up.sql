CREATE TABLE account_roles (
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    role_id    BIGINT NOT NULL REFERENCES roles(id),
    PRIMARY KEY (account_id, role_id)
);

COMMENT ON TABLE  account_roles            IS 'Many-to-many join: which accounts hold which roles.';
COMMENT ON COLUMN account_roles.account_id IS 'FK → accounts. Cascades on account deletion.';
COMMENT ON COLUMN account_roles.role_id    IS 'FK → roles.';
