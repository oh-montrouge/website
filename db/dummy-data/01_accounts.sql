-- Dummy accounts for local development only.
-- All accounts share the same password: devpassword
-- Hash generated with: mise run hash-password -- devpassword
--   $argon2id$v=19$m=65536,t=3,p=4$c1/VVeJxVfd7wHHFjoeYRg$p7F3VLaMo47nnjxCxP9uBv3tk+Vlw4C+QXwBsfQ+dqY

TRUNCATE accounts RESTART IDENTITY CASCADE;

INSERT INTO accounts (first_name, last_name, email, password_hash, main_instrument_id, status) VALUES
  ('Alice',     'Martin',  'alice@ohm-dev.local',     '$argon2id$v=19$m=65536,t=3,p=4$c1/VVeJxVfd7wHHFjoeYRg$p7F3VLaMo47nnjxCxP9uBv3tk+Vlw4C+QXwBsfQ+dqY', 15, 'active'),
  ('Bob',       'Dupont',  'bob@ohm-dev.local',       '$argon2id$v=19$m=65536,t=3,p=4$c1/VVeJxVfd7wHHFjoeYRg$p7F3VLaMo47nnjxCxP9uBv3tk+Vlw4C+QXwBsfQ+dqY', 10, 'active'),
  ('Claire',    'Leblanc', 'claire@ohm-dev.local',    '$argon2id$v=19$m=65536,t=3,p=4$c1/VVeJxVfd7wHHFjoeYRg$p7F3VLaMo47nnjxCxP9uBv3tk+Vlw4C+QXwBsfQ+dqY',  3, 'active'),
  ('David',     'Morin',   'david@ohm-dev.local',     '$argon2id$v=19$m=65536,t=3,p=4$c1/VVeJxVfd7wHHFjoeYRg$p7F3VLaMo47nnjxCxP9uBv3tk+Vlw4C+QXwBsfQ+dqY',  1, 'active'),
  ('Emma',      'Petit',   'emma@ohm-dev.local',      '$argon2id$v=19$m=65536,t=3,p=4$c1/VVeJxVfd7wHHFjoeYRg$p7F3VLaMo47nnjxCxP9uBv3tk+Vlw4C+QXwBsfQ+dqY',  2, 'active'),
  ('François',  'Garnier', 'francois@ohm-dev.local',  NULL,                                                                                                   14, 'pending'),
  ('Gabrielle', 'Simon',   'gabrielle@ohm-dev.local', '$argon2id$v=19$m=65536,t=3,p=4$c1/VVeJxVfd7wHHFjoeYRg$p7F3VLaMo47nnjxCxP9uBv3tk+Vlw4C+QXwBsfQ+dqY',  7, 'active');
