package models

import (
	"time"

	"github.com/gobuffalo/pop/v6"
)

// PasswordResetToken is the DB model for the password_reset_tokens table.
type PasswordResetToken struct {
	ID        int64     `db:"id"`
	AccountID int64     `db:"account_id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
	Used      bool      `db:"used"`
}

// PasswordResetTokenRecord is the result of the FindByToken JOIN query.
type PasswordResetTokenRecord struct {
	TokenID       int64  `db:"token_id"`
	AccountID     int64  `db:"account_id"`
	AccountStatus string `db:"account_status"`
}

// PasswordResetTokenStore is the production implementation of services.PasswordResetTokenRepository.
type PasswordResetTokenStore struct{}

func (PasswordResetTokenStore) Generate(tx *pop.Connection, accountID int64, token string, expiresAt time.Time) error {
	return tx.RawQuery(
		`INSERT INTO password_reset_tokens (account_id, token, expires_at, used) VALUES (?, ?, ?, false)`,
		accountID, token, expiresAt,
	).Exec()
}

// FindByToken returns the token record joined with account status.
// Returns nil, nil when the token is not found, expired, used, or the account is not active.
func (PasswordResetTokenStore) FindByToken(tx *pop.Connection, token string) (*PasswordResetTokenRecord, error) {
	var row PasswordResetTokenRecord
	err := tx.RawQuery(`
		SELECT
			t.id        AS token_id,
			t.account_id AS account_id,
			a.status    AS account_status
		FROM password_reset_tokens t
		JOIN accounts a ON a.id = t.account_id
		WHERE t.token = ?
		  AND t.used = false
		  AND t.expires_at > NOW()
		  AND a.status = 'active'`,
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

func (PasswordResetTokenStore) MarkUsed(tx *pop.Connection, tokenID int64) error {
	return tx.RawQuery(
		`UPDATE password_reset_tokens SET used = true WHERE id = ?`, tokenID,
	).Exec()
}

func (PasswordResetTokenStore) InvalidateExisting(tx *pop.Connection, accountID int64) error {
	return tx.RawQuery(
		`UPDATE password_reset_tokens SET used = true WHERE account_id = ? AND used = false`, accountID,
	).Exec()
}

// FindActiveForAccount returns the current active (unused, non-expired) reset token for an account, or nil.
func (PasswordResetTokenStore) FindActiveForAccount(tx *pop.Connection, accountID int64) (*PasswordResetToken, error) {
	var tok PasswordResetToken
	err := tx.RawQuery(`
		SELECT id, account_id, token, expires_at, used
		FROM password_reset_tokens
		WHERE account_id = ? AND used = false AND expires_at > NOW()
		ORDER BY expires_at DESC
		LIMIT 1`, accountID,
	).First(&tok)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &tok, nil
}
