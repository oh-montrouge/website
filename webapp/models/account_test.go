package models

import (
	"testing"

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
