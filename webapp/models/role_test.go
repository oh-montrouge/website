package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// insertRoleTestAccount creates an instrument and one active account for role tests.
func insertRoleTestAccount(t *testing.T) int64 {
	t.Helper()
	var instrRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`INSERT INTO instruments (name) VALUES ('Flûte') RETURNING id`).First(&instrRow))
	var accRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(
		`INSERT INTO accounts (email, password_hash, main_instrument_id, status) VALUES ('role@example.com', 'h', ?, 'active') RETURNING id`,
		instrRow.ID,
	).First(&accRow))
	return accRow.ID
}

// insertAdminRole re-inserts the admin role (wiped by truncateAll) and returns its ID.
func insertAdminRole(t *testing.T) int64 {
	t.Helper()
	var row struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`INSERT INTO roles (name) VALUES ('admin') RETURNING id`).First(&row))
	return row.ID
}

func TestAccountRoleStore_HasRole_True(t *testing.T) {
	truncateAll(t)
	accountID := insertRoleTestAccount(t)
	roleID := insertAdminRole(t)
	store := AccountRoleStore{}

	require.NoError(t, store.AssignRole(DB, accountID, roleID))

	has, err := store.HasRole(DB, accountID, "admin")
	require.NoError(t, err)
	assert.True(t, has)
}

func TestAccountRoleStore_HasRole_False(t *testing.T) {
	truncateAll(t)
	accountID := insertRoleTestAccount(t)
	insertAdminRole(t)

	has, err := AccountRoleStore{}.HasRole(DB, accountID, "admin")
	require.NoError(t, err)
	assert.False(t, has)
}

func TestAccountRoleStore_HasActiveRoleHolder_True(t *testing.T) {
	truncateAll(t)
	accountID := insertRoleTestAccount(t)
	roleID := insertAdminRole(t)
	store := AccountRoleStore{}

	require.NoError(t, store.AssignRole(DB, accountID, roleID))

	has, err := store.HasActiveRoleHolder(DB, "admin")
	require.NoError(t, err)
	assert.True(t, has)
}

func TestAccountRoleStore_HasActiveRoleHolder_False(t *testing.T) {
	truncateAll(t)
	insertAdminRole(t)

	has, err := AccountRoleStore{}.HasActiveRoleHolder(DB, "admin")
	require.NoError(t, err)
	assert.False(t, has)
}

func TestAccountRoleStore_GetIDByName(t *testing.T) {
	truncateAll(t)
	roleID := insertAdminRole(t)

	got, err := AccountRoleStore{}.GetIDByName(DB, "admin")
	require.NoError(t, err)
	assert.Equal(t, roleID, got)
}

func TestAccountRoleStore_AssignRole(t *testing.T) {
	truncateAll(t)
	accountID := insertRoleTestAccount(t)
	roleID := insertAdminRole(t)

	require.NoError(t, AccountRoleStore{}.AssignRole(DB, accountID, roleID))

	var cnt struct {
		Count int `db:"count"`
	}
	require.NoError(t, DB.RawQuery(
		`SELECT COUNT(*) AS count FROM account_roles WHERE account_id=? AND role_id=?`,
		accountID, roleID,
	).First(&cnt))
	assert.Equal(t, 1, cnt.Count)
}

func TestAccountRoleStore_CountActiveAdmins(t *testing.T) {
	truncateAll(t)
	accountID := insertRoleTestAccount(t)
	roleID := insertAdminRole(t)
	store := AccountRoleStore{}

	n, err := store.CountActiveAdmins(DB)
	require.NoError(t, err)
	assert.Equal(t, 0, n)

	require.NoError(t, store.AssignRole(DB, accountID, roleID))

	n, err = store.CountActiveAdmins(DB)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
}

func TestAccountRoleStore_RevokeRole(t *testing.T) {
	truncateAll(t)
	accountID := insertRoleTestAccount(t)
	roleID := insertAdminRole(t)
	store := AccountRoleStore{}

	require.NoError(t, store.AssignRole(DB, accountID, roleID))
	require.NoError(t, store.RevokeRole(DB, accountID, roleID))

	has, err := store.HasRole(DB, accountID, "admin")
	require.NoError(t, err)
	assert.False(t, has)
}

func TestAccountRoleStore_RemoveAllRoles(t *testing.T) {
	truncateAll(t)
	accountID := insertRoleTestAccount(t)
	adminRoleID := insertAdminRole(t)

	var role2Row struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(`INSERT INTO roles (name) VALUES ('moderator') RETURNING id`).First(&role2Row))

	store := AccountRoleStore{}
	require.NoError(t, store.AssignRole(DB, accountID, adminRoleID))
	require.NoError(t, store.AssignRole(DB, accountID, role2Row.ID))

	require.NoError(t, store.RemoveAllRoles(DB, accountID))

	has, err := store.HasRole(DB, accountID, "admin")
	require.NoError(t, err)
	assert.False(t, has)

	var cnt struct {
		Count int `db:"count"`
	}
	require.NoError(t, DB.RawQuery(`SELECT COUNT(*) AS count FROM account_roles WHERE account_id=?`, accountID).First(&cnt))
	assert.Equal(t, 0, cnt.Count)
}
