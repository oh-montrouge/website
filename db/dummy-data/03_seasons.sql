-- Dummy seasons for local development only.

TRUNCATE seasons RESTART IDENTITY CASCADE;

INSERT INTO seasons (label, start_date, end_date, is_current) VALUES
  ('2025-2026', '2025-09-01', '2026-08-31', true),
  ('2024-2025', '2024-09-01', '2025-08-31', false);
