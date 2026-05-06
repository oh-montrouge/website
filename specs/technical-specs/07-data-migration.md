# OHM Website — Data Migration

> **Depends on:** [Data Model](00-data-model.md), [Configuration and Bootstrap](02-configuration.md)
>
> This spec covers the one-time migration from OHM Agenda (MySQL) and the seasonal Google
> Sheets to the new PostgreSQL schema. It is implementation guidance, not a feature spec.

---

## Sources

| Source | Location | Content |
|--------|----------|---------|
| OHM Agenda schema | `current/ohm-agenda/db_schema.sql` | MySQL DDL |
| OHM Agenda export | `current/ohm-agenda/db_export.sql` | Full data dump |
| Google Sheet exports | `current/gdrive-liste-musiciens/` | 4 seasonal XLSX files (2022–2026) |

---

## Pre-Migration Steps

### 1. Encoding repair

OHM Agenda data contains UTF-8/ISO-8859-1 mojibake (~30% of musician names and most
concert titles). Examples: "FranÃ§oise" → "Françoise", "RÃ©my" → "Rémy".

Repair before insertion: detect the encoding mismatch and re-encode to UTF-8. Run a
post-repair verification query to flag any remaining non-ASCII sequences that are not valid
French characters.

### 2. Google Sheet staging

Import all four XLSX files into a temporary staging table. The main data is on the "Fichier"
sheet (sheet index 1) in all files.

Column mapping across files (some columns appear only in later files; treat as NULL if
absent):

| Staging column | Source column | Notes |
|----------------|--------------|-------|
| nom | Nom | — |
| prenom | Prénom | — |
| instrument | Instrument | — |
| email | Email | — |
| phone_mobile | Port | Mobile; 92–98% populated |
| phone_home | Tel | Home; 21–25% populated — fallback only |
| adresse | Adresse | — |
| cp | CP | Stored as float (e.g. 92240.0); cast to integer then text |
| commune | Commune | — |
| birth_date | DN | Excel serial — convert: `date(1899,12,30) + timedelta(days=int(serial))` |
| inscription | Inscription | 22-23/23-24: text "2017-2018"; 24-25: text; 25-26: Excel serial |
| saison_label | (derived from filename) | e.g. "2022/2023" from "Liste musiciens_22_23.xlsx" |

**Address concatenation:** build a single address string:
`TRIM(adresse) || ', ' || CAST(cp AS INT)::TEXT || ' ' || commune`
where any null component is omitted gracefully.

**Inscription normalisation:** unify all formats to a season label string ("YYYY/YYYY"):
- Text "2017-2018" → "2017/2018"
- Excel serial → convert to date → derive season year from month (Sept–Aug)

### 3. Referential integrity pre-check

Before migrating, verify in the OHM Agenda export:
- Every `ID_CONCERT` in `presence_concert` exists in `concert`
- Every `ID_MUSICIEN` in `presence_*` exists in `musicien`
- Every `ID_SAISON` in `cotisation` exists in `saison`
- No duplicate `(ID_MUSICIEN, ID_SAISON)` in `cotisation`

Flag and resolve any violations before proceeding.

---

## Migration Steps

### Step 1 — Seasons (`saison` → `seasons`)

Migrate the 8 OHM Agenda seasons plus any additional seasons referenced by the Google Sheet
`Inscription` column that predate 2018/2019.

| OHM Agenda column | New column | Transformation |
|-------------------|-----------|----------------|
| ID_SAISON | id | Preserve |
| LIBELLE | label | Strip "Saison " prefix if present; normalise to "YYYY/YYYY" |
| — | start_date | Infer: September 1 of the first year (e.g. 2018-09-01 for "2018/2019") |
| — | end_date | Infer: August 31 of the second year (e.g. 2019-08-31 for "2018/2019") |
| — | is_current | Set false for all; admin designates current season post-migration |

For any season label found in the Google Sheet `Inscription` column that does not exist in
OHM Agenda, create a new row with inferred start/end dates and `is_current = false`. Assign
a new ID beyond the OHM Agenda range.

