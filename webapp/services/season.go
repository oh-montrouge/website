package services

import (
	"time"

	"github.com/gobuffalo/pop/v6"
)

// SeasonDTO is the Membership-context view of a season.
type SeasonDTO struct {
	ID        int64
	Label     string
	StartDate time.Time
	EndDate   time.Time
	IsCurrent bool
}

// SeasonManager is the interface SeasonsHandler depends on.
type SeasonManager interface {
	Create(tx *pop.Connection, label string, startDate, endDate time.Time) error
	List(tx *pop.Connection) ([]SeasonDTO, error)
	DesignateCurrent(tx *pop.Connection, id int64) error
}

// SeasonService implements domain logic for season operations.
type SeasonService struct {
	Seasons SeasonRepository
}

func (s SeasonService) Create(tx *pop.Connection, label string, startDate, endDate time.Time) error {
	_, err := s.Seasons.Create(tx, label, startDate, endDate)
	return err
}

func (s SeasonService) List(tx *pop.Connection) ([]SeasonDTO, error) {
	ss, err := s.Seasons.List(tx)
	if err != nil {
		return nil, err
	}
	dtos := make([]SeasonDTO, len(ss))
	for i, season := range ss {
		dtos[i] = SeasonDTO{
			ID:        season.ID,
			Label:     season.Label,
			StartDate: season.StartDate,
			EndDate:   season.EndDate,
			IsCurrent: season.IsCurrent,
		}
	}
	return dtos, nil
}

func (s SeasonService) DesignateCurrent(tx *pop.Connection, id int64) error {
	return s.Seasons.DesignateCurrent(tx, id)
}
