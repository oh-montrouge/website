package actions

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ohmontrouge/webapp/models"
	"ohmontrouge/webapp/services"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/stretchr/testify/assert"
)

// --- stubs ---

type stubEventManager struct {
	summaries      []services.EventSummaryDTO
	listErr        error
	detail         *services.EventDetailDTO
	detailErr      error
	createErr      error
	updateErr      error
	deleteErr      error
	rsvpErr        error
	field          *services.EventFieldDTO
	fieldErr       error
	addFieldErr    error
	updateFieldErr error
	deleteFieldErr error
}

func (s *stubEventManager) ListForMember(_ *pop.Connection, _ int64) ([]services.EventSummaryDTO, error) {
	return s.summaries, s.listErr
}
func (s *stubEventManager) ListAll(_ *pop.Connection, _ int64) ([]services.EventSummaryDTO, error) {
	return s.summaries, s.listErr
}
func (s *stubEventManager) AdminListAll(_ *pop.Connection) ([]services.EventSummaryDTO, error) {
	return s.summaries, s.listErr
}
func (s *stubEventManager) GetDetail(_ *pop.Connection, _, _ int64) (*services.EventDetailDTO, error) {
	return s.detail, s.detailErr
}
func (s *stubEventManager) Create(_ *pop.Connection, _, _ string, _ time.Time) error {
	return s.createErr
}
func (s *stubEventManager) Update(_ *pop.Connection, _ int64, _, _ string, _ time.Time) error {
	return s.updateErr
}
func (s *stubEventManager) Delete(_ *pop.Connection, _ int64) error { return s.deleteErr }
func (s *stubEventManager) UpdateRSVP(_ *pop.Connection, _, _ int64, _ string, _ *int64, _ []services.FieldResponseInput) error {
	return s.rsvpErr
}
func (s *stubEventManager) GetField(_ *pop.Connection, _ int64) (*services.EventFieldDTO, error) {
	return s.field, s.fieldErr
}
func (s *stubEventManager) AddField(_ *pop.Connection, _ int64, _, _ string, _ bool, _ int, _ []services.FieldChoiceInput) error {
	return s.addFieldErr
}
func (s *stubEventManager) UpdateField(_ *pop.Connection, _ int64, _, _ string, _ bool, _ int, _ []services.FieldChoiceInput) error {
	return s.updateFieldErr
}
func (s *stubEventManager) DeleteField(_ *pop.Connection, _ int64) error         { return s.deleteFieldErr }
func (s *stubEventManager) SeedRSVPsForAccount(_ *pop.Connection, _ int64) error { return nil }

type stubInstrumentRepoE struct {
	instruments models.Instruments
	err         error
}

func (s stubInstrumentRepoE) List(_ *pop.Connection) (models.Instruments, error) {
	return s.instruments, s.err
}

type stubMusicianProfileMgrE struct {
	profile *services.MusicianProfile
	err     error
}

func (s *stubMusicianProfileMgrE) GetProfile(_ *pop.Connection, _ int64) (*services.MusicianProfile, error) {
	return s.profile, s.err
}
func (s *stubMusicianProfileMgrE) SetInitialProfile(_ *pop.Connection, _ int64, _, _ string, _ *time.Time, _ string) error {
	return nil
}
func (s *stubMusicianProfileMgrE) UpdateProfile(_ *pop.Connection, _ int64, _, _, _ string, _ int64, _ *time.Time, _, _, _ string) error {
	return nil
}
func (s *stubMusicianProfileMgrE) ListNonAnonymized(_ *pop.Connection) ([]services.MusicianProfileSummary, error) {
	return nil, nil
}
func (s *stubMusicianProfileMgrE) ConsentWithdrawal(_ *pop.Connection, _ int64) error { return nil }
func (s *stubMusicianProfileMgrE) ToggleProcessingRestriction(_ *pop.Connection, _ int64) error {
	return nil
}

// injectAccount injects current_account and is_admin into every request context.
func injectAccount(account *services.AccountDTO, isAdmin bool) buffalo.MiddlewareFunc {
	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			c.Set("current_account", account)
			c.Set("is_admin", isAdmin)
			return next(c)
		}
	}
}

func newEventsTestApp(h EventsHandler, account *services.AccountDTO, isAdmin bool, register func(*buffalo.App)) http.Handler {
	return newTestApp(func(a *buffalo.App) {
		a.Use(injectAccount(account, isAdmin))
		register(a)
	})
}

func defaultAccount() *services.AccountDTO {
	return &services.AccountDTO{ID: 1, Email: "alice@ohm.test", Status: services.StatusActive}
}

// --- Dashboard ---

