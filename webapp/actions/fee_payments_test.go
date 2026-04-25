package actions

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/stretchr/testify/assert"
	"ohmontrouge/webapp/services"
)

// stubFeePaymentManager implements services.FeePaymentManager for handler unit tests.
type stubFeePaymentManager struct {
	recordErr    error
	updateErr    error
	deleteErr    error
	payments     []services.FeePaymentDTO
	listErr      error
	payment      *services.FeePaymentDTO
	getErr       error
	firstDate    *time.Time
	firstDateErr error
}

func (s *stubFeePaymentManager) Record(_ *pop.Connection, _, _ int64, _ float64, _ time.Time, _, _ string) error {
	return s.recordErr
}

func (s *stubFeePaymentManager) Update(_ *pop.Connection, _ int64, _ float64, _ time.Time, _, _ string) error {
	return s.updateErr
}

func (s *stubFeePaymentManager) Delete(_ *pop.Connection, _ int64) error {
	return s.deleteErr
}

func (s *stubFeePaymentManager) ListByAccount(_ *pop.Connection, _ int64) ([]services.FeePaymentDTO, error) {
	return s.payments, s.listErr
}

func (s *stubFeePaymentManager) GetByID(_ *pop.Connection, _ int64) (*services.FeePaymentDTO, error) {
	return s.payment, s.getErr
}

func (s *stubFeePaymentManager) GetFirstInscriptionDate(_ *pop.Connection, _ int64) (*time.Time, error) {
	return s.firstDate, s.firstDateErr
}

func defaultPayment() *services.FeePaymentDTO {
	return &services.FeePaymentDTO{
		ID:          1,
		AccountID:   2,
		SeasonID:    3,
		SeasonLabel: "2025-2026",
		Amount:      50.00,
		PaymentDate: time.Date(2025, 9, 10, 0, 0, 0, 0, time.UTC),
		PaymentType: "chèque",
		Comment:     "",
	}
}

func newFeePaymentsTestApp(h FeePaymentsHandler) http.Handler {
	return newTestApp(func(a *buffalo.App) {
		a.POST("/admin/musiciens/{account_id}/cotisations", h.Create)
		a.GET("/admin/cotisations/{id}/modifier", h.EditForm)
		a.PUT("/admin/cotisations/{id}", h.Update)
		a.DELETE("/admin/cotisations/{id}", h.Delete)
	})
}

// --- Create ---

