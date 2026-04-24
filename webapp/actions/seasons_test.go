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

type stubSeasonManager struct {
	seasons      []services.SeasonDTO
	listErr      error
	createErr    error
	designateErr error
}

func (s stubSeasonManager) Create(_ *pop.Connection, _ string, _, _ time.Time) error {
	return s.createErr
}

func (s stubSeasonManager) List(_ *pop.Connection) ([]services.SeasonDTO, error) {
	return s.seasons, s.listErr
}

func (s stubSeasonManager) DesignateCurrent(_ *pop.Connection, _ int64) error {
	return s.designateErr
}

func newSeasonsTestApp(h SeasonsHandler) http.Handler {
	return newTestApp(func(a *buffalo.App) {
		a.GET("/admin/saisons", h.Index)
		a.POST("/admin/saisons", h.Create)
		a.POST("/admin/saisons/{id}/courante", h.DesignateCurrent)
	})
}

// --- Index ---

func TestSeasonsHandler_Index_Success(t *testing.T) {
	h := SeasonsHandler{Seasons: stubSeasonManager{seasons: []services.SeasonDTO{
		{ID: 1, Label: "2025-2026", IsCurrent: true},
		{ID: 2, Label: "2024-2025", IsCurrent: false},
	}}}
	app := newSeasonsTestApp(h)

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/saisons", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "2025-2026")
	assert.Contains(t, res.Body.String(), "2024-2025")
}

func TestSeasonsHandler_Index_ListError(t *testing.T) {
	h := SeasonsHandler{Seasons: stubSeasonManager{listErr: errors.New("db error")}}
	app := newSeasonsTestApp(h)

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/saisons", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

// --- Create ---

func TestSeasonsHandler_Create_Success(t *testing.T) {
	h := SeasonsHandler{Seasons: stubSeasonManager{}}
	app := newSeasonsTestApp(h)

	res := httptest.NewRecorder()
	body := strings.NewReader("label=2025-2026&start_date=2025-09-01&end_date=2026-08-31")
	req := httptest.NewRequest(http.MethodPost, "/admin/saisons", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/admin/saisons", res.Header().Get("Location"))
}

func TestSeasonsHandler_Create_EmptyLabel(t *testing.T) {
	h := SeasonsHandler{Seasons: stubSeasonManager{}}
	app := newSeasonsTestApp(h)

	res := httptest.NewRecorder()
	body := strings.NewReader("label=&start_date=2025-09-01&end_date=2026-08-31")
	req := httptest.NewRequest(http.MethodPost, "/admin/saisons", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
	assert.Contains(t, res.Body.String(), "libellé")
}

func TestSeasonsHandler_Create_InvalidStartDate(t *testing.T) {
	h := SeasonsHandler{Seasons: stubSeasonManager{}}
	app := newSeasonsTestApp(h)

	res := httptest.NewRecorder()
	body := strings.NewReader("label=2025-2026&start_date=not-a-date&end_date=2026-08-31")
	req := httptest.NewRequest(http.MethodPost, "/admin/saisons", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
	assert.Contains(t, res.Body.String(), "début")
}

func TestSeasonsHandler_Create_InvalidEndDate(t *testing.T) {
	h := SeasonsHandler{Seasons: stubSeasonManager{}}
	app := newSeasonsTestApp(h)

	res := httptest.NewRecorder()
	body := strings.NewReader("label=2025-2026&start_date=2025-09-01&end_date=not-a-date")
	req := httptest.NewRequest(http.MethodPost, "/admin/saisons", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
	assert.Contains(t, res.Body.String(), "fin")
}

func TestSeasonsHandler_Create_EndBeforeStart(t *testing.T) {
	h := SeasonsHandler{Seasons: stubSeasonManager{}}
	app := newSeasonsTestApp(h)

	res := httptest.NewRecorder()
	body := strings.NewReader("label=2025-2026&start_date=2026-08-31&end_date=2025-09-01")
	req := httptest.NewRequest(http.MethodPost, "/admin/saisons", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
	assert.Contains(t, res.Body.String(), "fin")
}

func TestSeasonsHandler_Create_ServiceError(t *testing.T) {
	h := SeasonsHandler{Seasons: stubSeasonManager{createErr: errors.New("db error")}}
	app := newSeasonsTestApp(h)

	res := httptest.NewRecorder()
	body := strings.NewReader("label=2025-2026&start_date=2025-09-01&end_date=2026-08-31")
	req := httptest.NewRequest(http.MethodPost, "/admin/saisons", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

// --- DesignateCurrent ---

func TestSeasonsHandler_DesignateCurrent_Success(t *testing.T) {
	h := SeasonsHandler{Seasons: stubSeasonManager{}}
	app := newSeasonsTestApp(h)

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/saisons/1/courante", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/admin/saisons", res.Header().Get("Location"))
}

func TestSeasonsHandler_DesignateCurrent_InvalidID(t *testing.T) {
	h := SeasonsHandler{Seasons: stubSeasonManager{}}
	app := newSeasonsTestApp(h)

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/saisons/not-a-number/courante", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
}

func TestSeasonsHandler_DesignateCurrent_ServiceError(t *testing.T) {
	h := SeasonsHandler{Seasons: stubSeasonManager{designateErr: errors.New("db error")}}
	app := newSeasonsTestApp(h)

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/saisons/1/courante", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}
