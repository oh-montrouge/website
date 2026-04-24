package models

import "github.com/gobuffalo/pop/v6"

// HTTPSessionStore is the production implementation of services.SessionRepository.
type HTTPSessionStore struct{}

func (HTTPSessionStore) BindAccount(db *pop.Connection, sessionKey string, accountID int64) error {
	return db.RawQuery(
		"UPDATE http_sessions SET account_id = ? WHERE key = ?",
		accountID, sessionKey,
	).Exec()
}

func (HTTPSessionStore) DeleteByAccount(db *pop.Connection, accountID int64) error {
	return db.RawQuery(
		"DELETE FROM http_sessions WHERE account_id = ?", accountID,
	).Exec()
}
