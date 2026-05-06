package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	evPast   = time.Now().UTC().Add(-48 * time.Hour)
	evFuture = time.Now().UTC().Add(48 * time.Hour)
)

// insertEventAccount inserts one instrument and one active account, returning the account ID.
func insertEventAccount(t *testing.T) int64 {
	t.Helper()
	var instrRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`INSERT INTO instruments (name) VALUES ('Violon') RETURNING id`).First(&instrRow))
	var accRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(
		`INSERT INTO accounts (email, password_hash, main_instrument_id, status) VALUES ('ev@example.com', 'h', ?, 'active') RETURNING id`,
		instrRow.ID,
	).First(&accRow))
	return accRow.ID
}

// insertEventAccounts inserts one shared instrument and one active account per email,
// returning the instrument ID and the account IDs in the same order as emails.
func insertEventAccounts(t *testing.T, emails ...string) (instrID int64, ids []int64) {
	t.Helper()
	var instrRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`INSERT INTO instruments (name) VALUES ('Violon') RETURNING id`).First(&instrRow))
	for _, email := range emails {
		var acc struct {
			ID int64 `db:"id"`
		}
		require.NoError(t, DB.RawQuery(
			`INSERT INTO accounts (email, password_hash, main_instrument_id, status) VALUES (?, 'h', ?, 'active') RETURNING id`,
			email, instrRow.ID,
		).First(&acc))
		ids = append(ids, acc.ID)
	}
	return instrRow.ID, ids
}

func insertTestEvent(t *testing.T, name, eventType string, dt time.Time) int64 {
	t.Helper()
	id, err := EventStore{}.Create(DB, name, eventType, dt)
	require.NoError(t, err)
	return id
}

// insertRSVPRow inserts one RSVP row and returns its ID.
func insertRSVPRow(t *testing.T, accountID, eventID int64, state string) int64 {
	t.Helper()
	var row struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(
		`INSERT INTO rsvps (account_id, event_id, state) VALUES (?, ?, ?) RETURNING id`,
		accountID, eventID, state,
	).First(&row))
	return row.ID
}

// ── EventStore ────────────────────────────────────────────────────────────────

func TestEventStore_Create(t *testing.T) {
	truncateAll(t)
	store := EventStore{}
	id, err := store.Create(DB, "Concert d'automne", "concert", evFuture)
	require.NoError(t, err)
	assert.Positive(t, id)

	row, err := store.GetByID(DB, id)
	require.NoError(t, err)
	require.NotNil(t, row)
	assert.Equal(t, "Concert d'automne", row.Name)
	assert.Equal(t, "concert", row.EventType)
}

func TestEventStore_GetByID_Found(t *testing.T) {
	truncateAll(t)
	id := insertTestEvent(t, "Répétition", "other", evFuture)

	row, err := EventStore{}.GetByID(DB, id)
	require.NoError(t, err)
	require.NotNil(t, row)
	assert.Equal(t, id, row.ID)
	assert.Equal(t, "Répétition", row.Name)
}

func TestEventStore_GetByID_NotFound(t *testing.T) {
	truncateAll(t)
	row, err := EventStore{}.GetByID(DB, 99999)
	require.NoError(t, err)
	assert.Nil(t, row)
}

func TestEventStore_Update(t *testing.T) {
	truncateAll(t)
	store := EventStore{}
	id := insertTestEvent(t, "Old Name", "concert", evPast)

	require.NoError(t, store.Update(DB, id, "New Name", "other", evFuture))

	row, err := store.GetByID(DB, id)
	require.NoError(t, err)
	require.NotNil(t, row)
	assert.Equal(t, "New Name", row.Name)
	assert.Equal(t, "other", row.EventType)
	assert.True(t, row.Datetime.After(time.Now()))
}

func TestEventStore_Delete(t *testing.T) {
	truncateAll(t)
	store := EventStore{}
	id := insertTestEvent(t, "To Delete", "concert", evFuture)

	require.NoError(t, store.Delete(DB, id))

	row, err := store.GetByID(DB, id)
	require.NoError(t, err)
	assert.Nil(t, row)
}

func TestEventStore_ListUpcoming_Empty(t *testing.T) {
	truncateAll(t)
	rows, err := EventStore{}.ListUpcoming(DB, 0)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestEventStore_ListUpcoming_FiltersOld(t *testing.T) {
	truncateAll(t)
	insertTestEvent(t, "Past", "concert", evPast)
	futureID := insertTestEvent(t, "Future", "concert", evFuture)

	rows, err := EventStore{}.ListUpcoming(DB, 0)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, futureID, rows[0].ID)
}

