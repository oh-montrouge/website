package actions

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
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
func (s *stubEventManager) GetDetail(_ *pop.Connection, _, _ int64) (*services.EventDetailDTO, error) {
	return s.detail, s.detailErr
}
func (s *stubEventManager) Create(_ *pop.Connection, _, _, _ string, _ time.Time) error {
	return s.createErr
}
func (s *stubEventManager) Update(_ *pop.Connection, _ int64, _, _, _ string, _ time.Time) error {
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

// newEventsHandler creates an EventsHandler with the given Events service and empty default stubs.
func newEventsHandler(events *stubEventManager) EventsHandler {
	return EventsHandler{
		Events:      events,
		Instruments: stubInstrumentRepoE{},
		Membership:  &stubMusicianProfileMgrE{},
	}
}

// runEventsGET builds the test app, registers handler at routeTemplate, and runs a GET to path.
func runEventsGET(h EventsHandler, isAdmin bool, routeTemplate, path string, fn buffalo.Handler) *httptest.ResponseRecorder {
	app := newEventsTestApp(h, defaultAccount(), isAdmin, func(a *buffalo.App) {
		a.GET(routeTemplate, fn)
	})
	res := httptest.NewRecorder()
	app.ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
	return res
}

// jscpd: unwanted abstraction — handler-specific form helpers share HTTP boilerplate; handler types differ, making a shared generic impractical
// jscpd:ignore-start
// runEventsPost builds the test app, registers handler at routeTemplate, and POSTs body to path.
func runEventsPost(h EventsHandler, isAdmin bool, routeTemplate, path, body string, fn buffalo.Handler) *httptest.ResponseRecorder {
	app := newEventsTestApp(h, defaultAccount(), isAdmin, func(a *buffalo.App) {
		a.POST(routeTemplate, fn)
	})
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	return res
}

// runEventsDelete builds the test app, registers handler at routeTemplate, and sends DELETE to path.
func runEventsDelete(h EventsHandler, isAdmin bool, routeTemplate, path, body string, fn buffalo.Handler) *httptest.ResponseRecorder {
	app := newEventsTestApp(h, defaultAccount(), isAdmin, func(a *buffalo.App) {
		a.DELETE(routeTemplate, fn)
	})
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req := httptest.NewRequest(http.MethodDelete, path, bodyReader)
	if body != "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	return res
}

// runEventsPut builds the test app, registers a PUT handler, and sends form body to path.
func runEventsPut(h EventsHandler, isAdmin bool, routeTemplate, path, body string, fn buffalo.Handler) *httptest.ResponseRecorder {
	app := newEventsTestApp(h, defaultAccount(), isAdmin, func(a *buffalo.App) {
		a.PUT(routeTemplate, fn)
	})
	req := httptest.NewRequest(http.MethodPut, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	return res
}

// jscpd:ignore-end

// runEventsJSON builds the test app, registers a PATCH handler, and sends JSON payload to path.
func runEventsJSON(h EventsHandler, isAdmin bool, routeTemplate, path string, payload []byte, fn buffalo.Handler) *httptest.ResponseRecorder {
	app := newEventsTestApp(h, defaultAccount(), isAdmin, func(a *buffalo.App) {
		a.PATCH(routeTemplate, fn)
	})
	req := httptest.NewRequest(http.MethodPatch, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	return res
}

// --- Dashboard ---

func TestEventsHandler_Dashboard_RendersList(t *testing.T) {
	h := newEventsHandler(&stubEventManager{summaries: []services.EventSummaryDTO{
		{ID: 1, Name: "Répétition de mai", EventType: "rehearsal", Datetime: time.Now().Add(7 * 24 * time.Hour), RSVPState: "yes"},
	}})
	res := runEventsGET(h, false, "/tableau-de-bord", "/tableau-de-bord", h.Dashboard)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "Répétition de mai")
	assert.Contains(t, res.Body.String(), "Tableau de bord")
}

func TestEventsHandler_Dashboard_Empty(t *testing.T) {
	h := newEventsHandler(&stubEventManager{summaries: []services.EventSummaryDTO{}})
	res := runEventsGET(h, false, "/tableau-de-bord", "/tableau-de-bord", h.Dashboard)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "Aucun événement")
}

func TestEventsHandler_Dashboard_ServiceError(t *testing.T) {
	h := newEventsHandler(&stubEventManager{listErr: errors.New("db error")})
	res := runEventsGET(h, false, "/tableau-de-bord", "/tableau-de-bord", h.Dashboard)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestEventsHandler_Dashboard_UsesDashboardTemplate(t *testing.T) {
	h := newEventsHandler(&stubEventManager{summaries: []services.EventSummaryDTO{
		{ID: 1, Name: "Concert de gala", EventType: "concert", Datetime: time.Now().Add(7 * 24 * time.Hour), RSVPState: "yes"},
	}})
	res := runEventsGET(h, false, "/tableau-de-bord", "/tableau-de-bord", h.Dashboard)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.NotContains(t, res.Body.String(), `aria-label="Liste des événements"`, "dashboard must not render the events table")
}

func TestEventsHandler_Dashboard_RendersDescription(t *testing.T) {
	h := newEventsHandler(&stubEventManager{summaries: []services.EventSummaryDTO{
		{ID: 1, Name: "Concert", EventType: "concert", Datetime: time.Now().Add(7 * 24 * time.Hour), RSVPState: "yes", Description: "**Gala** de printemps"},
	}})
	res := runEventsGET(h, false, "/tableau-de-bord", "/tableau-de-bord", h.Dashboard)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "<strong>Gala</strong>")
}

// --- Index ---

func TestEventsHandler_Index_RendersList(t *testing.T) {
	h := newEventsHandler(&stubEventManager{summaries: []services.EventSummaryDTO{
		{ID: 2, Name: "Concert de printemps", EventType: "concert", Datetime: time.Now().Add(30 * 24 * time.Hour), RSVPState: "unanswered"},
	}})
	res := runEventsGET(h, false, "/evenements", "/evenements", h.Index)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "Concert de printemps")
}

