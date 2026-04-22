package models

import (
	"time"

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

// Activate transitions a pending account to active in a single UPDATE.
// When phoneAddressConsent is false, phone and address are cleared.
func (AccountStore) Activate(tx *pop.Connection, id int64, passwordHash string, phoneAddressConsent bool) error {
	return tx.RawQuery(
		`UPDATE accounts
		 SET status = 'active',
		     password_hash = ?,
		     phone_address_consent = ?,
		     phone    = CASE WHEN ? THEN phone    ELSE NULL END,
		     address  = CASE WHEN ? THEN address  ELSE NULL END
		 WHERE id = ?`,
		passwordHash, phoneAddressConsent, phoneAddressConsent, phoneAddressConsent, id,
	).Exec()
}

func (AccountStore) CreatePending(tx *pop.Connection, email string, instrumentID int64) (int64, error) {
	var row struct {
		ID int64 `db:"id"`
	}
	err := tx.RawQuery(
		`INSERT INTO accounts (email, main_instrument_id, status, phone_address_consent, processing_restricted)
		 VALUES (?, ?, 'pending', false, false) RETURNING id`,
		email, instrumentID,
	).First(&row)
	return row.ID, err
}

func (AccountStore) UpdateEmail(tx *pop.Connection, id int64, email string) error {
	return tx.RawQuery(
		`UPDATE accounts SET email = ? WHERE id = ?`, email, id,
	).Exec()
}

func (AccountStore) Delete(tx *pop.Connection, id int64) error {
	return tx.RawQuery(`DELETE FROM accounts WHERE id = ?`, id).Exec()
}

func (AccountStore) AnonymizeAccount(tx *pop.Connection, id int64, token string) error {
	return tx.RawQuery(`
		UPDATE accounts
		SET email                  = NULL,
		    password_hash          = NULL,
		    status                 = 'anonymized',
		    anonymization_token    = ?
		WHERE id = ?`,
		token, id,
	).Exec()
}

// MusicianProfileRow is the result of a GetProfile JOIN query.
type MusicianProfileRow struct {
	ID                   int64        `db:"id"`
	FirstName            nulls.String `db:"first_name"`
	LastName             nulls.String `db:"last_name"`
	MainInstrumentID     int64        `db:"main_instrument_id"`
	InstrumentName       string       `db:"instrument_name"`
	BirthDate            nulls.Time   `db:"birth_date"`
	ParentalConsentURI   nulls.String `db:"parental_consent_uri"`
	Phone                nulls.String `db:"phone"`
	Address              nulls.String `db:"address"`
	PhoneAddressConsent  bool         `db:"phone_address_consent"`
	ProcessingRestricted bool         `db:"processing_restricted"`
}

// MusicianListRow is the result of a ListNonAnonymized JOIN query.
type MusicianListRow struct {
	ID                 int64        `db:"id"`
	FirstName          nulls.String `db:"first_name"`
	LastName           nulls.String `db:"last_name"`
	AnonymizationToken nulls.String `db:"anonymization_token"`
	InstrumentName     string       `db:"instrument_name"`
	Status             string       `db:"status"`
	IsAdmin            bool         `db:"is_admin"`
}

// RetentionRow is the result of a ListForRetentionReview query.
type RetentionRow struct {
	ID                int64     `db:"id"`
	FirstName         string    `db:"first_name"`
	LastName          string    `db:"last_name"`
	InstrumentName    string    `db:"instrument_name"`
	LastSeasonLabel   string    `db:"last_season_label"`
	LastSeasonEndDate time.Time `db:"end_date"`
}

func (AccountStore) GetProfile(tx *pop.Connection, accountID int64) (*MusicianProfileRow, error) {
	var row MusicianProfileRow
	err := tx.RawQuery(`
		SELECT
		    a.id,
		    a.first_name,
		    a.last_name,
		    a.main_instrument_id,
		    i.name AS instrument_name,
		    a.birth_date,
		    a.parental_consent_uri,
		    a.phone,
		    a.address,
		    a.phone_address_consent,
		    a.processing_restricted
		FROM accounts a
		JOIN instruments i ON i.id = a.main_instrument_id
		WHERE a.id = ?`, accountID,
	).First(&row)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (AccountStore) SetProfile(tx *pop.Connection, accountID int64, firstName, lastName string, birthDate *time.Time, parentalConsentURI string) error {
	return tx.RawQuery(`
		UPDATE accounts
		SET first_name           = ?,
		    last_name            = ?,
		    birth_date           = ?,
		    parental_consent_uri = ?
		WHERE id = ?`,
		firstName, lastName, birthDate, parentalConsentURI, accountID,
	).Exec()
}

func (AccountStore) UpdateProfile(tx *pop.Connection, accountID int64, firstName, lastName, email string, instrumentID int64, birthDate *time.Time, parentalConsentURI, phone, address string) error {
	return tx.RawQuery(`
		UPDATE accounts
		SET first_name           = ?,
		    last_name            = ?,
		    email                = ?,
		    main_instrument_id   = ?,
		    birth_date           = ?,
		    parental_consent_uri = ?,
		    phone                = ?,
		    address              = ?
		WHERE id = ?`,
		firstName, lastName, email, instrumentID, birthDate, parentalConsentURI, phone, address, accountID,
	).Exec()
}

func (AccountStore) ListNonAnonymized(tx *pop.Connection) ([]MusicianListRow, error) {
	var rows []MusicianListRow
	err := tx.RawQuery(`
		SELECT
		    a.id,
		    a.first_name,
		    a.last_name,
		    a.anonymization_token,
		    i.name AS instrument_name,
		    a.status,
		    CASE WHEN admin_roles.account_id IS NOT NULL THEN true ELSE false END AS is_admin
		FROM accounts a
		JOIN instruments i ON i.id = a.main_instrument_id
		LEFT JOIN (
		    SELECT ar.account_id
		    FROM account_roles ar
		    JOIN roles r ON r.id = ar.role_id AND r.name = 'admin'
		) admin_roles ON admin_roles.account_id = a.id
		ORDER BY a.last_name ASC NULLS LAST, a.first_name ASC NULLS LAST`,
	).All(&rows)
	return rows, err
}

func (AccountStore) ListForRetentionReview(tx *pop.Connection) ([]RetentionRow, error) {
	var rows []RetentionRow
	err := tx.RawQuery(`
		SELECT
		    a.id,
		    a.first_name,
		    a.last_name,
		    i.name  AS instrument_name,
		    s.label AS last_season_label,
		    s.end_date
		FROM accounts a
		JOIN LATERAL (
		    SELECT season_id
		    FROM fee_payments
		    WHERE account_id = a.id
		    ORDER BY payment_date DESC
		    LIMIT 1
		) last_fp ON true
		JOIN seasons s ON s.id = last_fp.season_id
		JOIN instruments i ON i.id = a.main_instrument_id
		WHERE a.status != 'anonymized'
		  AND s.end_date < NOW() - INTERVAL '5 years'
		ORDER BY s.end_date ASC`,
	).All(&rows)
	return rows, err
}

func (AccountStore) ClearMembershipFields(tx *pop.Connection, accountID int64) error {
	return tx.RawQuery(`
		UPDATE accounts
		SET first_name            = NULL,
		    last_name             = NULL,
		    birth_date            = NULL,
		    parental_consent_uri  = NULL,
		    phone                 = NULL,
		    address               = NULL,
		    phone_address_consent  = false,
		    processing_restricted = false
		WHERE id = ?`, accountID,
	).Exec()
}

func (AccountStore) WithdrawConsent(tx *pop.Connection, accountID int64) error {
	return tx.RawQuery(`
		UPDATE accounts
		SET phone                = NULL,
		    address              = NULL,
		    phone_address_consent = false
		WHERE id = ?`, accountID,
	).Exec()
}

func (AccountStore) ToggleProcessingRestriction(tx *pop.Connection, accountID int64) error {
	return tx.RawQuery(`
		UPDATE accounts
		SET processing_restricted = NOT processing_restricted
		WHERE id = ?`, accountID,
	).Exec()
}