func TestEventStore_ListUpcoming_IncludesRSVPState(t *testing.T) {
	truncateAll(t)
	accountID := insertEventAccount(t)
	eventID := insertTestEvent(t, "Concert", "concert", evFuture)
	insertRSVPRow(t, accountID, eventID, "yes")

	rows, err := EventStore{}.ListUpcoming(DB, accountID)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.True(t, rows[0].RSVPState.Valid)
	assert.Equal(t, "yes", rows[0].RSVPState.String)

	rows, err = EventStore{}.ListUpcoming(DB, 0)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.False(t, rows[0].RSVPState.Valid)
}

func TestEventStore_ListAll_IncludesPastAndFuture(t *testing.T) {
	truncateAll(t)
	insertTestEvent(t, "Past", "concert", evPast)
	insertTestEvent(t, "Future", "concert", evFuture)

	rows, err := EventStore{}.ListAll(DB, 0)
	require.NoError(t, err)
	assert.Len(t, rows, 2)
}

func TestEventStore_DeleteFields(t *testing.T) {
	truncateAll(t)
	store := EventStore{}
	eventID := insertTestEvent(t, "Workshop", "other", evFuture)

	_, err := store.AddField(DB, eventID, "T-shirt size", "text", false, 1)
	require.NoError(t, err)

	require.NoError(t, store.DeleteFields(DB, eventID))

	fields, err := store.ListFields(DB, eventID)
	require.NoError(t, err)
	assert.Empty(t, fields)
}

func TestEventStore_AddField_GetFieldByID(t *testing.T) {
	truncateAll(t)
	store := EventStore{}
	eventID := insertTestEvent(t, "Workshop", "other", evFuture)

	fieldID, err := store.AddField(DB, eventID, "Diet", "choice", true, 1)
	require.NoError(t, err)
	assert.Positive(t, fieldID)

	field, err := store.GetFieldByID(DB, fieldID)
	require.NoError(t, err)
	require.NotNil(t, field)
	assert.Equal(t, fieldID, field.ID)
	assert.Equal(t, eventID, field.EventID)
	assert.Equal(t, "Diet", field.Label)
	assert.Equal(t, "choice", field.FieldType)
	assert.True(t, field.Required)
	assert.Equal(t, 1, field.Position)
}

func TestEventStore_GetFieldByID_NotFound(t *testing.T) {
	truncateAll(t)
	field, err := EventStore{}.GetFieldByID(DB, 99999)
	require.NoError(t, err)
	assert.Nil(t, field)
}

func TestEventStore_UpdateField(t *testing.T) {
	truncateAll(t)
	store := EventStore{}
	eventID := insertTestEvent(t, "Workshop", "other", evFuture)
	fieldID, err := store.AddField(DB, eventID, "OldLabel", "text", false, 1)
	require.NoError(t, err)

	require.NoError(t, store.UpdateField(DB, fieldID, "NewLabel", "choice", true, 2))

	field, err := store.GetFieldByID(DB, fieldID)
	require.NoError(t, err)
	require.NotNil(t, field)
	assert.Equal(t, "NewLabel", field.Label)
	assert.Equal(t, "choice", field.FieldType)
	assert.True(t, field.Required)
	assert.Equal(t, 2, field.Position)
}

func TestEventStore_DeleteField(t *testing.T) {
	truncateAll(t)
	store := EventStore{}
	eventID := insertTestEvent(t, "Workshop", "other", evFuture)
	fieldID, err := store.AddField(DB, eventID, "Label", "text", false, 1)
	require.NoError(t, err)

	require.NoError(t, store.DeleteField(DB, fieldID))

	field, err := store.GetFieldByID(DB, fieldID)
	require.NoError(t, err)
	assert.Nil(t, field)
}

func TestEventStore_ListFields(t *testing.T) {
	truncateAll(t)
	store := EventStore{}
	eventID := insertTestEvent(t, "Workshop", "other", evFuture)

	_, err := store.AddField(DB, eventID, "B", "text", false, 2)
	require.NoError(t, err)
	_, err = store.AddField(DB, eventID, "A", "text", false, 1)
	require.NoError(t, err)

	fields, err := store.ListFields(DB, eventID)
	require.NoError(t, err)
	require.Len(t, fields, 2)
	assert.Equal(t, 1, fields[0].Position, "ordered by position ASC")
	assert.Equal(t, "A", fields[0].Label)
	assert.Equal(t, 2, fields[1].Position)
}

