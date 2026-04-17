package models

import (
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
)

// Account represents an OHM member account.
type Account struct {
	ID                   int64        `db:"id"`
	FirstName            nulls.String `db:"first_name"`
	LastName             nulls.String `db:"last_name"`
	Email                nulls.String `db:"email"`
	PasswordHash         nulls.String `db:"password_hash"`
	MainInstrumentID     int64        `db:"main_instrument_id"`
	BirthDate            nulls.Time   `db:"birth_date"`
	ParentalConsentURI   nulls.String `db:"parental_consent_uri"`
	Phone                nulls.String `db:"phone"`
	Address              nulls.String `db:"address"`
	PhoneAddressConsent  bool         `db:"phone_address_consent"`
	Status               string       `db:"status"`
	ProcessingRestricted bool         `db:"processing_restricted"`
	AnonymizationToken   nulls.String `db:"anonymization_token"`
}

// Accounts is a slice of Account.
type Accounts []Account

// AccountStore is the production implementation of services.AccountRepository.
type AccountStore struct{}

func (AccountStore) FindByEmail(tx *pop.Connection, email string) (*Account, error) {
	account := &Account{}
	err := tx.Where("email = ?", email).First(account)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (AccountStore) GetByID(tx *pop.Connection, id int64) (*Account, error) {
	account := &Account{}
	err := tx.Find(account, id)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (AccountStore) Create(tx *pop.Connection, email, passwordHash string, instrumentID int64) (int64, error) {
	var row struct {
		ID int64 `db:"id"`
	}
	err := tx.RawQuery(
		`INSERT INTO accounts (email, password_hash, main_instrument_id, status) VALUES (?, ?, ?, 'active') RETURNING id`,
		email, passwordHash, instrumentID,
	).First(&row)
	return row.ID, err
}

func (AccountStore) UpdatePasswordHash(tx *pop.Connection, id int64, hash string) error {
	return tx.RawQuery(
		`UPDATE accounts SET password_hash = ? WHERE id = ?`,
		hash, id,
	).Exec()
}
