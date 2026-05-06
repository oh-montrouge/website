package models

import (
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountStore_FindByEmail(t *testing.T) {
	truncateAll(t)

	inst := &Instrument{Name: "Clarinette"}
	require.NoError(t, DB.Create(inst))

	acc := &Account{
		Email:            nulls.NewString("alice@example.com"),
		MainInstrumentID: int64(inst.ID),
		Status:           "active",
	}
	require.NoError(t, DB.Create(acc))

	store := AccountStore{}

	t.Run("found", func(t *testing.T) {
		got, err := store.FindByEmail(DB, "alice@example.com")
		assert.NoError(t, err)
		assert.Equal(t, acc.ID, got.ID)
		assert.Equal(t, "alice@example.com", got.Email.String)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := store.FindByEmail(DB, "nobody@example.com")
		assert.Error(t, err)
	})
}

func TestAccountStore_GetByID(t *testing.T) {
	truncateAll(t)

	inst := &Instrument{Name: "Trompette"}
	require.NoError(t, DB.Create(inst))

	acc := &Account{
		Email:            nulls.NewString("bob@example.com"),
		MainInstrumentID: int64(inst.ID),
		Status:           "active",
	}
	require.NoError(t, DB.Create(acc))

	store := AccountStore{}

	t.Run("found", func(t *testing.T) {
		got, err := store.GetByID(DB, acc.ID)
		assert.NoError(t, err)
		assert.Equal(t, acc.ID, got.ID)
		assert.Equal(t, "bob@example.com", got.Email.String)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := store.GetByID(DB, 99999)
		assert.Error(t, err)
	})
}

// jscpd: unwanted abstraction — per-file test helpers share only the RawQuery scaffolding pattern; merging them would couple unrelated test files and obscure per-file SQL differences
// jscpd:ignore-start
// insertAcctInstr creates a minimal instrument + active account for account store tests.
func insertAcctInstr(t *testing.T, email string) (instrID int64, accID int64) {
	t.Helper()
	var instrRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`INSERT INTO instruments (name) VALUES ('Violon') RETURNING id`).First(&instrRow))
	var accRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(
		`INSERT INTO accounts (email, password_hash, main_instrument_id, status, phone_address_consent, processing_restricted)
		 VALUES (?, 'testhash', ?, 'active', false, false) RETURNING id`,
		email, instrRow.ID,
	).First(&accRow))
	return instrRow.ID, accRow.ID
}

// jscpd:ignore-end

func TestAccountStore_Create(t *testing.T) {
	truncateAll(t)
	var instrRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`INSERT INTO instruments (name) VALUES ('Piano') RETURNING id`).First(&instrRow))

	store := AccountStore{}
	id, err := store.Create(DB, "create@example.com", "securehash", instrRow.ID)
	require.NoError(t, err)
	assert.Positive(t, id)

	got, err := store.GetByID(DB, id)
	require.NoError(t, err)
	assert.Equal(t, "create@example.com", got.Email.String)
	assert.Equal(t, "active", got.Status)
}

func TestAccountStore_UpdatePasswordHash(t *testing.T) {
	truncateAll(t)
	_, accID := insertAcctInstr(t, "pwtest@example.com")

	store := AccountStore{}
	require.NoError(t, store.UpdatePasswordHash(DB, accID, "newhash"))

	got, err := store.GetByID(DB, accID)
	require.NoError(t, err)
	assert.Equal(t, "newhash", got.PasswordHash.String)
}

func TestAccountStore_Activate_WithConsent(t *testing.T) {
	truncateAll(t)
	var instrRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`INSERT INTO instruments (name) VALUES ('Hautbois') RETURNING id`).First(&instrRow))
	var accRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`
		INSERT INTO accounts (email, main_instrument_id, status, phone, address, phone_address_consent, processing_restricted)
		VALUES ('pending@example.com', ?, 'pending', '555-1234', '1 Rue de la Paix', false, false) RETURNING id`,
		instrRow.ID,
	).First(&accRow))

	store := AccountStore{}
	require.NoError(t, store.Activate(DB, accRow.ID, "activehash", true))

	got, err := store.GetByID(DB, accRow.ID)
	require.NoError(t, err)
	assert.Equal(t, "active", got.Status)
	assert.Equal(t, "activehash", got.PasswordHash.String)
	assert.True(t, got.PhoneAddressConsent)
	assert.Equal(t, "555-1234", got.Phone.String)
	assert.Equal(t, "1 Rue de la Paix", got.Address.String)
}

