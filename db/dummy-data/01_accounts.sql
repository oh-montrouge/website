-- Dummy accounts for local development only.
-- All accounts share the same password: devpassword
-- Hash generated with: mise run hash-password -- devpassword
--   $argon2id$v=19$m=65536,t=3,p=4$c1/VVeJxVfd7wHHFjoeYRg$p7F3VLaMo47nnjxCxP9uBv3tk+Vlw4C+QXwBsfQ+dqY

TRUNCATE accounts RESTART IDENTITY CASCADE;

INSERT INTO accounts (first_name, last_name, email, password_hash, main_instrument_id, birth_date, phone, address, phone_address_consent, processing_restricted, status) VALUES
  -- Alice: admin, active, with birth date and phone/address consent
  ('Alice',     'Martin',    'alice@ohm-dev.local',     '$argon2id$v=19$m=65536,t=3,p=4$c1/VVeJxVfd7wHHFjoeYRg$p7F3VLaMo47nnjxCxP9uBv3tk+Vlw4C+QXwBsfQ+dqY', 15, '1985-06-15', '06 10 20 30 40', '12 rue des Lilas, 92120 Montrouge',    true,  false, 'active'),
  -- Bob: admin, active, with birth date, no phone/address consent
  ('Bob',       'Dupont',    'bob@ohm-dev.local',       '$argon2id$v=19$m=65536,t=3,p=4$c1/VVeJxVfd7wHHFjoeYRg$p7F3VLaMo47nnjxCxP9uBv3tk+Vlw4C+QXwBsfQ+dqY', 10, '1990-03-22', NULL,            NULL,                                   false, false, 'active'),
  -- Claire: active, no birth date, processing restricted
  ('Claire',    'Leblanc',   'claire@ohm-dev.local',    '$argon2id$v=19$m=65536,t=3,p=4$c1/VVeJxVfd7wHHFjoeYRg$p7F3VLaMo47nnjxCxP9uBv3tk+Vlw4C+QXwBsfQ+dqY',  3, NULL,         NULL,            NULL,                                   false, true,  'active'),
  -- David: active, with birth date and phone/address consent
  ('David',     'Morin',     'david@ohm-dev.local',     '$argon2id$v=19$m=65536,t=3,p=4$c1/VVeJxVfd7wHHFjoeYRg$p7F3VLaMo47nnjxCxP9uBv3tk+Vlw4C+QXwBsfQ+dqY',  1, '1978-11-05', '07 55 44 33 22', '3 avenue de la Gare, 92120 Montrouge',  true,  false, 'active'),
  -- Emma: active
  ('Emma',      'Petit',     'emma@ohm-dev.local',      '$argon2id$v=19$m=65536,t=3,p=4$c1/VVeJxVfd7wHHFjoeYRg$p7F3VLaMo47nnjxCxP9uBv3tk+Vlw4C+QXwBsfQ+dqY',  2, '2000-08-30', NULL,            NULL,                                   false, false, 'active'),
  -- François: pending (invited but not yet activated)
  ('François',  'Garnier',   'francois@ohm-dev.local',  NULL,                                                                                                   14, NULL,         NULL,            NULL,                                   false, false, 'pending'),
  -- Gabrielle: active, with birth date and phone/address consent
  ('Gabrielle', 'Simon',     'gabrielle@ohm-dev.local', '$argon2id$v=19$m=65536,t=3,p=4$c1/VVeJxVfd7wHHFjoeYRg$p7F3VLaMo47nnjxCxP9uBv3tk+Vlw4C+QXwBsfQ+dqY',  7, '1995-02-14', '06 77 88 99 00', '7 place du Théâtre, 92120 Montrouge',  true,  false, 'active'),
  -- Hugo: young musician (under 15), pending activation, parental consent required
  ('Hugo',      'Bernard',   'hugo@ohm-dev.local',      NULL,                                                                                                    5, '2014-04-10', NULL,            NULL,                                   false, false, 'pending'),
  -- Isabelle: anonymized (GDPR erasure — demonstrates anonymized state in admin UI)
  (NULL,        NULL,        NULL,                       NULL,                                                                                                    8, NULL,         NULL,            NULL,                                   false, false, 'anonymized');

UPDATE accounts SET parental_consent_uri = 'https://example.com/consentements/hugo-bernard.pdf' WHERE email = 'hugo@ohm-dev.local';

UPDATE accounts SET anonymization_token = 'anon_dev_isabelle_8f3a2c1d' WHERE status = 'anonymized';