func TestEventStore_ListFieldChoices(t *testing.T) {
	truncateAll(t)
	store := EventStore{}
	eventID := insertTestEvent(t, "Workshop", "other", evFuture)
	fieldID, err := store.AddField(DB, eventID, "Diet", "choice", true, 1)
	require.NoError(t, err)

	_, err = store.AddFieldChoice(DB, fieldID, "Vegan", 2)
	require.NoError(t, err)
	_, err = store.AddFieldChoice(DB, fieldID, "Standard", 1)
	require.NoError(t, err)

	choices, err := store.ListFieldChoices(DB, fieldID)
	require.NoError(t, err)
	require.Len(t, choices, 2)
	assert.Equal(t, 1, choices[0].Position, "ordered by position ASC")
	assert.Equal(t, "Standard", choices[0].Label)
}

func TestEventStore_CountFieldResponses_Zero(t *testing.T) {
	truncateAll(t)
	store := EventStore{}
	eventID := insertTestEvent(t, "Workshop", "other", evFuture)
	fieldID, err := store.AddField(DB, eventID, "Q", "text", false, 1)
	require.NoError(t, err)

	n, err := store.CountFieldResponses(DB, fieldID)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestEventStore_CountFieldResponses_NonZero(t *testing.T) {
	truncateAll(t)
	store := EventStore{}
	accountID := insertEventAccount(t)
	eventID := insertTestEvent(t, "Workshop", "other", evFuture)

	fieldID, err := store.AddField(DB, eventID, "Q", "text", false, 1)
	require.NoError(t, err)
	rsvpID := insertRSVPRow(t, accountID, eventID, "unanswered")

	require.NoError(t, DB.RawQuery(
		`INSERT INTO rsvp_field_responses (rsvp_id, event_field_id, value) VALUES (?, ?, 'answer')`,
		rsvpID, fieldID,
	).Exec())

	n, err := store.CountFieldResponses(DB, fieldID)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
}

func TestEventStore_AddFieldChoice(t *testing.T) {
	truncateAll(t)
	store := EventStore{}
	eventID := insertTestEvent(t, "Workshop", "other", evFuture)
	fieldID, err := store.AddField(DB, eventID, "Diet", "choice", true, 1)
	require.NoError(t, err)

	choiceID, err := store.AddFieldChoice(DB, fieldID, "Halal", 1)
	require.NoError(t, err)
	assert.Positive(t, choiceID)

	choices, err := store.ListFieldChoices(DB, fieldID)
	require.NoError(t, err)
	require.Len(t, choices, 1)
	assert.Equal(t, "Halal", choices[0].Label)
}

func TestEventStore_DeleteFieldChoices(t *testing.T) {
	truncateAll(t)
	store := EventStore{}
	eventID := insertTestEvent(t, "Workshop", "other", evFuture)
	fieldID, err := store.AddField(DB, eventID, "Diet", "choice", true, 1)
	require.NoError(t, err)

	_, err = store.AddFieldChoice(DB, fieldID, "A", 1)
	require.NoError(t, err)
	_, err = store.AddFieldChoice(DB, fieldID, "B", 2)
	require.NoError(t, err)

	require.NoError(t, store.DeleteFieldChoices(DB, fieldID))

	choices, err := store.ListFieldChoices(DB, fieldID)
	require.NoError(t, err)
	assert.Empty(t, choices)
}

// ── RSVPStore ─────────────────────────────────────────────────────────────────

func TestRSVPStore_SeedForEvent(t *testing.T) {
	truncateAll(t)
	var instrRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`INSERT INTO instruments (name) VALUES ('Violon') RETURNING id`).First(&instrRow))

	var acc1 struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`INSERT INTO accounts (email, password_hash, main_instrument_id, status) VALUES ('a1@ev.com', 'h', ?, 'active') RETURNING id`, instrRow.ID).First(&acc1))
	var acc2 struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`INSERT INTO accounts (email, password_hash, main_instrument_id, status) VALUES ('a2@ev.com', 'h', ?, 'active') RETURNING id`, instrRow.ID).First(&acc2))
	require.NoError(t, DB.RawQuery(`INSERT INTO accounts (email, main_instrument_id, status, phone_address_consent, processing_restricted) VALUES ('a3@ev.com', ?, 'pending', false, false)`, instrRow.ID).Exec())

	eventID := insertTestEvent(t, "Concert", "concert", evFuture)
	store := RSVPStore{}

	require.NoError(t, store.SeedForEvent(DB, eventID))

	var cnt struct {
		Count int `db:"count"`
	}
	require.NoError(t, DB.RawQuery(`SELECT COUNT(*) AS count FROM rsvps WHERE event_id=?`, eventID).First(&cnt))
	assert.Equal(t, 2, cnt.Count, "one RSVP per active account only")

	require.NoError(t, store.SeedForEvent(DB, eventID))
	require.NoError(t, DB.RawQuery(`SELECT COUNT(*) AS count FROM rsvps WHERE event_id=?`, eventID).First(&cnt))
	assert.Equal(t, 2, cnt.Count, "ON CONFLICT DO NOTHING — idempotent")
}

