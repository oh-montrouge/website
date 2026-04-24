package services_test

import (
	"errors"
	"testing"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/stretchr/testify/assert"
	"ohmontrouge/webapp/models"
	"ohmontrouge/webapp/services"
)

type stubSeasonRepo struct {
	seasons      models.Seasons
	createdID    int64
	listErr      error
	createErr    error
	designateErr error
}

func (s stubSeasonRepo) Create(_ *pop.Connection, _ string, _, _ time.Time) (int64, error) {
	return s.createdID, s.createErr
}

func (s stubSeasonRepo) List(_ *pop.Connection) (models.Seasons, error) {
	return s.seasons, s.listErr
}

func (s stubSeasonRepo) DesignateCurrent(_ *pop.Connection, _ int64) error {
	return s.designateErr
}

var (
	testSeasonStart = time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	testSeasonEnd   = time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
)

func TestSeasonService_Create_Success(t *testing.T) {
	svc := services.SeasonService{Seasons: stubSeasonRepo{createdID: 1}}
	err := svc.Create(nil, "2025-2026", testSeasonStart, testSeasonEnd)
	assert.NoError(t, err)
}

func TestSeasonService_Create_Error(t *testing.T) {
	svc := services.SeasonService{Seasons: stubSeasonRepo{createErr: errors.New("db error")}}
	err := svc.Create(nil, "2025-2026", testSeasonStart, testSeasonEnd)
	assert.Error(t, err)
}

func TestSeasonService_List_Success(t *testing.T) {
	svc := services.SeasonService{Seasons: stubSeasonRepo{seasons: models.Seasons{
		{ID: 1, Label: "2025-2026", StartDate: testSeasonStart, EndDate: testSeasonEnd, IsCurrent: true},
		{ID: 2, Label: "2024-2025", StartDate: testSeasonStart.AddDate(-1, 0, 0), EndDate: testSeasonEnd.AddDate(-1, 0, 0), IsCurrent: false},
	}}}
	dtos, err := svc.List(nil)
	assert.NoError(t, err)
	assert.Len(t, dtos, 2)
	assert.Equal(t, "2025-2026", dtos[0].Label)
	assert.True(t, dtos[0].IsCurrent)
	assert.Equal(t, "2024-2025", dtos[1].Label)
	assert.False(t, dtos[1].IsCurrent)
}

func TestSeasonService_List_Empty(t *testing.T) {
	svc := services.SeasonService{Seasons: stubSeasonRepo{seasons: models.Seasons{}}}
	dtos, err := svc.List(nil)
	assert.NoError(t, err)
	assert.Empty(t, dtos)
}

func TestSeasonService_List_Error(t *testing.T) {
	svc := services.SeasonService{Seasons: stubSeasonRepo{listErr: errors.New("db error")}}
	_, err := svc.List(nil)
	assert.Error(t, err)
}

func TestSeasonService_DesignateCurrent_Success(t *testing.T) {
	svc := services.SeasonService{Seasons: stubSeasonRepo{}}
	err := svc.DesignateCurrent(nil, 1)
	assert.NoError(t, err)
}

func TestSeasonService_DesignateCurrent_Error(t *testing.T) {
	svc := services.SeasonService{Seasons: stubSeasonRepo{designateErr: errors.New("db error")}}
	err := svc.DesignateCurrent(nil, 1)
	assert.Error(t, err)
}