func TestEventsHandler_Dashboard_RendersList(t *testing.T) {
	h := EventsHandler{
		Events: &stubEventManager{summaries: []services.EventSummaryDTO{
			{ID: 1, Name: "Répétition de mai", EventType: "rehearsal", Datetime: time.Now().Add(7 * 24 * time.Hour), RSVPState: "yes"},
		}},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), false, func(a *buffalo.App) {
		a.GET("/tableau-de-bord", h.Dashboard)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tableau-de-bord", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "Répétition de mai")
	assert.Contains(t, res.Body.String(), "Tableau de bord")
}

func TestEventsHandler_Dashboard_Empty(t *testing.T) {
	h := EventsHandler{
		Events:      &stubEventManager{summaries: []services.EventSummaryDTO{}},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), false, func(a *buffalo.App) {
		a.GET("/tableau-de-bord", h.Dashboard)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tableau-de-bord", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "Aucun événement")
}

func TestEventsHandler_Dashboard_ServiceError(t *testing.T) {
	h := EventsHandler{
		Events:      &stubEventManager{listErr: errors.New("db error")},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), false, func(a *buffalo.App) {
		a.GET("/tableau-de-bord", h.Dashboard)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tableau-de-bord", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

// --- Index ---

func TestEventsHandler_Index_RendersList(t *testing.T) {
	h := EventsHandler{
		Events: &stubEventManager{summaries: []services.EventSummaryDTO{
			{ID: 2, Name: "Concert de printemps", EventType: "concert", Datetime: time.Now().Add(30 * 24 * time.Hour), RSVPState: "unanswered"},
		}},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), false, func(a *buffalo.App) {
		a.GET("/evenements", h.Index)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/evenements", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "Concert de printemps")
}

// --- Show ---

func TestEventsHandler_Show_RendersDetail(t *testing.T) {
	h := EventsHandler{
		Events: &stubEventManager{detail: &services.EventDetailDTO{
			ID:        1,
			Name:      "Concert de printemps",
			EventType: "concert",
			Datetime:  time.Now().Add(30 * 24 * time.Hour),
			OwnRSVP:   &services.OwnRSVPDTO{ID: 10, State: "unanswered"},
		}},
		Instruments: stubInstrumentRepoE{instruments: models.Instruments{{ID: 1, Name: "Clarinette"}}},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), false, func(a *buffalo.App) {
		a.GET("/evenements/{id}", h.Show)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/evenements/1", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "Concert de printemps")
	assert.Contains(t, res.Body.String(), "Ma participation")
}

func TestEventsHandler_Show_NotFound(t *testing.T) {
	h := EventsHandler{
		Events:      &stubEventManager{detailErr: services.ErrEventNotFound},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), false, func(a *buffalo.App) {
		a.GET("/evenements/{id}", h.Show)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/evenements/99", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestEventsHandler_Show_InvalidID(t *testing.T) {
	h := EventsHandler{
		Events:      &stubEventManager{},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), false, func(a *buffalo.App) {
		a.GET("/evenements/{id}", h.Show)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/evenements/not-a-number", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNotFound, res.Code)
}

// --- UpdateRSVP ---

func TestEventsHandler_UpdateRSVP_Success(t *testing.T) {
	h := EventsHandler{
		Events:      &stubEventManager{},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), false, func(a *buffalo.App) {
		a.POST("/evenements/{id}/rsvp", h.UpdateRSVP)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("state=yes&instrument_id=1")
	req := httptest.NewRequest(http.MethodPost, "/evenements/1/rsvp", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/evenements/1", res.Header().Get("Location"))
}

func TestEventsHandler_UpdateRSVP_EmptyState(t *testing.T) {
	h := EventsHandler{
		Events:      &stubEventManager{},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), false, func(a *buffalo.App) {
		a.POST("/evenements/{id}/rsvp", h.UpdateRSVP)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("state=")
	req := httptest.NewRequest(http.MethodPost, "/evenements/1/rsvp", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
}

func TestEventsHandler_UpdateRSVP_ServiceError(t *testing.T) {
	h := EventsHandler{
		Events:      &stubEventManager{rsvpErr: errors.New("db error")},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), false, func(a *buffalo.App) {
		a.POST("/evenements/{id}/rsvp", h.UpdateRSVP)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("state=yes")
	req := httptest.NewRequest(http.MethodPost, "/evenements/1/rsvp", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestEventsHandler_UpdateRSVP_InstrumentRequired(t *testing.T) {
	h := EventsHandler{
		Events:      &stubEventManager{rsvpErr: services.ErrInstrumentRequired},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), false, func(a *buffalo.App) {
		a.POST("/evenements/{id}/rsvp", h.UpdateRSVP)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("state=yes")
	req := httptest.NewRequest(http.MethodPost, "/evenements/1/rsvp", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
}

// --- New ---

func TestEventsHandler_New_RendersForm(t *testing.T) {
	h := EventsHandler{
		Events:      &stubEventManager{},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), true, func(a *buffalo.App) {
		a.GET("/admin/evenements/nouveau", h.New)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/evenements/nouveau", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "Nouvel événement")
}

// --- Create ---

func TestEventsHandler_Create_Success(t *testing.T) {
	h := EventsHandler{
		Events:      &stubEventManager{},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), true, func(a *buffalo.App) {
		a.POST("/admin/evenements", h.Create)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("name=Concert+de+printemps&date=2026-06-01&time=20:00&event_type=concert")
	req := httptest.NewRequest(http.MethodPost, "/admin/evenements", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/admin/evenements", res.Header().Get("Location"))
}

func TestEventsHandler_Create_MissingName(t *testing.T) {
	h := EventsHandler{
		Events:      &stubEventManager{},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), true, func(a *buffalo.App) {
		a.POST("/admin/evenements", h.Create)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("name=&date=2026-06-01&time=20:00&event_type=concert")
	req := httptest.NewRequest(http.MethodPost, "/admin/evenements", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
	assert.Contains(t, res.Body.String(), "nom")
}

func TestEventsHandler_Create_InvalidType(t *testing.T) {
	h := EventsHandler{
		Events:      &stubEventManager{},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), true, func(a *buffalo.App) {
		a.POST("/admin/evenements", h.Create)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("name=Concert&date=2026-06-01&time=20:00&event_type=invalid")
	req := httptest.NewRequest(http.MethodPost, "/admin/evenements", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
}

func TestEventsHandler_Create_InvalidDate(t *testing.T) {
	h := EventsHandler{
		Events:      &stubEventManager{},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), true, func(a *buffalo.App) {
		a.POST("/admin/evenements", h.Create)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("name=Concert&date=not-a-date&time=20:00&event_type=concert")
	req := httptest.NewRequest(http.MethodPost, "/admin/evenements", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
}

// --- Delete ---

func TestEventsHandler_Delete_Success(t *testing.T) {
	h := EventsHandler{
		Events:      &stubEventManager{},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), true, func(a *buffalo.App) {
		a.DELETE("/admin/evenements/{id}", h.Delete)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("confirmed=true")
	req := httptest.NewRequest(http.MethodDelete, "/admin/evenements/1", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/admin/evenements", res.Header().Get("Location"))
}

func TestEventsHandler_Delete_NotConfirmed(t *testing.T) {
	h := EventsHandler{
		Events:      &stubEventManager{},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), true, func(a *buffalo.App) {
		a.DELETE("/admin/evenements/{id}", h.Delete)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/admin/evenements/1", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
}

// --- AdminUpdateRSVP (JSON endpoint) ---

func TestEventsHandler_AdminUpdateRSVP_Success(t *testing.T) {
	h := EventsHandler{
		Events:      &stubEventManager{},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), true, func(a *buffalo.App) {
		a.PATCH("/admin/evenements/{id}/rsvp/{musician_id}", h.AdminUpdateRSVP)
	})

	payload, _ := json.Marshal(map[string]string{"state": "yes"})
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/admin/evenements/1/rsvp/2", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), `"ok":true`)
}

func TestEventsHandler_AdminUpdateRSVP_InvalidEventID(t *testing.T) {
	h := EventsHandler{
		Events:      &stubEventManager{},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), true, func(a *buffalo.App) {
		a.PATCH("/admin/evenements/{id}/rsvp/{musician_id}", h.AdminUpdateRSVP)
	})

	payload, _ := json.Marshal(map[string]string{"state": "yes"})
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/admin/evenements/not-a-number/rsvp/2", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
}

func TestEventsHandler_AdminUpdateRSVP_ServiceError(t *testing.T) {
	h := EventsHandler{
		Events:      &stubEventManager{rsvpErr: errors.New("db error")},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), true, func(a *buffalo.App) {
		a.PATCH("/admin/evenements/{id}/rsvp/{musician_id}", h.AdminUpdateRSVP)
	})

	payload, _ := json.Marshal(map[string]string{"state": "yes"})
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/admin/evenements/1/rsvp/2", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestEventsHandler_AdminUpdateRSVP_RSVPNotFound(t *testing.T) {
	h := EventsHandler{
		Events:      &stubEventManager{rsvpErr: services.ErrRSVPNotFound},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), true, func(a *buffalo.App) {
		a.PATCH("/admin/evenements/{id}/rsvp/{musician_id}", h.AdminUpdateRSVP)
	})

	payload, _ := json.Marshal(map[string]string{"state": "yes"})
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/admin/evenements/1/rsvp/2", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestEventsHandler_AdminUpdateRSVP_InstrumentRequired(t *testing.T) {
	h := EventsHandler{
		Events:      &stubEventManager{rsvpErr: services.ErrInstrumentRequired},
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
	app := newEventsTestApp(h, defaultAccount(), true, func(a *buffalo.App) {
		a.PATCH("/admin/evenements/{id}/rsvp/{musician_id}", h.AdminUpdateRSVP)
	})

	payload, _ := json.Marshal(map[string]string{"state": "yes"})
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/admin/evenements/1/rsvp/2", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
}
