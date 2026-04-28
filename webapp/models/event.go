package models

import (
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
)

// EventDetailRow is the result of a single-event lookup (GetByID).
type EventDetailRow struct {
	ID          int64        `db:"id"`
	Name        string       `db:"name"`
	Datetime    time.Time    `db:"datetime"`
	EventType   string       `db:"event_type"`
	Description nulls.String `db:"description"`
}

// EventListRow is the result of list queries that include the viewer's own RSVP state.
type EventListRow struct {
	ID        int64        `db:"id"`
	Name      string       `db:"name"`
	Datetime  time.Time    `db:"datetime"`
	EventType string       `db:"event_type"`
	RSVPState nulls.String `db:"rsvp_state"` // null when no RSVP record exists for viewer
}

// RSVPRow is the raw RSVP record (used for own-RSVP lookup).
type RSVPRow struct {
	ID           int64       `db:"id"`
	AccountID    int64       `db:"account_id"`
	EventID      int64       `db:"event_id"`
	State        string      `db:"state"`
	InstrumentID nulls.Int64 `db:"instrument_id"`
}

// RSVPListRow is one row in the full RSVP list for an event detail page.
type RSVPListRow struct {
	RSVPID             int64        `db:"rsvp_id"`
	AccountID          int64        `db:"account_id"`
	FirstName          nulls.String `db:"first_name"`
	LastName           nulls.String `db:"last_name"`
	AnonymizationToken nulls.String `db:"anonymization_token"`
	State              string       `db:"state"`
	InstrumentID       nulls.Int64  `db:"instrument_id"`
	RSVPInstrumentName nulls.String `db:"rsvp_instrument_name"` // concert instrument selected for this RSVP
	MainInstrumentName string       `db:"main_instrument_name"` // musician's primary instrument
}

// EventFieldRow is a custom field definition on an other-type event.
type EventFieldRow struct {
	ID        int64  `db:"id"`
	EventID   int64  `db:"event_id"`
	Label     string `db:"label"`
	FieldType string `db:"field_type"`
	Required  bool   `db:"required"`
	Position  int    `db:"position"`
}

// EventFieldChoiceRow is one selectable option for a choice-type field.
type EventFieldChoiceRow struct {
	ID           int64  `db:"id"`
	EventFieldID int64  `db:"event_field_id"`
	Label        string `db:"label"`
	Position     int    `db:"position"`
}

// RSVPFieldResponseRow is a musician's response to one custom event field.
type RSVPFieldResponseRow struct {
	RSVPID       int64  `db:"rsvp_id"`
	EventFieldID int64  `db:"event_field_id"`
	FieldLabel   string `db:"field_label"`
	FieldType    string `db:"field_type"`
	Value        string `db:"value"`
}

// EventStore is the production implementation of services.EventRepository.
type EventStore struct{}

func (EventStore) Create(tx *pop.Connection, name, eventType string, datetime time.Time) (int64, error) {
	var row struct {
		ID int64 `db:"id"`
	}
	err := tx.RawQuery(
		`INSERT INTO events (name, event_type, datetime) VALUES (?, ?, ?) RETURNING id`,
		name, eventType, datetime,
	).First(&row)
	return row.ID, err
}