func TestAccountStore_Activate_WithoutConsent(t *testing.T) {
	truncateAll(t)
	var instrRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`INSERT INTO instruments (name) VALUES ('Clarinette basse') RETURNING id`).First(&instrRow))
	var accRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`
		INSERT INTO accounts (email, main_instrument_id, status, phone, address, phone_address_consent, processing_restricted)
		VALUES ('pending2@example.com', ?, 'pending', '555-5678', '2 Avenue Kléber', true, false) RETURNING id`,
		instrRow.ID,
	).First(&accRow))

	store := AccountStore{}
	require.NoError(t, store.Activate(DB, accRow.ID, "activehash2", false))

	got, err := store.GetByID(DB, accRow.ID)
	require.NoError(t, err)
	assert.Equal(t, "active", got.Status)
	assert.False(t, got.PhoneAddressConsent)
	assert.False(t, got.Phone.Valid)
	assert.False(t, got.Address.Valid)
}

func TestAccountStore_CreatePending(t *testing.T) {
	truncateAll(t)
	var instrRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`INSERT INTO instruments (name) VALUES ('Cor') RETURNING id`).First(&instrRow))

	store := AccountStore{}
	id, err := store.CreatePending(DB, "pend@example.com", instrRow.ID)
	require.NoError(t, err)
	assert.Positive(t, id)

	got, err := store.GetByID(DB, id)
	require.NoError(t, err)
	assert.Equal(t, "pending", got.Status)
	assert.False(t, got.PhoneAddressConsent)
	assert.False(t, got.ProcessingRestricted)
}

func TestAccountStore_UpdateEmail(t *testing.T) {
	truncateAll(t)
	_, accID := insertAcctInstr(t, "old@example.com")

	store := AccountStore{}
	require.NoError(t, store.UpdateEmail(DB, accID, "new@example.com"))

	got, err := store.GetByID(DB, accID)
	require.NoError(t, err)
	assert.Equal(t, "new@example.com", got.Email.String)
}

func TestAccountStore_Delete(t *testing.T) {
	truncateAll(t)
	_, accID := insertAcctInstr(t, "delete@example.com")

	store := AccountStore{}
	require.NoError(t, store.Delete(DB, accID))

	_, err := store.GetByID(DB, accID)
	assert.Error(t, err)
}

func TestAccountStore_AnonymizeAccount(t *testing.T) {
	truncateAll(t)
	_, accID := insertAcctInstr(t, "anon@example.com")

	store := AccountStore{}
	require.NoError(t, store.AnonymizeAccount(DB, accID, "abc12345"))

	got, err := store.GetByID(DB, accID)
	require.NoError(t, err)
	assert.Equal(t, "anonymized", got.Status)
	assert.False(t, got.Email.Valid)
	assert.False(t, got.PasswordHash.Valid)
	assert.Equal(t, "abc12345", got.AnonymizationToken.String)
}

func TestAccountStore_GetProfile(t *testing.T) {
	truncateAll(t)
	var instrRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`INSERT INTO instruments (name) VALUES ('Alto') RETURNING id`).First(&instrRow))
	var accRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`
		INSERT INTO accounts (email, password_hash, main_instrument_id, status, first_name, last_name, phone_address_consent, processing_restricted)
		VALUES ('profile@example.com', 'h', ?, 'active', 'Alice', 'Dupont', false, false) RETURNING id`,
		instrRow.ID,
	).First(&accRow))

	got, err := AccountStore{}.GetProfile(DB, accRow.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, accRow.ID, got.ID)
	assert.Equal(t, instrRow.ID, got.MainInstrumentID)
	assert.Equal(t, "Alto", got.InstrumentName)
	assert.Equal(t, "Alice", got.FirstName.String)
	assert.Equal(t, "Dupont", got.LastName.String)
	assert.False(t, got.BirthDate.Valid)
	assert.False(t, got.PhoneAddressConsent)
	assert.False(t, got.ProcessingRestricted)
}

