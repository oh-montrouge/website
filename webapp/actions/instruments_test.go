package actions

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/stretchr/testify/assert"
	"ohmontrouge/webapp/models"
)

type stubInstruments struct {
	data models.Instruments
}

func (s stubInstruments) List(_ *pop.Connection) (models.Instruments, error) { return s.data, nil }

type errInstruments struct{ err error }

func (e errInstruments) List(_ *pop.Connection) (models.Instruments, error) { return nil, e.err }

func TestInstrumentsIndex_ReturnsInstruments(t *testing.T) {
	h := InstrumentsHandler{
		Instruments: stubInstruments{data: models.Instruments{
			{ID: 1, Name: "Alto"},
			{ID: 2, Name: "Violon"},
		}},
	}
	app := newTestApp(func(a *buffalo.App) {
		a.GET("/instruments", h.Index)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/instruments", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "Alto")
	assert.Contains(t, res.Body.String(), "Violon")
}

func TestInstrumentsIndex_ReturnsErrorOnRepositoryFailure(t *testing.T) {
	h := InstrumentsHandler{
		Instruments: errInstruments{err: errors.New("db failure")},
	}
	app := newTestApp(func(a *buffalo.App) {
		a.GET("/instruments", h.Index)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/instruments", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}