func TestRSVPStore_SeedForAccount(t *testing.T) {
	truncateAll(t)
	accountID := insertEventAccount(t)
	insertTestEvent(t, "Future1", "concert", evFuture)
	insertTestEvent(t, "Future2", "concert", evFuture)
	insertTestEvent(t, "Past", "concert", evPast)

	store := RSVPStore{}
	require.NoError(t, store.SeedForAccount(DB, accountID))

	var cnt struct {
		Count int `db:"count"`
	}
	require.NoError(t, DB.RawQuery(`SELECT COUNT(*) AS count FROM rsvps WHERE account_id=?`, accountID).First(&cnt))
	assert.Equal(t, 2, cnt.Count, "one RSVP per future event only")

	require.NoError(t, store.SeedForAccount(DB, accountID))
	require.NoError(t, DB.RawQuery(`SELECT COUNT(*) AS count FROM rsvps WHERE account_id=?`, accountID).First(&cnt))
	assert.Equal(t, 2, cnt.Count, "ON CONFLICT DO NOTHING — idempotent")
}

func TestRSVPStore_GetByAccountAndEvent_Found(t *testing.T) {
	truncateAll(t)
	accountID := insertEventAccount(t)
	eventID := insertTestEvent(t, "Concert", "concert", evFuture)

	require.NoError(t, DB.RawQuery(
		`INSERT INTO rsvps (account_id, event_id, state) VALUES (?, ?, 'maybe')`,
		accountID, eventID,
	).Exec())

	rsvp, err := RSVPStore{}.GetByAccountAndEvent(DB, accountID, eventID)
	require.NoError(t, err)
	require.NotNil(t, rsvp)
	assert.Equal(t, accountID, rsvp.AccountID)
	assert.Equal(t, eventID, rsvp.EventID)
	assert.Equal(t, "maybe", rsvp.State)
}

func TestRSVPStore_GetByAccountAndEvent_NotFound(t *testing.T) {
	truncateAll(t)
	rsvp, err := RSVPStore{}.GetByAccountAndEvent(DB, 1, 2)
	require.NoError(t, err)
	assert.Nil(t, rsvp)
}

func TestRSVPStore_Update_NilInstrument(t *testing.T) {
	truncateAll(t)
	accountID := insertEventAccount(t)
	eventID := insertTestEvent(t, "Concert", "concert", evFuture)

	var rsvpRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(
		`INSERT INTO rsvps (account_id, event_id, state) VALUES (?, ?, 'unanswered') RETURNING id`,
		accountID, eventID,
	).First(&rsvpRow))

	require.NoError(t, RSVPStore{}.Update(DB, rsvpRow.ID, "no", nil))

	rsvp, err := RSVPStore{}.GetByAccountAndEvent(DB, accountID, eventID)
	require.NoError(t, err)
	assert.Equal(t, "no", rsvp.State)
	assert.False(t, rsvp.InstrumentID.Valid)
}

func TestRSVPStore_Update_WithInstrument(t *testing.T) {
	truncateAll(t)
	accountID := insertEventAccount(t)
	eventID := insertTestEvent(t, "Concert", "concert", evFuture)

	var instrRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`SELECT id FROM instruments LIMIT 1`).First(&instrRow))
	instrID := instrRow.ID

	var rsvpRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(
		`INSERT INTO rsvps (account_id, event_id, state) VALUES (?, ?, 'unanswered') RETURNING id`,
		accountID, eventID,
	).First(&rsvpRow))

	require.NoError(t, RSVPStore{}.Update(DB, rsvpRow.ID, "yes", &instrID))

	rsvp, err := RSVPStore{}.GetByAccountAndEvent(DB, accountID, eventID)
	require.NoError(t, err)
	assert.Equal(t, "yes", rsvp.State)
	assert.True(t, rsvp.InstrumentID.Valid)
	assert.Equal(t, instrID, rsvp.InstrumentID.Int64)
}

