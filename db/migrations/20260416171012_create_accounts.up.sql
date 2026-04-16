CREATE TABLE accounts (
    id                    BIGSERIAL PRIMARY KEY,
    first_name            TEXT,
    last_name             TEXT,
    email                 TEXT,
    password_hash         TEXT,
    main_instrument_id    BIGINT NOT NULL REFERENCES instruments(id),
    birth_date            DATE,
    parental_consent_uri  TEXT,
    phone                 TEXT,
    address               TEXT,
    phone_address_consent BOOLEAN NOT NULL DEFAULT FALSE,
    status                TEXT    NOT NULL CHECK (status IN ('pending', 'active', 'anonymized')),
    processing_restricted BOOLEAN NOT NULL DEFAULT FALSE,
    anonymization_token   TEXT
);

CREATE UNIQUE INDEX accounts_email_unique_active
    ON accounts (email)
    WHERE status != 'anonymized';

COMMENT ON TABLE  accounts                         IS 'OHM member accounts. Three lifecycle states: pending (invited, not yet activated), active, and anonymized (GDPR erasure).';
COMMENT ON COLUMN accounts.id                      IS 'Surrogate primary key.';
COMMENT ON COLUMN accounts.first_name              IS 'Null when anonymized.';
COMMENT ON COLUMN accounts.last_name               IS 'Null when anonymized.';
COMMENT ON COLUMN accounts.email                   IS 'Null when anonymized. Unique among non-anonymized accounts via partial index.';
COMMENT ON COLUMN accounts.password_hash           IS 'Argon2id hash. Null when anonymized or not yet activated.';
COMMENT ON COLUMN accounts.main_instrument_id      IS 'Primary instrument. Retained on anonymization.';
COMMENT ON COLUMN accounts.birth_date              IS 'Optional. Null when anonymized or not provided.';
COMMENT ON COLUMN accounts.parental_consent_uri    IS 'URI to parental consent document for under-15 members. Null when anonymized or not applicable.';
COMMENT ON COLUMN accounts.phone                   IS 'Null when anonymized, consent withdrawn, or not provided.';
COMMENT ON COLUMN accounts.address                 IS 'Null when anonymized, consent withdrawn, or not provided.';
COMMENT ON COLUMN accounts.phone_address_consent   IS 'True when the member has consented to storing phone and address. Default false.';
COMMENT ON COLUMN accounts.status                  IS 'Account lifecycle state: pending | active | anonymized.';
COMMENT ON COLUMN accounts.processing_restricted   IS 'Admin flag indicating data processing should be restricted (e.g. pending legal hold). Default false.';
COMMENT ON COLUMN accounts.anonymization_token     IS 'Opaque token set at anonymization time, used to reference anonymized accounts in fee payment records. Null otherwise.';
