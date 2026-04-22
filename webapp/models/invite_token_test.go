package models

import (
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedPendingAccount(t *testing.T) (accountID int64, instrumentID int64) {
	t.Helper()
	inst := &Instrument{Name: "Clarinette"}
	require.NoError(t, DB.Create(inst))
	acc := &Account{
		FirstName:        nulls.NewString("Alice"),
		LastName:         nulls.NewString("Dupont"),
		Email:            nulls.NewString("alice@example.com"),
		MainInstrumentID: int64(inst.ID),
		Status:           "pending",
	}
	require.NoError(t, DB.Create(acc))
	return acc.ID, int64(inst.ID)
}

func TestInviteTokenStore_Generate(t *testing.T) {
	truncateAll(t)
	accountID, _ := seedPendingAccount(t)
	store := InviteTokenStore{}
	err := store.Generate(DB, accountID, "tok123", time.Now().Add(7*24*time.Hour))
	assert.NoError(t, err)
}

func TestInviteTokenStore_FindByToken_Valid(t *testing.T) {
	truncateAll(t)
	accountID, _ := seedPendingAccount(t)
	store := InviteTokenStore{}
	require.NoError(t, store.Generate(DB, accountID, "validtoken", time.Now().Add(7*24*time.Hour)))

	row, err := store.FindByToken(DB, "validtoken")
	assert.NoError(t, err)
	require.NotNil(t, row)
	assert.Equal(t, accountID, row.AccountID)
	assert.Equal(t, "Alice", row.FirstName)
	assert.Equal(t, "Dupont", row.LastName)
	assert.Equal(t, "alice@example.com", row.Email)
	assert.Equal(t, "Clarinette", row.InstrumentName)
}

func TestInviteTokenStore_FindByToken_NotFound(t *testing.T) {
	truncateAll(t)
	row, err := InviteTokenStore{}.FindByToken(DB, "no-such-token")
	assert.NoError(t, err)
	assert.Nil(t, row)
}

func TestInviteTokenStore_FindByToken_Expired(t *testing.T) {
	truncateAll(t)
	accountID, _ := seedPendingAccount(t)
	store := InviteTokenStore{}
	require.NoError(t, store.Generate(DB, accountID, "expiredtok", time.Now().Add(-1*time.Hour)))

	row, err := store.FindByToken(DB, "expiredtok")
	assert.NoError(t, err)
	assert.Nil(t, row)
}

func TestInviteTokenStore_FindByToken_Used(t *testing.T) {
	truncateAll(t)
	accountID, _ := seedPendingAccount(t)
	store := InviteTokenStore{}
	require.NoError(t, store.Generate(DB, accountID, "usedtok", time.Now().Add(7*24*time.Hour)))
	require.NoError(t, DB.RawQuery(`UPDATE invite_tokens SET used = true WHERE token = 'usedtok'`).Exec())

	row, err := store.FindByToken(DB, "usedtok")
	assert.NoError(t, err)
	assert.Nil(t, row)
}

func TestInviteTokenStore_MarkUsed(t *testing.T) {
	truncateAll(t)
	accountID, _ := seedPendingAccount(t)
	store := InviteTokenStore{}
	require.NoError(t, store.Generate(DB, accountID, "marktok", time.Now().Add(7*24*time.Hour)))

	var tok InviteToken
	require.NoError(t, DB.RawQuery(`SELECT * FROM invite_tokens WHERE token = 'marktok'`).First(&tok))

	require.NoError(t, store.MarkUsed(DB, tok.ID))

	row, err := store.FindByToken(DB, "marktok")
	assert.NoError(t, err)
	assert.Nil(t, row)
}

func TestInviteTokenStore_InvalidateExisting(t *testing.T) {
	truncateAll(t)
	accountID, _ := seedPendingAccount(t)
	store := InviteTokenStore{}
	require.NoError(t, store.Generate(DB, accountID, "old-tok", time.Now().Add(7*24*time.Hour)))

	require.NoError(t, store.InvalidateExisting(DB, accountID))

	row, err := store.FindByToken(DB, "old-tok")
	assert.NoError(t, err)
	assert.Nil(t, row)
}

func TestAccountStore_Activate_ConsentFalse_ClearsPhoneAddress(t *testing.T) {
	truncateAll(t)
	inst := &Instrument{Name: "Trompette"}
	require.NoError(t, DB.Create(inst))
	acc := &Account{
		FirstName:        nulls.NewString("Bob"),
		LastName:         nulls.NewString("Martin"),
		Email:            nulls.NewString("bob@example.com"),
		Phone:            nulls.NewString("0600000000"),
		Address:          nulls.NewString("1 rue de la Paix"),
		MainInstrumentID: int64(inst.ID),
		Status:           "pending",
	}
	require.NoError(t, DB.Create(acc))

	require.NoError(t, AccountStore{}.Activate(DB, acc.ID, "hashed", false))

	var updated Account
	require.NoError(t, DB.Find(&updated, acc.ID))
	assert.Equal(t, "active", updated.Status)
	assert.Equal(t, "hashed", updated.PasswordHash.String)
	assert.False(t, updated.PhoneAddressConsent)
	assert.False(t, updated.Phone.Valid)
	assert.False(t, updated.Address.Valid)
}

func TestAccountStore_Activate_ConsentTrue_PreservesPhoneAddress(t *testing.T) {
	truncateAll(t)
	inst := &Instrument{Name: "Flûte"}
	require.NoError(t, DB.Create(inst))
	acc := &Account{
		FirstName:        nulls.NewString("Carol"),
		LastName:         nulls.NewString("Blanc"),
		Email:            nulls.NewString("carol@example.com"),
		Phone:            nulls.NewString("0611223344"),
		Address:          nulls.NewString("2 avenue des Arts"),
		MainInstrumentID: int64(inst.ID),
		Status:           "pending",
	}
	require.NoError(t, DB.Create(acc))

	require.NoError(t, AccountStore{}.Activate(DB, acc.ID, "hashed2", true))

	var updated Account
	require.NoError(t, DB.Find(&updated, acc.ID))
	assert.Equal(t, "active", updated.Status)
	assert.True(t, updated.PhoneAddressConsent)
	assert.Equal(t, "0611223344", updated.Phone.String)
	assert.Equal(t, "2 avenue des Arts", updated.Address.String)
}
