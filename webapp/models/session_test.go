package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func insertHTTPSession(t *testing.T, key string) {
	t.Helper()
	now := time.Now().UTC()
	require.NoError(t, DB.RawQuery(
		`INSERT INTO http_sessions (key, data, created_on, modified_on, expires_on) VALUES (?, ?, ?, ?, ?)`,
		key, []byte{0}, now, now, now.Add(24*time.Hour),
	).Exec())
}

func TestHTTPSessionStore_BindAccount(t *testing.T) {
	truncateAll(t)
	accountID := insertEventAccount(t)
	insertHTTPSession(t, "sess-bind")

	require.NoError(t, HTTPSessionStore{}.BindAccount(DB, "sess-bind", accountID))

	var row struct {
		AccountID *int64 `db:"account_id"`
	}
	require.NoError(t, DB.RawQuery(`SELECT account_id FROM http_sessions WHERE key='sess-bind'`).First(&row))
	require.NotNil(t, row.AccountID)
	assert.Equal(t, accountID, *row.AccountID)
}

func TestHTTPSessionStore_DeleteByAccount(t *testing.T) {
	truncateAll(t)
	accountID := insertEventAccount(t)
	insertHTTPSession(t, "sess-del")

	require.NoError(t, HTTPSessionStore{}.BindAccount(DB, "sess-del", accountID))
	require.NoError(t, HTTPSessionStore{}.DeleteByAccount(DB, accountID))

	var cnt struct {
		Count int `db:"count"`
	}
	require.NoError(t, DB.RawQuery(`SELECT COUNT(*) AS count FROM http_sessions WHERE account_id=?`, accountID).First(&cnt))
	assert.Equal(t, 0, cnt.Count)
}