**Note:** `is_current` is intentionally left false for all migrated seasons. The admin
designates the current season as part of initial setup (see
[Configuration spec](02-configuration.md)).

---

### Step 2 — Instruments (`instrument` → `instruments`)

Direct transfer. 15 rows, no transformation required.

| OHM Agenda column | New column | Transformation |
|-------------------|-----------|----------------|
| ID_INSTRUMENT | id | Preserve |
| NOM | name | Direct; verify UTF-8 is clean |

Verify that "Chef d'orchestre" (ID 15) is present. If the export uses "Chef" as the label,
rename to "Chef d'orchestre" during migration.

---

### Step 3 — Accounts (`musicien` + Google Sheet → `accounts`)

#### 3a. Migrate base fields from musicien

| OHM Agenda column | New column | Transformation |
|-------------------|-----------|----------------|
| ID_MUSICIEN | id | Preserve |
| PRENOM | first_name | Repair encoding |
| NOM | last_name | Repair encoding |
| ID_INSTRUMENT | main_instrument_id | Preserve FK |
| ACTIF=1 | status | `'active'` |
| ACTIF=0 | status | `'anonymized'` — see below |

**ACTIF=0 accounts (~60 rows):** These are former members. Set `status = 'anonymized'`,
generate an anonymization token (CSPRNG, 32 bytes, base64url), and set all personal fields
to NULL. Do not populate email, phone, address, or birth_date for these accounts.

All other nullable fields (`email`, `password_hash`, `birth_date`, `parental_consent_uri`,
`phone`, `address`) default to NULL. `phone_address_consent` defaults to false.
`processing_restricted` defaults to false.

**Flagged row:** ID_MUSICIEN=124, NOM='to-be-updated' — flag for manual review post-migration.

#### 3b. Enrich active accounts from Google Sheet

For each `active` account, attempt to match a row in the Google Sheet staging table using
`LOWER(TRIM(first_name)) = LOWER(TRIM(prenom)) AND LOWER(TRIM(last_name)) = LOWER(TRIM(nom))`.

When a match is found, populate:

| Staging column | New column | Notes |
|----------------|-----------|-------|
| email | email | Take from most recent season's sheet where non-null |
| phone_mobile ?? phone_home | phone | Use `phone_mobile` (Port); fall back to `phone_home` (Tel) if mobile is null |
| (concatenated) | address | `adresse || ', ' || cp || ' ' || commune`; null if all three are null |
| birth_date | birth_date | Convert Excel serial to DATE |

`phone_address_consent`: set to `true` for active accounts where phone or address was
populated from the sheet. These musicians have implicitly consented by providing this data
previously; the invite flow will re-confirm consent explicitly when they first log in.

**Unmatched active accounts:** leave personal fields null; admin provisions email and sends
invite manually.

---

### Step 4 — Roles seed

Insert a single row into `roles`:

```sql
INSERT INTO roles (id, name) VALUES (1, 'admin');
```

Insert `account_roles` rows for the initial admins. **Admin accounts must be specified
manually** — OHM Agenda has no role tracking. Obtain the list from stakeholders and insert
one row per admin:

```sql
INSERT INTO account_roles (account_id, role_id)
VALUES (<admin_id>, 1);
```

---

### Step 5 — Fee Payments (`cotisation` → `fee_payments`)

#### 5a. Migrate cotisation records (PAIEMENT=1 only)

| OHM Agenda column | New column | Transformation |
|-------------------|-----------|----------------|
| ID_COTISATION | id | Preserve |
| ID_SAISON | season_id | Preserve FK |
| ID_MUSICIEN | account_id | Preserve FK → accounts.id |
| — | amount | 30.00 (default for all historical records) |
| — | payment_date | Season start_date (September 1 of the season) |
| — | payment_type | `'chèque'` (migration default; actual method unknown) |
| — | comment | `'Migré depuis OHM Agenda'` |

Skip rows where `PAIEMENT = 0`.

#### 5b. Ensure first fee_payment from Google Sheet Inscription

For each active account matched to a Google Sheet row:

1. Read the `Inscription` column (normalised season label, e.g. "2017/2018").
2. Look up the corresponding `seasons` row.
3. Check whether a `fee_payments` row already exists for `(account_id, season_id)`.
4. If not: insert a fee_payment with the same defaults (amount=30.00, type='chèque',
   comment='Migré depuis Google Sheet — première inscription').

This ensures every active account has at least one fee_payment, making first inscription
date derivable for all musicians.

---

### Step 6 — Events (`concert` + `repetition` + `evenement` → `events`)

The three old event tables are consolidated into a single `events` table. IDs are remapped
with offsets to avoid collisions:

| Source table | ID range | event_type | Offset |
|-------------|---------|------------|--------|
| concert | 1–79 | `concert` | 0 (IDs preserved) |
| repetition | 1–5 | `rehearsal` | +10000 |
| evenement | 1 | `other` | +20000 |

Column mapping (same pattern for all three source tables):

| Source columns | New column | Transformation |
|---------------|-----------|----------------|
| NOM_CONCERT / NOM_REPET / NOM_EVENT | name | Repair encoding |
| DATE_* + HEURE_* | datetime | Combine into TIMESTAMPTZ; assume Europe/Paris local time; convert to UTC (see ADR 002) |
| — | event_type | Set per source table (above) |
| HEURE_RDV, LIEU, TENUE, INFOS, PROGRAMME, NBRE_PLACE, ADRESSE, TARIF | description | Concatenate and convert HTML to markdown |

**Dropped columns** (no equivalent in new model): ID_SAISON.

---

### Step 7 — RSVPs (`presence_concert` + `presence_repet` → `rsvps`)

#### Status mapping

| Old STATUT | New state | Notes |
|------------|----------|-------|
| Present | `yes` | — |
| Absent | `no` | — |
| Incertain | `maybe` | — |
| Absent Non Repondu | `unanswered` | No response recorded |
| Present Non Repondu | `unanswered` | No response recorded |
| Hors Effectif | *skip* | Not a valid participant; do not migrate this row |

#### presence_concert migration

| OHM Agenda column | New column | Transformation |
|-------------------|-----------|----------------|
| ID_PRESENCE | id | Preserve |
| ID_CONCERT | event_id | Preserve (concert IDs unchanged) |
| ID_MUSICIEN | account_id | Preserve FK |
| ID_INSTRUMENT | instrument_id | Set non-null only when state = `'yes'`; null otherwise |
| STATUT | state | Map per table above |

#### presence_repet migration

Same pattern; event_id = ID_REPET + 10000. `instrument_id` is always null (rehearsals do
not collect instrument selection).

#### presence_event migration

The single "other" event (evenement ID 1, mapped to events.id = 20001) has 42 attendance
rows. The 14 `NB_CHOIX_*` columns are dropped. Migrate only:

| OHM Agenda column | New column | Transformation |
|-------------------|-----------|----------------|
| ID_PRESENCE_EVENT | id | Preserve (with offset if needed to avoid rsvp ID collision) |
| ID_EVENT | event_id | ID_EVENT + 20000 |
| ID_MUSICIEN | account_id | Preserve FK |
| NB_PRESENT | state | NB_PRESENT > 0 → `'yes'`; NB_PRESENT = 0 → `'no'` |

No `rsvp_field_responses` are created for this event; the `NB_CHOIX_*` data is discarded.

---

### Step 8 — WordPress tables

All `wp_*` tables in the OHM Agenda export are dropped. No data of operational value.

---

### Step 9 — Sequence Reset

IDs preserved from OHM Agenda leave PostgreSQL sequences pointing at 1. Reset each sequence
to the current maximum before the application starts inserting rows:

```sql
SELECT setval('accounts_id_seq',              (SELECT MAX(id) FROM accounts));
SELECT setval('seasons_id_seq',               (SELECT MAX(id) FROM seasons));
SELECT setval('instruments_id_seq',           (SELECT MAX(id) FROM instruments));
SELECT setval('fee_payments_id_seq',          (SELECT MAX(id) FROM fee_payments));
SELECT setval('events_id_seq',                (SELECT MAX(id) FROM events));
SELECT setval('rsvps_id_seq',                 (SELECT MAX(id) FROM rsvps));
SELECT setval('roles_id_seq',                 (SELECT MAX(id) FROM roles));
```