func TestAccountStore_SetProfile(t *testing.T) {
	truncateAll(t)
	_, accID := insertAcctInstr(t, "setprofile@example.com")

	birthDate := time.Date(1985, 7, 4, 0, 0, 0, 0, time.UTC)
	store := AccountStore{}
	require.NoError(t, store.SetProfile(DB, accID, "Marie", "Martin", &birthDate, "uri://parental"))

	got, err := store.GetByID(DB, accID)
	require.NoError(t, err)
	assert.Equal(t, "Marie", got.FirstName.String)
	assert.Equal(t, "Martin", got.LastName.String)
	assert.True(t, got.BirthDate.Valid)
	assert.Equal(t, birthDate.Format("2006-01-02"), got.BirthDate.Time.Format("2006-01-02"))
	assert.Equal(t, "uri://parental", got.ParentalConsentURI.String)
}

func TestAccountStore_UpdateProfile(t *testing.T) {
	truncateAll(t)
	_, accID := insertAcctInstr(t, "updateprofile@example.com")
	var instr2Row struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`INSERT INTO instruments (name) VALUES ('Tuba') RETURNING id`).First(&instr2Row))

	birthDate := time.Date(1995, 11, 11, 0, 0, 0, 0, time.UTC)
	store := AccountStore{}
	require.NoError(t, store.UpdateProfile(DB, accID,
		"Jean", "Leblanc", "updated@example.com",
		instr2Row.ID, &birthDate, "uri://updated",
		"06-12-34-56-78", "99 Bd Haussmann",
	))

	got, err := store.GetByID(DB, accID)
	require.NoError(t, err)
	assert.Equal(t, "Jean", got.FirstName.String)
	assert.Equal(t, "Leblanc", got.LastName.String)
	assert.Equal(t, "updated@example.com", got.Email.String)
	assert.Equal(t, instr2Row.ID, got.MainInstrumentID)
	assert.True(t, got.BirthDate.Valid)
	assert.Equal(t, birthDate.Format("2006-01-02"), got.BirthDate.Time.Format("2006-01-02"))
	assert.Equal(t, "uri://updated", got.ParentalConsentURI.String)
	assert.Equal(t, "06-12-34-56-78", got.Phone.String)
	assert.Equal(t, "99 Bd Haussmann", got.Address.String)
}

// TestAccountStore_ListNonAnonymized verifies is_admin flag and that all accounts are
// returned — the SQL has no WHERE filter despite the function name.
func TestAccountStore_ListNonAnonymized(t *testing.T) {
	truncateAll(t)
	var instrRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`INSERT INTO instruments (name) VALUES ('Basson') RETURNING id`).First(&instrRow))

	var activeRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`
		INSERT INTO accounts (email, password_hash, main_instrument_id, status, phone_address_consent, processing_restricted)
		VALUES ('active@example.com', 'h', ?, 'active', false, false) RETURNING id`,
		instrRow.ID,
	).First(&activeRow))

	require.NoError(t, DB.RawQuery(`
		INSERT INTO accounts (main_instrument_id, status, phone_address_consent, processing_restricted)
		VALUES (?, 'anonymized', false, false)`,
		instrRow.ID,
	).Exec())

	store := AccountStore{}
	rows, err := store.ListNonAnonymized(DB)
	require.NoError(t, err)
	assert.Len(t, rows, 2)

	var activeResult *MusicianListRow
	for i := range rows {
		if rows[i].ID == activeRow.ID {
			activeResult = &rows[i]
		}
	}
	require.NotNil(t, activeResult)
	assert.False(t, activeResult.IsAdmin)

	roleID := insertAdminRole(t)
	require.NoError(t, DB.RawQuery(
		`INSERT INTO account_roles (account_id, role_id) VALUES (?, ?)`, activeRow.ID, roleID,
	).Exec())

	rows, err = store.ListNonAnonymized(DB)
	require.NoError(t, err)
	for i := range rows {
		if rows[i].ID == activeRow.ID {
			assert.True(t, rows[i].IsAdmin)
		}
	}
}

