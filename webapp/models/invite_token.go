package models

import (
	"time"

	"github.com/gobuffalo/pop/v6"
)

// InviteToken is the DB model for the invite_tokens table.
type InviteToken struct {
	ID        int64     `db:"id"`
	AccountID int64     `db:"account_id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
	Used      bool      `db:"used"`
}

// InviteTokenRecord is the result of the FindByToken JOIN query.
// Carries account and instrument fields needed to render the invite form.
type InviteTokenRecord struct {
	TokenID        int64  `db:"token_id"`
	AccountID      int64  `db:"account_id"`
	FirstName      string `db:"first_name"`
	LastName       string `db:"last_name"`
	Email          string `db:"email"`
	InstrumentName string `db:"instrument_name"`
}

// InviteTokenStore is the production implementation of services.InviteTokenRepository.
type InviteTokenStore struct{}

func (InviteTokenStore) Generate(tx *pop.Connection, accountID int64, token string, expiresAt time.Time) error {
	return tx.RawQuery(
		`INSERT INTO invite_tokens (account_id, token, expires_at, used) VALUES (?, ?, ?, false)`,
		accountID, token, expiresAt,
	).Exec()
}

// FindByToken returns the token record joined with account and instrument data.
// Returns nil, nil when the token is not found, expired, used, or the account status is not pending.
func (InviteTokenStore) FindByToken(tx *pop.Connection, token string) (*InviteTokenRecord, error) {
	var row InviteTokenRecord
	err := tx.RawQuery(`
		SELECT
			t.id           AS token_id,
			t.account_id   AS account_id,
			a.first_name   AS first_name,
			a.last_name    AS last_name,
			a.email        AS email,
			i.name         AS instrument_name
		FROM invite_tokens t
		JOIN accounts a    ON a.id = t.account_id
		JOIN instruments i ON i.id = a.main_instrument_id
		WHERE t.token = ?
		  AND t.used = false
		  AND t.expires_at > NOW()
		  AND a.status = 'pending'`,
		token,
	).First(&row)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (InviteTokenStore) MarkUsed(tx *pop.Connection, tokenID int64) error {
	return tx.RawQuery(
		`UPDATE invite_tokens SET used = true WHERE id = ?`, tokenID,
	).Exec()
}

func (InviteTokenStore) InvalidateExisting(tx *pop.Connection, accountID int64) error {
	return tx.RawQuery(
		`UPDATE invite_tokens SET used = true WHERE account_id = ? AND used = false`, accountID,
	).Exec()
}
