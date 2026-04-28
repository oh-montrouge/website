-- Dummy events and RSVPs for local development only.
-- Active accounts: 1=Alice, 2=Bob, 3=Claire, 4=David, 5=Emma, 7=Gabrielle

TRUNCATE events RESTART IDENTITY CASCADE;

-- Past events
INSERT INTO events (name, event_type, datetime) VALUES
  ('Répétition de septembre',         'rehearsal', NOW() - INTERVAL '60 days'),
  ('Concert d''automne',              'concert',   NOW() - INTERVAL '30 days'),
  ('Assemblée générale',              'other',     NOW() - INTERVAL '15 days');

-- Upcoming events
INSERT INTO events (name, event_type, datetime) VALUES
  ('Répétition de mai',               'rehearsal', NOW() + INTERVAL '7 days'),
  ('Concert de printemps',            'concert',   NOW() + INTERVAL '45 days'),
  ('Réunion de bureau',               'other',     NOW() + INTERVAL '60 days');

-- Custom fields on the upcoming "other" event (id=6 after RESTART IDENTITY)
INSERT INTO event_fields (event_id, label, field_type, required, position) VALUES
  (6, 'Serez-vous disponible pour la mise en place ?', 'choice', true, 1),
  (6, 'Commentaire', 'text', false, 2);

INSERT INTO event_field_choices (event_field_id, label, position) VALUES
  (1, 'Oui, dès 18h', 1),
  (1, 'Oui, dès 19h', 2),
  (1, 'Non', 3);

-- RSVPs for past rehearsal (event_id=1): varied states
INSERT INTO rsvps (account_id, event_id, state) VALUES
  (1, 1, 'yes'),
  (2, 1, 'yes'),
  (3, 1, 'yes'),
  (4, 1, 'no'),
  (5, 1, 'maybe'),
  (7, 1, 'unanswered');

-- RSVPs for past concert (event_id=2): yes with instruments, others varied
INSERT INTO rsvps (account_id, event_id, state, instrument_id) VALUES
  (1, 2, 'yes',        15),
  (2, 2, 'yes',        10),
  (3, 2, 'yes',         3),
  (4, 2, 'no',       NULL),
  (5, 2, 'maybe',    NULL),
  (7, 2, 'unanswered', NULL);

-- RSVPs for past other event (event_id=3)
INSERT INTO rsvps (account_id, event_id, state) VALUES
  (1, 3, 'yes'),
  (2, 3, 'yes'),
  (3, 3, 'no'),
  (4, 3, 'no'),
  (5, 3, 'yes'),
  (7, 3, 'unanswered');

-- RSVPs for upcoming rehearsal (event_id=4)
INSERT INTO rsvps (account_id, event_id, state) VALUES
  (1, 4, 'yes'),
  (2, 4, 'yes'),
  (3, 4, 'yes'),
  (4, 4, 'no'),
  (5, 4, 'unanswered'),
  (7, 4, 'maybe');

-- RSVPs for upcoming concert (event_id=5): yes with instruments
INSERT INTO rsvps (account_id, event_id, state, instrument_id) VALUES
  (1, 5, 'yes',        15),
  (2, 5, 'yes',        10),
  (3, 5, 'yes',         3),
  (4, 5, 'yes',         1),
  (5, 5, 'maybe',    NULL),
  (7, 5, 'unanswered', NULL);

-- RSVPs for upcoming other event (event_id=6)
INSERT INTO rsvps (account_id, event_id, state) VALUES
  (1, 6, 'yes'),
  (2, 6, 'no'),
  (3, 6, 'maybe'),
  (4, 6, 'unanswered'),
  (5, 6, 'yes'),
  (7, 6, 'yes');

-- Field responses for upcoming other event: Alice and Emma answered "yes" with responses
INSERT INTO rsvp_field_responses (rsvp_id, event_field_id, value) VALUES
  -- Alice's responses (rsvp for account=1, event=6 — rsvp_id depends on insert order)
  ((SELECT id FROM rsvps WHERE account_id=1 AND event_id=6), 1, '1'),
  ((SELECT id FROM rsvps WHERE account_id=1 AND event_id=6), 2, 'Je serai là avec ma famille'),
  -- Emma's responses (account=5)
  ((SELECT id FROM rsvps WHERE account_id=5 AND event_id=6), 1, '2'),
  -- Gabrielle's responses (account=7)
  ((SELECT id FROM rsvps WHERE account_id=7 AND event_id=6), 1, '3');
