CREATE TABLE fee_payments (
    id           BIGSERIAL PRIMARY KEY,
    account_id   BIGINT NOT NULL REFERENCES accounts(id),
    season_id    BIGINT NOT NULL REFERENCES seasons(id),
    amount       NUMERIC(10, 2) NOT NULL,
    payment_date DATE NOT NULL,
    payment_type TEXT NOT NULL,
    comment      TEXT,
    UNIQUE(account_id, season_id)
);

COMMENT ON TABLE  fee_payments              IS 'Records of fee payments per musician per season. At most one payment per (account, season) pair.';
COMMENT ON COLUMN fee_payments.payment_type IS 'One of: chèque, espèces, virement bancaire';
COMMENT ON COLUMN fee_payments.comment      IS 'Optional admin note';
