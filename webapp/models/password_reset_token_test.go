package models

import (
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedActiveAccount(t *testing.T) int64 {
	t.Helper()
	inst := &Instrument{Name: "Hautbois"}
	require.NoError(t, DB.Create(inst))
	acc := &Account{
		Email:            nulls.NewString("active@example.com"),
		PasswordHash:     nulls.NewString("somehash"),
		MainInstrumentID: int64(inst.ID),
		Status:           "active",
	}
	require.NoError(t, DB.Create(acc))
	return acc.ID
}

func TestPasswordResetTokenStore_Generate(t *testing.T) {
	truncateAll(t)
	accountID := seedActiveAccount(t)
	err := PasswordResetTokenStore{}.Generate(DB, accountID, "resettok", time.Now().Add(7*24*time.Hour))
	assert.NoError(t, err)
}

func TestPasswordResetTokenStore_FindByToken_Valid(t *testing.T) {
	truncateAll(t)
	accountID := seedActiveAccount(t)
	store := PasswordResetTokenStore{}
	require.NoError(t, store.Generate(DB, accountID, "validreset", time.Now().Add(7*24*time.Hour)))

	row, err := store.FindByToken(DB, "validreset")
	assert.NoError(t, err)
	require.NotNil(t, row)
	assert.Equal(t, accountID, row.AccountID)
	assert.Equal(t, "active", row.AccountStatus)
}

func TestPasswordResetTokenStore_FindByToken_NotFound(t *testing.T) {
	truncateAll(t)
	row, err := PasswordResetTokenStore{}.FindByToken(DB, "no-such-token")
	assert.NoError(t, err)
	assert.Nil(t, row)
}

func TestPasswordResetTokenStore_FindByToken_Expired(t *testing.T) {
	truncateAll(t)
	accountID := seedActiveAccount(t)
	store := PasswordResetTokenStore{}
	require.NoError(t, store.Generate(DB, accountID, "expiredreset", time.Now().Add(-1*time.Hour)))

	row, err := store.FindByToken(DB, "expiredreset")
	assert.NoError(t, err)
	assert.Nil(t, row)
}

func TestPasswordResetTokenStore_FindByToken_PendingAccountReturnsNil(t *testing.T) {
	truncateAll(t)
	inst := &Instrument{Name: "Cor"}
	require.NoError(t, DB.Create(inst))
	acc := &Account{
		Email:            nulls.NewString("pending2@example.com"),
		MainInstrumentID: int64(inst.ID),
		Status:           "pending",
	}
	require.NoError(t, DB.Create(acc))
	store := PasswordResetTokenStore{}
	require.NoError(t, store.Generate(DB, acc.ID, "pendingreset", time.Now().Add(7*24*time.Hour)))

	row, err := store.FindByToken(DB, "pendingreset")
	assert.NoError(t, err)
	assert.Nil(t, row)
}

func TestPasswordResetTokenStore_MarkUsed(t *testing.T) {
	truncateAll(t)
	accountID := seedActiveAccount(t)
	store := PasswordResetTokenStore{}
	require.NoError(t, store.Generate(DB, accountID, "markreset", time.Now().Add(7*24*time.Hour)))

	var tok PasswordResetToken
	require.NoError(t, DB.RawQuery(`SELECT * FROM password_reset_tokens WHERE token = 'markreset'`).First(&tok))

	require.NoError(t, store.MarkUsed(DB, tok.ID))

	row, err := store.FindByToken(DB, "markreset")
	assert.NoError(t, err)
	assert.Nil(t, row)
}

func TestPasswordResetTokenStore_InvalidateExisting(t *testing.T) {
	truncateAll(t)
	accountID := seedActiveAccount(t)
	store := PasswordResetTokenStore{}
	require.NoError(t, store.Generate(DB, accountID, "old-reset", time.Now().Add(7*24*time.Hour)))

	require.NoError(t, store.InvalidateExisting(DB, accountID))

	row, err := store.FindByToken(DB, "old-reset")
	assert.NoError(t, err)
	assert.Nil(t, row)
}