Run this after all insert steps and before the verification queries. Tables populated only
after go-live (`invite_tokens`, `password_reset_tokens`, `event_fields`, etc.) have no
preserved IDs; their sequences start at 1 as normal.

**Sequence name convention:** Pop generates sequences as `{table}_{column}_seq`. Verify
actual names with `\ds` in psql if the migration tool uses a different convention.

---

## Verification Queries

Run after migration is complete:

```sql
-- 1. Every active account has an email
SELECT COUNT(*) FROM accounts WHERE status = 'active' AND email IS NULL;
-- Expected: small number (unmatched accounts); review manually

-- 2. Every active account has at least one fee_payment (for first inscription date)
SELECT COUNT(*) FROM accounts a
WHERE a.status = 'active'
  AND NOT EXISTS (SELECT 1 FROM fee_payments fp WHERE fp.account_id = a.id);
-- Expected: 0

-- 3. Exactly one current season (once admin designates it)
SELECT COUNT(*) FROM seasons WHERE is_current = true;
-- Expected: 1 (after admin setup)

-- 4. No duplicate RSVPs
SELECT account_id, event_id, COUNT(*) FROM rsvps
GROUP BY account_id, event_id HAVING COUNT(*) > 1;
-- Expected: 0 rows

-- 5. No duplicate fee payments per (account, season)
SELECT account_id, season_id, COUNT(*) FROM fee_payments
GROUP BY account_id, season_id HAVING COUNT(*) > 1;
-- Expected: 0 rows

-- 6. No anonymized account has personal data
SELECT COUNT(*) FROM accounts
WHERE status = 'anonymized'
  AND (first_name IS NOT NULL OR email IS NOT NULL OR phone IS NOT NULL);
-- Expected: 0

-- 7. Instrument_id null for non-concert RSVPs and non-yes concert RSVPs
SELECT COUNT(*) FROM rsvps r
JOIN events e ON e.id = r.event_id
WHERE r.instrument_id IS NOT NULL
  AND (e.event_type != 'concert' OR r.state != 'yes');
-- Expected: 0

-- 8. All events have valid event_type
SELECT COUNT(*) FROM events
WHERE event_type NOT IN ('concert', 'rehearsal', 'other');
-- Expected: 0
```

---

## Post-Migration Admin Steps

Before the application is usable, the following must be done manually via the admin UI:

1. **Designate current season** — pick the active season (2025/2026) and mark it as current.
2. **Review flagged accounts** — ID_MUSICIEN=124 (NOM='to-be-updated'); correct name.
3. **Review unmatched accounts** — active accounts with no email; decide whether to send invites or leave pending.
4. **Send invite links** — for all active accounts lacking a password_hash (all of them), generate and send invite links so musicians can set their password and confirm consent.
5. **Verify fee payment amounts** — migrated records use 30€ default; 2024/2025 season may differ; correct manually if needed.

---

## Dropped Data Summary

The following OHM Agenda data has no equivalent in the new model and is not migrated:

| Table | Columns | Rationale |
|-------|---------|-----------|
| concert | HEURE_RDV, LIEU, TENUE, INFOS, PROGRAMME | Operational details not in V1 scope |
| repetition | NBRE_PLACE, INFOS, PROGRAMME | Not in V1 scope |
| evenement | LIEU, ID_SAISON, INFOS, ADRESSE, TARIF | Not in V1 scope |
| presence_event | NB_CHOIX_A1–D2 | Semantics unclear; dropped by decision |
| cotisation | PAIEMENT=0 rows | No payment made; not a payment record |
| Google Sheet | job (Prof), birth_place (LN), nationality (Nat) | Out of V1 scope |
| Google Sheet | autorisation_* columns | Always empty; consent re-collected via invite flow |
| All wp_* tables | (all) | WordPress plugin artefacts; no operational value |