// --- Show ---

func TestEventsHandler_Show_RendersDetail(t *testing.T) {
	h := newEventsHandler(&stubEventManager{detail: &services.EventDetailDTO{
		ID:        1,
		Name:      "Concert de printemps",
		EventType: "concert",
		Datetime:  time.Now().Add(30 * 24 * time.Hour),
		OwnRSVP:   &services.OwnRSVPDTO{ID: 10, State: "unanswered"},
	}})
	h.Instruments = stubInstrumentRepoE{instruments: models.Instruments{{ID: 1, Name: "Clarinette"}}}
	res := runEventsGET(h, false, "/evenements/{id}", "/evenements/1", h.Show)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "Concert de printemps")
	assert.Contains(t, res.Body.String(), "Ma participation")
}

func TestEventsHandler_Show_NotFound(t *testing.T) {
	h := newEventsHandler(&stubEventManager{detailErr: services.ErrEventNotFound})
	res := runEventsGET(h, false, "/evenements/{id}", "/evenements/99", h.Show)
	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestEventsHandler_Show_InvalidID(t *testing.T) {
	h := newEventsHandler(&stubEventManager{})
	res := runEventsGET(h, false, "/evenements/{id}", "/evenements/not-a-number", h.Show)
	assert.Equal(t, http.StatusNotFound, res.Code)
}

// --- UpdateRSVP ---

func TestEventsHandler_UpdateRSVP_Success(t *testing.T) {
	h := newEventsHandler(&stubEventManager{})
	res := runEventsPost(h, false, "/evenements/{id}/rsvp", "/evenements/1/rsvp", "state=yes&instrument_id=1", h.UpdateRSVP)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/evenements/1", res.Header().Get("Location"))
}

func TestEventsHandler_UpdateRSVP_EmptyState(t *testing.T) {
	h := newEventsHandler(&stubEventManager{})
	res := runEventsPost(h, false, "/evenements/{id}/rsvp", "/evenements/1/rsvp", "state=", h.UpdateRSVP)
	assert.Equal(t, http.StatusSeeOther, res.Code)
}

func TestEventsHandler_UpdateRSVP_ServiceError(t *testing.T) {
	h := newEventsHandler(&stubEventManager{rsvpErr: errors.New("db error")})
	res := runEventsPost(h, false, "/evenements/{id}/rsvp", "/evenements/1/rsvp", "state=yes", h.UpdateRSVP)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestEventsHandler_UpdateRSVP_InstrumentRequired(t *testing.T) {
	h := newEventsHandler(&stubEventManager{rsvpErr: services.ErrInstrumentRequired})
	res := runEventsPost(h, false, "/evenements/{id}/rsvp", "/evenements/1/rsvp", "state=yes", h.UpdateRSVP)
	assert.Equal(t, http.StatusSeeOther, res.Code)
}

// --- New ---

func TestEventsHandler_New_RendersForm(t *testing.T) {
	h := newEventsHandler(&stubEventManager{})
	res := runEventsGET(h, true, "/admin/evenements/nouveau", "/admin/evenements/nouveau", h.New)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "Nouvel événement")
}

// --- Create ---

func TestEventsHandler_Create_Success(t *testing.T) {
	h := newEventsHandler(&stubEventManager{})
	res := runEventsPost(h, true, "/admin/evenements", "/admin/evenements",
		"name=Concert+de+printemps&date=2026-06-01&time=20:00&event_type=concert", h.Create)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/evenements", res.Header().Get("Location"))
}

func TestEventsHandler_Create_WithDescription(t *testing.T) {
	h := newEventsHandler(&stubEventManager{})
	res := runEventsPost(h, true, "/admin/evenements", "/admin/evenements",
		"name=Concert&date=2026-06-01&time=20:00&event_type=concert&description=**Gala**", h.Create)
	assert.Equal(t, http.StatusSeeOther, res.Code)
}

func TestEventsHandler_Create_WithoutDescription(t *testing.T) {
	h := newEventsHandler(&stubEventManager{})
	res := runEventsPost(h, true, "/admin/evenements", "/admin/evenements",
		"name=Concert&date=2026-06-01&time=20:00&event_type=concert", h.Create)
	assert.Equal(t, http.StatusSeeOther, res.Code)
}

func TestEventsHandler_Create_MissingName(t *testing.T) {
	h := newEventsHandler(&stubEventManager{})
	res := runEventsPost(h, true, "/admin/evenements", "/admin/evenements",
		"name=&date=2026-06-01&time=20:00&event_type=concert", h.Create)
	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
	assert.Contains(t, res.Body.String(), "nom")
}

func TestEventsHandler_Create_InvalidType(t *testing.T) {
	h := newEventsHandler(&stubEventManager{})
	res := runEventsPost(h, true, "/admin/evenements", "/admin/evenements",
		"name=Concert&date=2026-06-01&time=20:00&event_type=invalid", h.Create)
	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
}

func TestEventsHandler_Create_InvalidDate(t *testing.T) {
	h := newEventsHandler(&stubEventManager{})
	res := runEventsPost(h, true, "/admin/evenements", "/admin/evenements",
		"name=Concert&date=not-a-date&time=20:00&event_type=concert", h.Create)
	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
}

// --- Delete ---

func TestEventsHandler_Delete_Success(t *testing.T) {
	h := newEventsHandler(&stubEventManager{})
	res := runEventsDelete(h, true, "/admin/evenements/{id}", "/admin/evenements/1", "confirmed=true", h.Delete)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/evenements", res.Header().Get("Location"))
}

func TestEventsHandler_Delete_NotConfirmed(t *testing.T) {
	h := newEventsHandler(&stubEventManager{})
	res := runEventsDelete(h, true, "/admin/evenements/{id}", "/admin/evenements/1", "", h.Delete)
	assert.Equal(t, http.StatusSeeOther, res.Code)
}

// --- AdminUpdateRSVP (JSON endpoint) ---

func runAdminUpdateRSVP(h EventsHandler, eventPath string, payload []byte) *httptest.ResponseRecorder {
	return runEventsJSON(h, true,
		"/admin/evenements/{id}/rsvp/{musician_id}", eventPath, payload, h.AdminUpdateRSVP)
}

func TestEventsHandler_AdminUpdateRSVP_Success(t *testing.T) {
	payload, _ := json.Marshal(map[string]string{"state": "yes"})
	res := runAdminUpdateRSVP(newEventsHandler(&stubEventManager{}), "/admin/evenements/1/rsvp/2", payload)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), `"ok":true`)
}

