package models

import (
	"errors"
	"strings"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
	"github.com/lib/pq"
)

// ErrFeePaymentDuplicate is returned by FeePaymentStore.Create when the
// UNIQUE(account_id, season_id) constraint is violated.
var ErrFeePaymentDuplicate = errors.New("duplicate fee payment for this account and season")

// FeePayment is the raw DB row for the fee_payments table.
type FeePayment struct {
	ID          int64        `db:"id"`
	AccountID   int64        `db:"account_id"`
	SeasonID    int64        `db:"season_id"`
	Amount      float64      `db:"amount"`
	PaymentDate time.Time    `db:"payment_date"`
	PaymentType string       `db:"payment_type"`
	Comment     nulls.String `db:"comment"`
}

// FeePaymentRow is a read model that includes the season label via JOIN.
type FeePaymentRow struct {
	ID          int64        `db:"id"`
	AccountID   int64        `db:"account_id"`
	SeasonID    int64        `db:"season_id"`
	SeasonLabel string       `db:"season_label"`
	Amount      float64      `db:"amount"`
	PaymentDate time.Time    `db:"payment_date"`
	PaymentType string       `db:"payment_type"`
	Comment     nulls.String `db:"comment"`
}

// FeePaymentRows is a slice of FeePaymentRow.
type FeePaymentRows []FeePaymentRow

// FeePaymentStore is the production implementation of services.FeePaymentRepository.
type FeePaymentStore struct{}

func (FeePaymentStore) Create(tx *pop.Connection, accountID, seasonID int64, amount float64, paymentDate time.Time, paymentType, comment string) (int64, error) {
	var row struct {
		ID int64 `db:"id"`
	}
	err := tx.RawQuery(
		`INSERT INTO fee_payments (account_id, season_id, amount, payment_date, payment_type, comment)
		 VALUES (?, ?, ?, ?, ?, NULLIF(?, '')) RETURNING id`,
		accountID, seasonID, amount, paymentDate, paymentType, comment,
	).First(&row)
	if err != nil {
		if isDuplicateError(err) {
			return 0, ErrFeePaymentDuplicate
		}
		return 0, err
	}
	return row.ID, nil
}

func (FeePaymentStore) Update(tx *pop.Connection, id int64, amount float64, paymentDate time.Time, paymentType, comment string) error {
	return tx.RawQuery(
		`UPDATE fee_payments SET amount=?, payment_date=?, payment_type=?, comment=NULLIF(?, '') WHERE id=?`,
		amount, paymentDate, paymentType, comment, id,
	).Exec()
}

func (FeePaymentStore) Delete(tx *pop.Connection, id int64) error {
	return tx.RawQuery(`DELETE FROM fee_payments WHERE id=?`, id).Exec()
}

func (FeePaymentStore) ListByAccount(tx *pop.Connection, accountID int64) (FeePaymentRows, error) {
	var rows FeePaymentRows
	err := tx.RawQuery(
		`SELECT fp.id, fp.account_id, fp.season_id, s.label AS season_label,
		        fp.amount, fp.payment_date, fp.payment_type, fp.comment
		 FROM fee_payments fp
		 JOIN seasons s ON s.id = fp.season_id
		 WHERE fp.account_id = ?
		 ORDER BY fp.payment_date DESC`,
		accountID,
	).All(&rows)
	return rows, err
}

func (FeePaymentStore) GetByID(tx *pop.Connection, id int64) (*FeePaymentRow, error) {
	var rows FeePaymentRows
	err := tx.RawQuery(
		`SELECT fp.id, fp.account_id, fp.season_id, s.label AS season_label,
		        fp.amount, fp.payment_date, fp.payment_type, fp.comment
		 FROM fee_payments fp
		 JOIN seasons s ON s.id = fp.season_id
		 WHERE fp.id = ?`,
		id,
	).All(&rows)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

// isDuplicateError reports whether err is a PostgreSQL unique-constraint violation (code 23505).
// Tries errors.As first; falls back to a string check because Pop does not always preserve
// the error chain when wrapping driver errors.
func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	// lib/pq includes the code in the error message as "ERROR: ... (SQLSTATE 23505)"
	// or via the Error string which may embed the code.
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return string(pqErr.Code) == "23505"
	}
	return strings.Contains(err.Error(), "23505")
}

func (FeePaymentStore) GetFirstInscriptionDate(tx *pop.Connection, accountID int64) (*time.Time, error) {
	var result struct {
		Date nulls.Time `db:"first_inscription_date"`
	}
	err := tx.RawQuery(
		`SELECT MIN(payment_date) AS first_inscription_date FROM fee_payments WHERE account_id = ?`,
		accountID,
	).First(&result)
	if err != nil {
		return nil, err
	}
	if !result.Date.Valid {
		return nil, nil
	}
	return &result.Date.Time, nil
}