func (EventStore) GetByID(tx *pop.Connection, id int64) (*EventDetailRow, error) {
	var rows []EventDetailRow
	err := tx.RawQuery(
		`SELECT id, name, datetime, event_type, description FROM events WHERE id = ?`, id,
	).All(&rows)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

func (EventStore) Update(tx *pop.Connection, id int64, name, eventType string, datetime time.Time) error {
	return tx.RawQuery(
		`UPDATE events SET name=?, event_type=?, datetime=? WHERE id=?`,
		name, eventType, datetime, id,
	).Exec()
}

func (EventStore) Delete(tx *pop.Connection, id int64) error {
	return tx.RawQuery(`DELETE FROM events WHERE id=?`, id).Exec()
}

func (EventStore) ListUpcoming(tx *pop.Connection, accountID int64) ([]EventListRow, error) {
	var rows []EventListRow
	err := tx.RawQuery(`
		SELECT e.id, e.name, e.datetime, e.event_type, r.state AS rsvp_state
		FROM events e
		LEFT JOIN rsvps r ON r.event_id = e.id AND r.account_id = ?
		WHERE e.datetime >= NOW()
		ORDER BY e.datetime ASC`, accountID,
	).All(&rows)
	return rows, err
}

func (EventStore) ListAll(tx *pop.Connection, accountID int64) ([]EventListRow, error) {
	var rows []EventListRow
	err := tx.RawQuery(`
		SELECT e.id, e.name, e.datetime, e.event_type, r.state AS rsvp_state
		FROM events e
		LEFT JOIN rsvps r ON r.event_id = e.id AND r.account_id = ?
		ORDER BY e.datetime ASC`, accountID,
	).All(&rows)
	return rows, err
}

func (EventStore) DeleteFields(tx *pop.Connection, eventID int64) error {
	return tx.RawQuery(`DELETE FROM event_fields WHERE event_id=?`, eventID).Exec()
}

func (EventStore) AddField(tx *pop.Connection, eventID int64, label, fieldType string, required bool, position int) (int64, error) {
	var row struct {
		ID int64 `db:"id"`
	}
	err := tx.RawQuery(
		`INSERT INTO event_fields (event_id, label, field_type, required, position) VALUES (?, ?, ?, ?, ?) RETURNING id`,
		eventID, label, fieldType, required, position,
	).First(&row)
	return row.ID, err
}

func (EventStore) GetFieldByID(tx *pop.Connection, fieldID int64) (*EventFieldRow, error) {
	var rows []EventFieldRow
	err := tx.RawQuery(
		`SELECT id, event_id, label, field_type, required, position FROM event_fields WHERE id=?`, fieldID,
	).All(&rows)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

func (EventStore) UpdateField(tx *pop.Connection, fieldID int64, label, fieldType string, required bool, position int) error {
	return tx.RawQuery(
		`UPDATE event_fields SET label=?, field_type=?, required=?, position=? WHERE id=?`,
		label, fieldType, required, position, fieldID,
	).Exec()
}

func (EventStore) DeleteField(tx *pop.Connection, fieldID int64) error {
	return tx.RawQuery(`DELETE FROM event_fields WHERE id=?`, fieldID).Exec()
}

func (EventStore) ListFields(tx *pop.Connection, eventID int64) ([]EventFieldRow, error) {
	var rows []EventFieldRow
	err := tx.RawQuery(
		`SELECT id, event_id, label, field_type, required, position FROM event_fields WHERE event_id=? ORDER BY position ASC`,
		eventID,
	).All(&rows)
	return rows, err
}

func (EventStore) ListFieldChoices(tx *pop.Connection, fieldID int64) ([]EventFieldChoiceRow, error) {
	var rows []EventFieldChoiceRow
	err := tx.RawQuery(
		`SELECT id, event_field_id, label, position FROM event_field_choices WHERE event_field_id=? ORDER BY position ASC`,
		fieldID,
	).All(&rows)
	return rows, err
}

func (EventStore) CountFieldResponses(tx *pop.Connection, fieldID int64) (int, error) {
	var result struct {
		Count int `db:"count"`
	}
	err := tx.RawQuery(
		`SELECT COUNT(*) AS count FROM rsvp_field_responses WHERE event_field_id=?`, fieldID,
	).First(&result)
	return result.Count, err
}

func (EventStore) AddFieldChoice(tx *pop.Connection, fieldID int64, label string, position int) (int64, error) {
	var row struct {
		ID int64 `db:"id"`
	}
	err := tx.RawQuery(
		`INSERT INTO event_field_choices (event_field_id, label, position) VALUES (?, ?, ?) RETURNING id`,
		fieldID, label, position,
	).First(&row)
	return row.ID, err
}

func (EventStore) DeleteFieldChoices(tx *pop.Connection, fieldID int64) error {
	return tx.RawQuery(`DELETE FROM event_field_choices WHERE event_field_id=?`, fieldID).Exec()
}

// RSVPStore is the production implementation of services.RSVPRepository.
type RSVPStore struct{}

func (RSVPStore) SeedForEvent(tx *pop.Connection, eventID int64) error {
	return tx.RawQuery(`
		INSERT INTO rsvps (account_id, event_id, state)
		SELECT id, ?, 'unanswered'
		FROM accounts
		WHERE status = 'active'
		ON CONFLICT (account_id, event_id) DO NOTHING`, eventID,
	).Exec()
}

func (RSVPStore) SeedForAccount(tx *pop.Connection, accountID int64) error {
	return tx.RawQuery(`
		INSERT INTO rsvps (account_id, event_id, state)
		SELECT ?, id, 'unanswered'
		FROM events
		WHERE datetime > NOW()
		ON CONFLICT (account_id, event_id) DO NOTHING`, accountID,
	).Exec()
}

func (RSVPStore) GetByAccountAndEvent(tx *pop.Connection, accountID, eventID int64) (*RSVPRow, error) {
	var rows []RSVPRow
	err := tx.RawQuery(
		`SELECT id, account_id, event_id, state, instrument_id FROM rsvps WHERE account_id=? AND event_id=?`,
		accountID, eventID,
	).All(&rows)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

func (RSVPStore) Update(tx *pop.Connection, rsvpID int64, state string, instrumentID *int64) error {
	if instrumentID == nil {
		return tx.RawQuery(
			`UPDATE rsvps SET state=?, instrument_id=NULL WHERE id=?`, state, rsvpID,
		).Exec()
	}
	return tx.RawQuery(
		`UPDATE rsvps SET state=?, instrument_id=? WHERE id=?`, state, *instrumentID, rsvpID,
	).Exec()
}

func (RSVPStore) DeleteByAccount(tx *pop.Connection, accountID int64) error {
	return tx.RawQuery(`DELETE FROM rsvps WHERE account_id=?`, accountID).Exec()
}

func (RSVPStore) ClearFieldResponses(tx *pop.Connection, rsvpID int64) error {
	return tx.RawQuery(`DELETE FROM rsvp_field_responses WHERE rsvp_id=?`, rsvpID).Exec()
}

func (RSVPStore) ListForEvent(tx *pop.Connection, eventID int64) ([]RSVPListRow, error) {
	var rows []RSVPListRow
	err := tx.RawQuery(`
		SELECT
		    r.id   AS rsvp_id,
		    r.account_id,
		    a.first_name,
		    a.last_name,
		    a.anonymization_token,
		    r.state,
		    r.instrument_id,
		    ri.name AS rsvp_instrument_name,
		    mi.name AS main_instrument_name
		FROM rsvps r
		JOIN accounts a ON a.id = r.account_id
		LEFT JOIN instruments ri ON ri.id = r.instrument_id
		JOIN instruments mi ON mi.id = a.main_instrument_id
		WHERE r.event_id = ?
		ORDER BY a.last_name ASC NULLS LAST, a.first_name ASC NULLS LAST`,
		eventID,
	).All(&rows)
	return rows, err
}

func (RSVPStore) ResetYesRSVPs(tx *pop.Connection, eventID int64) error {
	return tx.RawQuery(
		`UPDATE rsvps SET state='unanswered', instrument_id=NULL WHERE event_id=? AND state='yes'`,
		eventID,
	).Exec()
}

func (RSVPStore) ClearInstruments(tx *pop.Connection, eventID int64) error {
	return tx.RawQuery(
		`UPDATE rsvps SET instrument_id=NULL WHERE event_id=? AND instrument_id IS NOT NULL`,
		eventID,
	).Exec()
}

func (RSVPStore) AddFieldResponse(tx *pop.Connection, rsvpID, fieldID int64, value string) error {
	return tx.RawQuery(`
		INSERT INTO rsvp_field_responses (rsvp_id, event_field_id, value)
		VALUES (?, ?, ?)
		ON CONFLICT (rsvp_id, event_field_id) DO UPDATE SET value=EXCLUDED.value`,
		rsvpID, fieldID, value,
	).Exec()
}

func (RSVPStore) ListFieldResponses(tx *pop.Connection, eventID int64) ([]RSVPFieldResponseRow, error) {
	var rows []RSVPFieldResponseRow
	err := tx.RawQuery(`
		SELECT
		    rfr.rsvp_id,
		    rfr.event_field_id,
		    ef.label AS field_label,
		    ef.field_type,
		    rfr.value
		FROM rsvp_field_responses rfr
		JOIN event_fields ef ON ef.id = rfr.event_field_id
		WHERE ef.event_id = ?
		ORDER BY rfr.rsvp_id, ef.position ASC`,
		eventID,
	).All(&rows)
	return rows, err
}