func TestFeePaymentsHandler_Create_Success(t *testing.T) {
	h := FeePaymentsHandler{FeePayments: &stubFeePaymentManager{}}
	app := newFeePaymentsTestApp(h)

	body := strings.NewReader("season_id=1&amount=50.00&payment_date=2025-09-10&payment_type=ch%C3%A8que&comment=")
	req := httptest.NewRequest(http.MethodPost, "/admin/musiciens/1/cotisations", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/admin/musiciens/1", res.Header().Get("Location"))
}

func TestFeePaymentsHandler_Create_DuplicateFlash(t *testing.T) {
	h := FeePaymentsHandler{FeePayments: &stubFeePaymentManager{recordErr: services.ErrDuplicatePayment}}
	app := newFeePaymentsTestApp(h)

	body := strings.NewReader("season_id=1&amount=50.00&payment_date=2025-09-10&payment_type=ch%C3%A8que")
	req := httptest.NewRequest(http.MethodPost, "/admin/musiciens/1/cotisations", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/admin/musiciens/1", res.Header().Get("Location"))
}

func TestFeePaymentsHandler_Create_InvalidSeasonID(t *testing.T) {
	h := FeePaymentsHandler{FeePayments: &stubFeePaymentManager{}}
	app := newFeePaymentsTestApp(h)

	body := strings.NewReader("season_id=notanumber&amount=50.00&payment_date=2025-09-10&payment_type=ch%C3%A8que")
	req := httptest.NewRequest(http.MethodPost, "/admin/musiciens/1/cotisations", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
}

func TestFeePaymentsHandler_Create_InvalidAmount(t *testing.T) {
	h := FeePaymentsHandler{FeePayments: &stubFeePaymentManager{}}
	app := newFeePaymentsTestApp(h)

	body := strings.NewReader("season_id=1&amount=notanumber&payment_date=2025-09-10&payment_type=ch%C3%A8que")
	req := httptest.NewRequest(http.MethodPost, "/admin/musiciens/1/cotisations", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
}

func TestFeePaymentsHandler_Create_InvalidDate(t *testing.T) {
	h := FeePaymentsHandler{FeePayments: &stubFeePaymentManager{}}
	app := newFeePaymentsTestApp(h)

	body := strings.NewReader("season_id=1&amount=50.00&payment_date=not-a-date&payment_type=ch%C3%A8que")
	req := httptest.NewRequest(http.MethodPost, "/admin/musiciens/1/cotisations", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
}

func TestFeePaymentsHandler_Create_MissingPaymentType(t *testing.T) {
	h := FeePaymentsHandler{FeePayments: &stubFeePaymentManager{}}
	app := newFeePaymentsTestApp(h)

	body := strings.NewReader("season_id=1&amount=50.00&payment_date=2025-09-10&payment_type=")
	req := httptest.NewRequest(http.MethodPost, "/admin/musiciens/1/cotisations", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
}

func TestFeePaymentsHandler_Create_ServiceError(t *testing.T) {
	h := FeePaymentsHandler{FeePayments: &stubFeePaymentManager{recordErr: errors.New("db error")}}
	app := newFeePaymentsTestApp(h)

	body := strings.NewReader("season_id=1&amount=50.00&payment_date=2025-09-10&payment_type=ch%C3%A8que")
	req := httptest.NewRequest(http.MethodPost, "/admin/musiciens/1/cotisations", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestFeePaymentsHandler_Update_InvalidDate(t *testing.T) {
	h := FeePaymentsHandler{FeePayments: &stubFeePaymentManager{payment: defaultPayment()}}
	app := newFeePaymentsTestApp(h)

	body := strings.NewReader("amount=75.00&payment_date=not-a-date&payment_type=esp%C3%A8ces")
	req := httptest.NewRequest(http.MethodPut, "/admin/cotisations/1", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
}

func TestFeePaymentsHandler_Update_MissingPaymentType(t *testing.T) {
	h := FeePaymentsHandler{FeePayments: &stubFeePaymentManager{payment: defaultPayment()}}
	app := newFeePaymentsTestApp(h)

	body := strings.NewReader("amount=75.00&payment_date=2025-10-01&payment_type=")
	req := httptest.NewRequest(http.MethodPut, "/admin/cotisations/1", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
}

// --- EditForm ---

func TestFeePaymentsHandler_EditForm_Success(t *testing.T) {
	h := FeePaymentsHandler{FeePayments: &stubFeePaymentManager{payment: defaultPayment()}}
	app := newFeePaymentsTestApp(h)

	req := httptest.NewRequest(http.MethodGet, "/admin/cotisations/1/modifier", nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "2025-2026")
}

func TestFeePaymentsHandler_EditForm_NotFound(t *testing.T) {
	h := FeePaymentsHandler{FeePayments: &stubFeePaymentManager{getErr: services.ErrFeePaymentNotFound}}
	app := newFeePaymentsTestApp(h)

	req := httptest.NewRequest(http.MethodGet, "/admin/cotisations/999/modifier", nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNotFound, res.Code)
}

// --- Update ---

func TestFeePaymentsHandler_Update_Success(t *testing.T) {
	h := FeePaymentsHandler{FeePayments: &stubFeePaymentManager{payment: defaultPayment()}}
	app := newFeePaymentsTestApp(h)

	body := strings.NewReader("amount=75.00&payment_date=2025-10-01&payment_type=esp%C3%A8ces&comment=updated")
	req := httptest.NewRequest(http.MethodPut, "/admin/cotisations/1", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/admin/musiciens/2", res.Header().Get("Location"))
}

func TestFeePaymentsHandler_Update_NotFound(t *testing.T) {
	h := FeePaymentsHandler{FeePayments: &stubFeePaymentManager{getErr: services.ErrFeePaymentNotFound}}
	app := newFeePaymentsTestApp(h)

	body := strings.NewReader("amount=75.00&payment_date=2025-10-01&payment_type=esp%C3%A8ces")
	req := httptest.NewRequest(http.MethodPut, "/admin/cotisations/999", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestFeePaymentsHandler_Update_InvalidAmount(t *testing.T) {
	h := FeePaymentsHandler{FeePayments: &stubFeePaymentManager{payment: defaultPayment()}}
	app := newFeePaymentsTestApp(h)

	body := strings.NewReader("amount=bad&payment_date=2025-10-01&payment_type=esp%C3%A8ces")
	req := httptest.NewRequest(http.MethodPut, "/admin/cotisations/1", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
}

// --- Delete ---

func TestFeePaymentsHandler_Delete_Success(t *testing.T) {
	h := FeePaymentsHandler{FeePayments: &stubFeePaymentManager{payment: defaultPayment()}}
	app := newFeePaymentsTestApp(h)

	req := httptest.NewRequest(http.MethodDelete, "/admin/cotisations/1", nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/admin/musiciens/2", res.Header().Get("Location"))
}

func TestFeePaymentsHandler_Delete_NotFound(t *testing.T) {
	h := FeePaymentsHandler{FeePayments: &stubFeePaymentManager{getErr: services.ErrFeePaymentNotFound}}
	app := newFeePaymentsTestApp(h)

	req := httptest.NewRequest(http.MethodDelete, "/admin/cotisations/999", nil)
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNotFound, res.Code)
}
