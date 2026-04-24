CREATE TABLE seasons (
    id         BIGSERIAL PRIMARY KEY,
    label      TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date   DATE NOT NULL,
    is_current BOOLEAN NOT NULL DEFAULT FALSE
);

-- At most one season may be current at the DB level.
-- Application invariant: exactly one is current once any season has been designated.
CREATE UNIQUE INDEX seasons_one_current ON seasons (is_current) WHERE is_current = true;

COMMENT ON TABLE  seasons            IS 'Annual membership periods. Immutable after creation. Exactly one may be designated current.';
COMMENT ON COLUMN seasons.is_current IS 'At most one row has is_current=true (partial unique index). Designation transfers are executed in a single transaction.';