func TestEventsHandler_AdminUpdateRSVP_InvalidEventID(t *testing.T) {
	payload, _ := json.Marshal(map[string]string{"state": "yes"})
	res := runAdminUpdateRSVP(newEventsHandler(&stubEventManager{}), "/admin/evenements/not-a-number/rsvp/2", payload)
	assert.Equal(t, http.StatusBadRequest, res.Code)
}

func TestEventsHandler_AdminUpdateRSVP_ServiceError(t *testing.T) {
	payload, _ := json.Marshal(map[string]string{"state": "yes"})
	res := runAdminUpdateRSVP(newEventsHandler(&stubEventManager{rsvpErr: errors.New("db error")}), "/admin/evenements/1/rsvp/2", payload)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestEventsHandler_AdminUpdateRSVP_RSVPNotFound(t *testing.T) {
	payload, _ := json.Marshal(map[string]string{"state": "yes"})
	res := runAdminUpdateRSVP(newEventsHandler(&stubEventManager{rsvpErr: services.ErrRSVPNotFound}), "/admin/evenements/1/rsvp/2", payload)
	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestEventsHandler_AdminUpdateRSVP_InstrumentRequired(t *testing.T) {
	payload, _ := json.Marshal(map[string]string{"state": "yes"})
	res := runAdminUpdateRSVP(newEventsHandler(&stubEventManager{rsvpErr: services.ErrInstrumentRequired}), "/admin/evenements/1/rsvp/2", payload)
	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
}

// --- Gaps on already-tested handlers ---

func TestEventsHandler_Index_ServiceError(t *testing.T) {
	h := newEventsHandler(&stubEventManager{listErr: errors.New("db")})
	res := runEventsGET(h, false, "/evenements", "/evenements", h.Index)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestEventsHandler_AdminUpdateRSVP_InvalidMusicianID(t *testing.T) {
	payload, _ := json.Marshal(map[string]string{"state": "yes"})
	res := runAdminUpdateRSVP(newEventsHandler(&stubEventManager{}), "/admin/evenements/1/rsvp/not-a-number", payload)
	assert.Equal(t, http.StatusBadRequest, res.Code)
}

func TestEventsHandler_AdminUpdateRSVP_InvalidBody(t *testing.T) {
	res := runAdminUpdateRSVP(newEventsHandler(&stubEventManager{}), "/admin/evenements/1/rsvp/2", []byte("not-json"))
	assert.Equal(t, http.StatusBadRequest, res.Code)
}

// --- Edit ---

func TestEventsHandler_Edit_RendersForm(t *testing.T) {
	h := newEventsHandler(&stubEventManager{detail: &services.EventDetailDTO{
		ID: 1, Name: "Concert de printemps", EventType: "concert", Datetime: time.Now().Add(30 * 24 * time.Hour),
	}})
	res := runEventsGET(h, true, "/admin/evenements/{id}/modifier", "/admin/evenements/1/modifier", h.Edit)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "Concert de printemps")
}

func TestEventsHandler_Edit_InvalidID(t *testing.T) {
	h := newEventsHandler(&stubEventManager{})
	res := runEventsGET(h, true, "/admin/evenements/{id}/modifier", "/admin/evenements/not-a-number/modifier", h.Edit)
	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestEventsHandler_Edit_EventNotFound(t *testing.T) {
	h := newEventsHandler(&stubEventManager{detailErr: services.ErrEventNotFound})
	res := runEventsGET(h, true, "/admin/evenements/{id}/modifier", "/admin/evenements/99/modifier", h.Edit)
	assert.Equal(t, http.StatusNotFound, res.Code)
}

// --- Update ---

func TestEventsHandler_Update_Success(t *testing.T) {
	h := newEventsHandler(&stubEventManager{})
	res := runEventsPut(h, true, "/admin/evenements/{id}", "/admin/evenements/1",
		"name=Concert+de+printemps&date=2026-06-01&time=20:00&event_type=concert", h.Update)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/evenements", res.Header().Get("Location"))
}

func TestEventsHandler_Update_InvalidID(t *testing.T) {
	h := newEventsHandler(&stubEventManager{})
	res := runEventsPut(h, true, "/admin/evenements/{id}", "/admin/evenements/not-a-number",
		"name=Concert&date=2026-06-01&time=20:00&event_type=concert", h.Update)
	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestEventsHandler_Update_MissingName(t *testing.T) {
	detail := &services.EventDetailDTO{ID: 1, Name: "Concert", EventType: "concert", Datetime: time.Now()}
	h := newEventsHandler(&stubEventManager{detail: detail})
	res := runEventsPut(h, true, "/admin/evenements/{id}", "/admin/evenements/1",
		"name=&date=2026-06-01&time=20:00&event_type=concert", h.Update)
	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
	assert.Contains(t, res.Body.String(), "nom")
}

func TestEventsHandler_Update_InvalidDate(t *testing.T) {
	detail := &services.EventDetailDTO{ID: 1, Name: "Concert", EventType: "concert", Datetime: time.Now()}
	h := newEventsHandler(&stubEventManager{detail: detail})
	res := runEventsPut(h, true, "/admin/evenements/{id}", "/admin/evenements/1",
		"name=Concert&date=not-a-date&time=20:00&event_type=concert", h.Update)
	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
}

func TestEventsHandler_Update_EventNotFound(t *testing.T) {
	h := newEventsHandler(&stubEventManager{updateErr: services.ErrEventNotFound})
	res := runEventsPut(h, true, "/admin/evenements/{id}", "/admin/evenements/1",
		"name=Concert&date=2026-06-01&time=20:00&event_type=concert", h.Update)
	assert.Equal(t, http.StatusNotFound, res.Code)
}

// --- AddField ---

func TestEventsHandler_AddField_Success(t *testing.T) {
	h := newEventsHandler(&stubEventManager{})
	res := runEventsPost(h, true, "/admin/evenements/{id}/champs", "/admin/evenements/1/champs",
		"label=Ma+question&field_type=text&required=false&position=1", h.AddField)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/admin/evenements/1/modifier", res.Header().Get("Location"))
}

func TestEventsHandler_AddField_InvalidID(t *testing.T) {
	h := newEventsHandler(&stubEventManager{})
	res := runEventsPost(h, true, "/admin/evenements/{id}/champs", "/admin/evenements/not-a-number/champs",
		"label=Ma+question&field_type=text", h.AddField)
	assert.Equal(t, http.StatusBadRequest, res.Code)
}

func TestEventsHandler_AddField_FieldOnlyForOther(t *testing.T) {
	h := newEventsHandler(&stubEventManager{addFieldErr: services.ErrFieldOnlyForOther})
	res := runEventsPost(h, true, "/admin/evenements/{id}/champs", "/admin/evenements/1/champs",
		"label=Ma+question&field_type=text", h.AddField)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/admin/evenements/1/modifier", res.Header().Get("Location"))
}

func TestEventsHandler_AddField_ServiceError(t *testing.T) {
	h := newEventsHandler(&stubEventManager{addFieldErr: errors.New("db error")})
	res := runEventsPost(h, true, "/admin/evenements/{id}/champs", "/admin/evenements/1/champs",
		"label=Ma+question&field_type=text", h.AddField)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

// --- EditFieldForm ---

func TestEventsHandler_EditFieldForm_RendersForm(t *testing.T) {
	field := &services.EventFieldDTO{ID: 1, EventID: 1, Label: "Ma question", FieldType: "text", Position: 1}
	h := newEventsHandler(&stubEventManager{field: field})
	res := runEventsGET(h, true,
		"/admin/evenements/{id}/champs/{field_id}/modifier",
		"/admin/evenements/1/champs/1/modifier",
		h.EditFieldForm)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "Ma question")
}

func TestEventsHandler_EditFieldForm_InvalidFieldID(t *testing.T) {
	h := newEventsHandler(&stubEventManager{})
	res := runEventsGET(h, true,
		"/admin/evenements/{id}/champs/{field_id}/modifier",
		"/admin/evenements/1/champs/not-a-number/modifier",
		h.EditFieldForm)
	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestEventsHandler_EditFieldForm_InvalidEventID(t *testing.T) {
	h := newEventsHandler(&stubEventManager{})
	res := runEventsGET(h, true,
		"/admin/evenements/{id}/champs/{field_id}/modifier",
		"/admin/evenements/not-a-number/champs/1/modifier",
		h.EditFieldForm)
	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestEventsHandler_EditFieldForm_FieldNotFound(t *testing.T) {
	h := newEventsHandler(&stubEventManager{fieldErr: services.ErrEventFieldNotFound})
	res := runEventsGET(h, true,
		"/admin/evenements/{id}/champs/{field_id}/modifier",
		"/admin/evenements/1/champs/99/modifier",
		h.EditFieldForm)
	assert.Equal(t, http.StatusNotFound, res.Code)
}

// --- UpdateField ---

func TestEventsHandler_UpdateField_Success(t *testing.T) {
	h := newEventsHandler(&stubEventManager{})
	res := runEventsPut(h, true,
		"/admin/evenements/{id}/champs/{field_id}",
		"/admin/evenements/1/champs/2",
		"label=Ma+question&field_type=text&required=false&position=1",
		h.UpdateField)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Contains(t, res.Header().Get("Location"), "/admin/evenements/1/modifier")
}

func TestEventsHandler_UpdateField_FieldHasResponses(t *testing.T) {
	h := newEventsHandler(&stubEventManager{updateFieldErr: services.ErrFieldHasResponses})
	res := runEventsPut(h, true,
		"/admin/evenements/{id}/champs/{field_id}",
		"/admin/evenements/1/champs/2",
		"label=Ma+question&field_type=text",
		h.UpdateField)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Contains(t, res.Header().Get("Location"), "/admin/evenements/1/modifier")
}

func TestEventsHandler_UpdateField_ServiceError(t *testing.T) {
	h := newEventsHandler(&stubEventManager{updateFieldErr: errors.New("db error")})
	res := runEventsPut(h, true,
		"/admin/evenements/{id}/champs/{field_id}",
		"/admin/evenements/1/champs/2",
		"label=Ma+question&field_type=text",
		h.UpdateField)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

// --- DeleteField ---

func TestEventsHandler_DeleteField_Success(t *testing.T) {
	h := newEventsHandler(&stubEventManager{})
	res := runEventsDelete(h, true,
		"/admin/evenements/{id}/champs/{field_id}",
		"/admin/evenements/1/champs/2",
		"", h.DeleteField)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Contains(t, res.Header().Get("Location"), "/admin/evenements/1/modifier")
}

func TestEventsHandler_DeleteField_FieldHasResponses(t *testing.T) {
	h := newEventsHandler(&stubEventManager{deleteFieldErr: services.ErrFieldHasResponses})
	res := runEventsDelete(h, true,
		"/admin/evenements/{id}/champs/{field_id}",
		"/admin/evenements/1/champs/2",
		"", h.DeleteField)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Contains(t, res.Header().Get("Location"), "/admin/evenements/1/modifier")
}

func TestEventsHandler_DeleteField_ServiceError(t *testing.T) {
	h := newEventsHandler(&stubEventManager{deleteFieldErr: errors.New("db error")})
	res := runEventsDelete(h, true,
		"/admin/evenements/{id}/champs/{field_id}",
		"/admin/evenements/1/champs/2",
		"", h.DeleteField)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}