func TestRSVPStore_DeleteByAccount(t *testing.T) {
	truncateAll(t)
	accountID := insertEventAccount(t)
	eventID := insertTestEvent(t, "Concert", "concert", evFuture)

	require.NoError(t, DB.RawQuery(
		`INSERT INTO rsvps (account_id, event_id, state) VALUES (?, ?, 'unanswered')`,
		accountID, eventID,
	).Exec())

	require.NoError(t, RSVPStore{}.DeleteByAccount(DB, accountID))

	var cnt struct {
		Count int `db:"count"`
	}
	require.NoError(t, DB.RawQuery(`SELECT COUNT(*) AS count FROM rsvps WHERE account_id=?`, accountID).First(&cnt))
	assert.Equal(t, 0, cnt.Count)
}

func TestRSVPStore_ClearFieldResponses(t *testing.T) {
	truncateAll(t)
	_, accIDs := insertEventAccounts(t, "c1@ev.com", "c2@ev.com")
	acc1ID, acc2ID := accIDs[0], accIDs[1]

	eventID := insertTestEvent(t, "Workshop", "other", evFuture)
	fieldID, err := EventStore{}.AddField(DB, eventID, "Q", "text", false, 1)
	require.NoError(t, err)

	rsvp1ID := insertRSVPRow(t, acc1ID, eventID, "unanswered")
	rsvp2ID := insertRSVPRow(t, acc2ID, eventID, "unanswered")

	require.NoError(t, DB.RawQuery(`INSERT INTO rsvp_field_responses (rsvp_id, event_field_id, value) VALUES (?, ?, 'ans1')`, rsvp1ID, fieldID).Exec())
	require.NoError(t, DB.RawQuery(`INSERT INTO rsvp_field_responses (rsvp_id, event_field_id, value) VALUES (?, ?, 'ans2')`, rsvp2ID, fieldID).Exec())

	require.NoError(t, RSVPStore{}.ClearFieldResponses(DB, rsvp1ID))

	var cnt struct {
		Count int `db:"count"`
	}
	require.NoError(t, DB.RawQuery(`SELECT COUNT(*) AS count FROM rsvp_field_responses WHERE rsvp_id=?`, rsvp1ID).First(&cnt))
	assert.Equal(t, 0, cnt.Count, "rsvp1 responses cleared")

	require.NoError(t, DB.RawQuery(`SELECT COUNT(*) AS count FROM rsvp_field_responses WHERE rsvp_id=?`, rsvp2ID).First(&cnt))
	assert.Equal(t, 1, cnt.Count, "rsvp2 responses untouched")
}

func TestRSVPStore_ListForEvent(t *testing.T) {
	truncateAll(t)
	accountID := insertEventAccount(t)
	eventID := insertTestEvent(t, "Concert", "concert", evFuture)
	insertRSVPRow(t, accountID, eventID, "yes")

	rows, err := RSVPStore{}.ListForEvent(DB, eventID)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, accountID, rows[0].AccountID)
	assert.Equal(t, "yes", rows[0].State)
	assert.Equal(t, "Violon", rows[0].MainInstrumentName)
	assert.False(t, rows[0].RSVPInstrumentName.Valid)
}

func TestRSVPStore_ResetYesRSVPs(t *testing.T) {
	truncateAll(t)
	instrID, accIDs := insertEventAccounts(t, "r1@ev.com", "r2@ev.com", "r3@ev.com")
	acc1ID, acc2ID, acc3ID := accIDs[0], accIDs[1], accIDs[2]

	eventID := insertTestEvent(t, "Concert", "concert", evFuture)

	require.NoError(t, DB.RawQuery(`INSERT INTO rsvps (account_id, event_id, state, instrument_id) VALUES (?, ?, 'yes', ?)`, acc1ID, eventID, instrID).Exec())
	require.NoError(t, DB.RawQuery(`INSERT INTO rsvps (account_id, event_id, state) VALUES (?, ?, 'no')`, acc2ID, eventID).Exec())
	require.NoError(t, DB.RawQuery(`INSERT INTO rsvps (account_id, event_id, state) VALUES (?, ?, 'maybe')`, acc3ID, eventID).Exec())

	require.NoError(t, RSVPStore{}.ResetYesRSVPs(DB, eventID))

	yes, err := RSVPStore{}.GetByAccountAndEvent(DB, acc1ID, eventID)
	require.NoError(t, err)
	assert.Equal(t, "unanswered", yes.State)
	assert.False(t, yes.InstrumentID.Valid, "instrument cleared on reset")

	no, err := RSVPStore{}.GetByAccountAndEvent(DB, acc2ID, eventID)
	require.NoError(t, err)
	assert.Equal(t, "no", no.State, "no RSVP untouched")

	maybe, err := RSVPStore{}.GetByAccountAndEvent(DB, acc3ID, eventID)
	require.NoError(t, err)
	assert.Equal(t, "maybe", maybe.State, "maybe RSVP untouched")
}

