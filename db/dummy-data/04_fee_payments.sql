-- Dummy fee payments for local development only.
-- Account IDs match 01_accounts.sql insertion order:
--   1=Alice, 2=Bob, 3=Claire, 4=David, 5=Emma, 6=François(pending),
--   7=Gabrielle, 8=Hugo(pending), 9=anonymized
-- Season IDs match 03_seasons.sql:
--   1=2025-2026 (current), 2=2024-2025
-- Emma (5) intentionally has no payments — retention review case.

TRUNCATE fee_payments RESTART IDENTITY CASCADE;

INSERT INTO fee_payments (account_id, season_id, amount, payment_date, payment_type, comment) VALUES
  -- Alice: both seasons (earliest payment: 2024-09-15)
  (1, 2, 50.00, '2024-09-15', 'chèque',            NULL),
  (1, 1, 50.00, '2025-09-10', 'virement bancaire',  NULL),
  -- Bob: current season only
  (2, 1, 50.00, '2025-10-01', 'espèces',             NULL),
  -- Claire: current season with comment
  (3, 1, 50.00, '2025-09-20', 'chèque',             'Paiement anticipé'),
  -- David: both seasons
  (4, 2, 45.00, '2024-10-05', 'virement bancaire',  NULL),
  (4, 1, 50.00, '2025-09-25', 'chèque',              NULL),
  -- Gabrielle: previous season only
  (7, 2, 45.00, '2024-09-30', 'espèces',             NULL),
  -- Anonymized account (9): payment retained after anonymization
  (9, 2, 45.00, '2024-11-01', 'chèque',              NULL);
