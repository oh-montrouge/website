-- Technically, seed_instruments don't exactly belong here
-- migrations are for schema (DDL), and mixing in DML seed data is a grey area. But it's the right call here
-- Instruments (and later: roles) are closer to an enum table — the app can't function without them,
--  they're identical across dev/test/prod, and there's no runtime UI to manage them. Putting them in
--  migrations guarantees any environment is in a consistent, runnable state after a single mise migrate.
INSERT INTO instruments (id, name) VALUES
  (1,  'Flûte'),
  (2,  'Hautbois'),
  (3,  'Clarinette'),
  (4,  'Clarinette Basse'),
  (5,  'Basson'),
  (6,  'Saxophone Soprano'),
  (7,  'Saxophone Alto'),
  (8,  'Saxophone Ténor'),
  (9,  'Saxophone Baryton'),
  (10, 'Trompette'),
  (11, 'Cor'),
  (12, 'Tuba'),
  (13, 'Trombone'),
  (14, 'Percussions'),
  (15, 'Chef d''orchestre');
