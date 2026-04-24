package models

import (
	"time"

	"github.com/gobuffalo/pop/v6"
)

// Season represents an annual membership period.
type Season struct {
	ID        int64     `db:"id"`
	Label     string    `db:"label"`
	StartDate time.Time `db:"start_date"`
	EndDate   time.Time `db:"end_date"`
	IsCurrent bool      `db:"is_current"`
}

// Seasons is a slice of Season.
type Seasons []Season

// SeasonStore is the production implementation of services.SeasonRepository.
type SeasonStore struct{}

func (SeasonStore) Create(tx *pop.Connection, label string, startDate, endDate time.Time) (int64, error) {
	var row struct {
		ID int64 `db:"id"`
	}
	err := tx.RawQuery(
		`INSERT INTO seasons (label, start_date, end_date, is_current) VALUES (?, ?, ?, false) RETURNING id`,
		label, startDate, endDate,
	).First(&row)
	return row.ID, err
}

func (SeasonStore) List(tx *pop.Connection) (Seasons, error) {
	var ss Seasons
	err := tx.Order("start_date DESC").All(&ss)
	return ss, err
}

// DesignateCurrent transfers the current designation to the given season.
// Both UPDATEs run on the same *pop.Connection (the middleware transaction),
// maintaining the exactly-one-current invariant atomically.
// Returns sql.ErrNoRows if id does not exist, preventing a silent clear with no replacement.
func (SeasonStore) DesignateCurrent(tx *pop.Connection, id int64) error {
	if err := tx.RawQuery(`UPDATE seasons SET is_current = false WHERE is_current = true`).Exec(); err != nil {
		return err
	}
	var row struct {
		ID int64 `db:"id"`
	}
	return tx.RawQuery(`UPDATE seasons SET is_current = true WHERE id = ? RETURNING id`, id).First(&row)
}