func TestRSVPStore_ClearInstruments(t *testing.T) {
	truncateAll(t)
	instrID, accIDs := insertEventAccounts(t, "ci1@ev.com", "ci2@ev.com")
	acc1ID, acc2ID := accIDs[0], accIDs[1]

	eventID := insertTestEvent(t, "Concert", "concert", evFuture)

	require.NoError(t, DB.RawQuery(`INSERT INTO rsvps (account_id, event_id, state, instrument_id) VALUES (?, ?, 'yes', ?)`, acc1ID, eventID, instrID).Exec())
	require.NoError(t, DB.RawQuery(`INSERT INTO rsvps (account_id, event_id, state) VALUES (?, ?, 'no')`, acc2ID, eventID).Exec())

	require.NoError(t, RSVPStore{}.ClearInstruments(DB, eventID))

	rsvp1, err := RSVPStore{}.GetByAccountAndEvent(DB, acc1ID, eventID)
	require.NoError(t, err)
	assert.False(t, rsvp1.InstrumentID.Valid, "instrument cleared")

	rsvp2, err := RSVPStore{}.GetByAccountAndEvent(DB, acc2ID, eventID)
	require.NoError(t, err)
	assert.False(t, rsvp2.InstrumentID.Valid, "still null")
}

func TestRSVPStore_AddFieldResponse_InsertAndUpsert(t *testing.T) {
	truncateAll(t)
	accountID := insertEventAccount(t)
	eventID := insertTestEvent(t, "Workshop", "other", evFuture)
	fieldID, err := EventStore{}.AddField(DB, eventID, "Q", "text", false, 1)
	require.NoError(t, err)

	rsvpID := insertRSVPRow(t, accountID, eventID, "unanswered")

	store := RSVPStore{}
	require.NoError(t, store.AddFieldResponse(DB, rsvpID, fieldID, "first"))

	var resp struct {
		Value string `db:"value"`
	}
	require.NoError(t, DB.RawQuery(`SELECT value FROM rsvp_field_responses WHERE rsvp_id=? AND event_field_id=?`, rsvpID, fieldID).First(&resp))
	assert.Equal(t, "first", resp.Value)

	require.NoError(t, store.AddFieldResponse(DB, rsvpID, fieldID, "updated"))
	require.NoError(t, DB.RawQuery(`SELECT value FROM rsvp_field_responses WHERE rsvp_id=? AND event_field_id=?`, rsvpID, fieldID).First(&resp))
	assert.Equal(t, "updated", resp.Value)
}

func TestRSVPStore_ListFieldResponses(t *testing.T) {
	truncateAll(t)
	accountID := insertEventAccount(t)
	eventID := insertTestEvent(t, "Workshop", "other", evFuture)

	// field1 at position 2, field2 at position 1 → expect field2 first
	field1ID, err := EventStore{}.AddField(DB, eventID, "LabelB", "text", false, 2)
	require.NoError(t, err)
	field2ID, err := EventStore{}.AddField(DB, eventID, "LabelA", "text", false, 1)
	require.NoError(t, err)

	rsvpID := insertRSVPRow(t, accountID, eventID, "unanswered")

	store := RSVPStore{}
	require.NoError(t, store.AddFieldResponse(DB, rsvpID, field1ID, "ansB"))
	require.NoError(t, store.AddFieldResponse(DB, rsvpID, field2ID, "ansA"))

	rows, err := store.ListFieldResponses(DB, eventID)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, field2ID, rows[0].EventFieldID, "position 1 first")
	assert.Equal(t, "ansA", rows[0].Value)
	assert.Equal(t, field1ID, rows[1].EventFieldID)
	assert.Equal(t, "ansB", rows[1].Value)
}
