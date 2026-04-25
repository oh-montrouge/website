package services_test

import (
	"errors"
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ohmontrouge/webapp/models"
	"ohmontrouge/webapp/services"
)

var (
	testPayDate1 = time.Date(2024, 9, 15, 0, 0, 0, 0, time.UTC)
	testPayDate2 = time.Date(2025, 9, 10, 0, 0, 0, 0, time.UTC)
)

// stubFeePaymentRepo implements services.FeePaymentRepository for service unit tests.
type stubFeePaymentRepo struct {
	createdID    int64
	createErr    error
	updateErr    error
	deleteErr    error
	rows         models.FeePaymentRows
	listErr      error
	row          *models.FeePaymentRow
	getErr       error
	firstDate    *time.Time
	firstDateErr error
}

func (s stubFeePaymentRepo) Create(_ *pop.Connection, _, _ int64, _ float64, _ time.Time, _, _ string) (int64, error) {
	return s.createdID, s.createErr
}

func (s stubFeePaymentRepo) Update(_ *pop.Connection, _ int64, _ float64, _ time.Time, _, _ string) error {
	return s.updateErr
}

func (s stubFeePaymentRepo) Delete(_ *pop.Connection, _ int64) error {
	return s.deleteErr
}

func (s stubFeePaymentRepo) ListByAccount(_ *pop.Connection, _ int64) (models.FeePaymentRows, error) {
	return s.rows, s.listErr
}

func (s stubFeePaymentRepo) GetByID(_ *pop.Connection, _ int64) (*models.FeePaymentRow, error) {
	return s.row, s.getErr
}

func (s stubFeePaymentRepo) GetFirstInscriptionDate(_ *pop.Connection, _ int64) (*time.Time, error) {
	return s.firstDate, s.firstDateErr
}

func TestFeePaymentService_Record_Success(t *testing.T) {
	svc := services.FeePaymentService{FeePayments: stubFeePaymentRepo{createdID: 1}}
	err := svc.Record(nil, 1, 1, 50.00, testPayDate1, "chèque", "")
	assert.NoError(t, err)
}

func TestFeePaymentService_Record_Duplicate(t *testing.T) {
	svc := services.FeePaymentService{FeePayments: stubFeePaymentRepo{createErr: models.ErrFeePaymentDuplicate}}
	err := svc.Record(nil, 1, 1, 50.00, testPayDate1, "chèque", "")
	assert.ErrorIs(t, err, services.ErrDuplicatePayment)
}

func TestFeePaymentService_Record_OtherError(t *testing.T) {
	svc := services.FeePaymentService{FeePayments: stubFeePaymentRepo{createErr: errors.New("db error")}}
	err := svc.Record(nil, 1, 1, 50.00, testPayDate1, "chèque", "")
	assert.Error(t, err)
	assert.NotErrorIs(t, err, services.ErrDuplicatePayment)
}

func TestFeePaymentService_Update(t *testing.T) {
	svc := services.FeePaymentService{FeePayments: stubFeePaymentRepo{}}
	err := svc.Update(nil, 1, 75.00, testPayDate2, "espèces", "updated")
	assert.NoError(t, err)
}

func TestFeePaymentService_Update_Error(t *testing.T) {
	svc := services.FeePaymentService{FeePayments: stubFeePaymentRepo{updateErr: errors.New("db error")}}
	err := svc.Update(nil, 1, 75.00, testPayDate2, "espèces", "")
	assert.Error(t, err)
}

func TestFeePaymentService_Delete(t *testing.T) {
	svc := services.FeePaymentService{FeePayments: stubFeePaymentRepo{}}
	assert.NoError(t, svc.Delete(nil, 1))
}

func TestFeePaymentService_Delete_Error(t *testing.T) {
	svc := services.FeePaymentService{FeePayments: stubFeePaymentRepo{deleteErr: errors.New("db error")}}
	assert.Error(t, svc.Delete(nil, 1))
}

func TestFeePaymentService_ListByAccount_Empty(t *testing.T) {
	svc := services.FeePaymentService{FeePayments: stubFeePaymentRepo{rows: models.FeePaymentRows{}}}
	dtos, err := svc.ListByAccount(nil, 1)
	require.NoError(t, err)
	assert.Empty(t, dtos)
}

func TestFeePaymentService_ListByAccount_MapsToDTO(t *testing.T) {
	rows := models.FeePaymentRows{
		{
			ID:          1,
			AccountID:   2,
			SeasonID:    3,
			SeasonLabel: "2025-2026",
			Amount:      50.00,
			PaymentDate: testPayDate1,
			PaymentType: "chèque",
			Comment:     nulls.NewString("note"),
		},
	}
	svc := services.FeePaymentService{FeePayments: stubFeePaymentRepo{rows: rows}}
	dtos, err := svc.ListByAccount(nil, 2)
	require.NoError(t, err)
	require.Len(t, dtos, 1)
	assert.Equal(t, int64(1), dtos[0].ID)
	assert.Equal(t, "2025-2026", dtos[0].SeasonLabel)
	assert.InDelta(t, 50.00, dtos[0].Amount, 0.001)
	assert.Equal(t, "note", dtos[0].Comment)
}

func TestFeePaymentService_GetByID_Found(t *testing.T) {
	row := &models.FeePaymentRow{
		ID:          5,
		AccountID:   1,
		SeasonID:    2,
		SeasonLabel: "2025-2026",
		Amount:      50.00,
		PaymentDate: testPayDate1,
		PaymentType: "espèces",
	}
	svc := services.FeePaymentService{FeePayments: stubFeePaymentRepo{row: row}}
	dto, err := svc.GetByID(nil, 5)
	require.NoError(t, err)
	require.NotNil(t, dto)
	assert.Equal(t, int64(5), dto.ID)
}

func TestFeePaymentService_GetByID_NotFound(t *testing.T) {
	svc := services.FeePaymentService{FeePayments: stubFeePaymentRepo{row: nil}}
	_, err := svc.GetByID(nil, 999)
	assert.ErrorIs(t, err, services.ErrFeePaymentNotFound)
}

// TestFeePaymentService_GetFirstInscriptionDate_WithPayments covers AC-M2.
func TestFeePaymentService_GetFirstInscriptionDate_WithPayments(t *testing.T) {
	svc := services.FeePaymentService{FeePayments: stubFeePaymentRepo{firstDate: &testPayDate1}}
	date, err := svc.GetFirstInscriptionDate(nil, 1)
	require.NoError(t, err)
	assert.NotNil(t, date)
	assert.Equal(t, testPayDate1, *date)
}

// TestFeePaymentService_GetFirstInscriptionDate_NoPayments covers AC-M2 nil case.
func TestFeePaymentService_GetFirstInscriptionDate_NoPayments(t *testing.T) {
	svc := services.FeePaymentService{FeePayments: stubFeePaymentRepo{firstDate: nil}}
	date, err := svc.GetFirstInscriptionDate(nil, 1)
	require.NoError(t, err)
	assert.Nil(t, date)
}