func TestAccountStore_ListForRetentionReview(t *testing.T) {
	truncateAll(t)
	var instrRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`INSERT INTO instruments (name) VALUES ('Contrebasse') RETURNING id`).First(&instrRow))

	// Account A: last payment in season that ended >5 years ago → must appear.
	var accARow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`
		INSERT INTO accounts (first_name, last_name, email, password_hash, main_instrument_id, status, phone_address_consent, processing_restricted)
		VALUES ('Jean', 'Vieux', 'old@example.com', 'h', ?, 'active', false, false) RETURNING id`,
		instrRow.ID,
	).First(&accARow))
	var oldSeasonRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(
		`INSERT INTO seasons (label, start_date, end_date, is_current) VALUES ('2014-2015', '2014-09-01', '2015-06-30', false) RETURNING id`,
	).First(&oldSeasonRow))
	_, err := FeePaymentStore{}.Create(DB, accARow.ID, oldSeasonRow.ID, 50.00, time.Date(2015, 1, 15, 0, 0, 0, 0, time.UTC), "chèque", "")
	require.NoError(t, err)

	// Account B: last payment in recent season → must NOT appear.
	var accBRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`
		INSERT INTO accounts (first_name, last_name, email, password_hash, main_instrument_id, status, phone_address_consent, processing_restricted)
		VALUES ('Paul', 'Recent', 'recent@example.com', 'h', ?, 'active', false, false) RETURNING id`,
		instrRow.ID,
	).First(&accBRow))
	var recentSeasonRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(
		`INSERT INTO seasons (label, start_date, end_date, is_current) VALUES ('2024-2025', '2024-09-01', '2025-06-30', false) RETURNING id`,
	).First(&recentSeasonRow))
	_, err = FeePaymentStore{}.Create(DB, accBRow.ID, recentSeasonRow.ID, 50.00, time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), "chèque", "")
	require.NoError(t, err)

	store := AccountStore{}
	rows, err := store.ListForRetentionReview(DB)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, accARow.ID, rows[0].ID)
	assert.Equal(t, "2014-2015", rows[0].LastSeasonLabel)
}

func TestAccountStore_ClearMembershipFields(t *testing.T) {
	truncateAll(t)
	var instrRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`INSERT INTO instruments (name) VALUES ('Piccolo') RETURNING id`).First(&instrRow))
	var accRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`
		INSERT INTO accounts (first_name, last_name, email, password_hash, main_instrument_id, status,
		    birth_date, parental_consent_uri, phone, address, phone_address_consent, processing_restricted)
		VALUES ('Test', 'User', 'clear@example.com', 'h', ?, 'active',
		    '2000-01-01', 'uri://consent', '555-0000', '1 Test St', true, true) RETURNING id`,
		instrRow.ID,
	).First(&accRow))

	store := AccountStore{}
	require.NoError(t, store.ClearMembershipFields(DB, accRow.ID))

	got, err := store.GetByID(DB, accRow.ID)
	require.NoError(t, err)
	assert.False(t, got.FirstName.Valid)
	assert.False(t, got.LastName.Valid)
	assert.False(t, got.BirthDate.Valid)
	assert.False(t, got.ParentalConsentURI.Valid)
	assert.False(t, got.Phone.Valid)
	assert.False(t, got.Address.Valid)
	assert.False(t, got.PhoneAddressConsent)
	assert.False(t, got.ProcessingRestricted)
}

func TestAccountStore_WithdrawConsent(t *testing.T) {
	truncateAll(t)
	var instrRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`INSERT INTO instruments (name) VALUES ('Trombone') RETURNING id`).First(&instrRow))
	var accRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`
		INSERT INTO accounts (email, password_hash, main_instrument_id, status, phone, address, phone_address_consent, processing_restricted)
		VALUES ('consent@example.com', 'h', ?, 'active', '555-1111', '2 Test Ave', true, false) RETURNING id`,
		instrRow.ID,
	).First(&accRow))

	store := AccountStore{}
	require.NoError(t, store.WithdrawConsent(DB, accRow.ID))

	got, err := store.GetByID(DB, accRow.ID)
	require.NoError(t, err)
	assert.False(t, got.Phone.Valid)
	assert.False(t, got.Address.Valid)
	assert.False(t, got.PhoneAddressConsent)
}

func TestAccountStore_ToggleProcessingRestriction(t *testing.T) {
	truncateAll(t)
	_, accID := insertAcctInstr(t, "toggle@example.com")

	store := AccountStore{}

	require.NoError(t, store.ToggleProcessingRestriction(DB, accID))
	got, err := store.GetByID(DB, accID)
	require.NoError(t, err)
	assert.True(t, got.ProcessingRestricted)

	require.NoError(t, store.ToggleProcessingRestriction(DB, accID))
	got, err = store.GetByID(DB, accID)
	require.NoError(t, err)
	assert.False(t, got.ProcessingRestricted)
}
