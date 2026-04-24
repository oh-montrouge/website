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
	"ohmontrouge/webapp/models"
	"ohmontrouge/webapp/services"
)

// --- stubs ---

type stubAccountAdmin struct {
	account     *services.AccountDTO
	accountErr  error
	isAdmin     bool
	isAdminErr  error
	createdID   int64
	createErr   error
	inviteToken *services.InviteTokenDTO
	inviteErr   error
	resetToken  *services.PasswordResetTokenDTO
	resetErr    error
	grantErr    error
	revokeErr   error
	deleteErr   error
}

func (s *stubAccountAdmin) GetByID(_ *pop.Connection, _ int64) (*services.AccountDTO, error) {
	return s.account, s.accountErr
}

func (s *stubAccountAdmin) IsAdmin(_ *pop.Connection, _ int64) (bool, error) {
	return s.isAdmin, s.isAdminErr
}

func (s *stubAccountAdmin) CreatePending(_ *pop.Connection, _ string, _ int64) (int64, error) {
	return s.createdID, s.createErr
}

func (s *stubAccountAdmin) GenerateInviteToken(_ *pop.Connection, _ int64, _ string) (*services.InviteTokenDTO, error) {
	return s.inviteToken, s.inviteErr
}

func (s *stubAccountAdmin) GetActiveInviteToken(_ *pop.Connection, _ int64, _ string) (*services.InviteTokenDTO, error) {
	return s.inviteToken, s.inviteErr
}

func (s *stubAccountAdmin) GeneratePasswordResetToken(_ *pop.Connection, _ int64, _ string) (*services.PasswordResetTokenDTO, error) {
	return s.resetToken, s.resetErr
}

func (s *stubAccountAdmin) GetActivePasswordResetToken(_ *pop.Connection, _ int64, _ string) (*services.PasswordResetTokenDTO, error) {
	return s.resetToken, s.resetErr
}

func (s *stubAccountAdmin) GrantAdmin(_ *pop.Connection, _ int64) error {
	return s.grantErr
}

func (s *stubAccountAdmin) RevokeAdmin(_ *pop.Connection, _ int64) error {
	return s.revokeErr
}

func (s *stubAccountAdmin) DeletePending(_ *pop.Connection, _ int64) error {
	return s.deleteErr
}

type stubMusicianProfile struct {
	profile       *services.MusicianProfile
	profileErr    error
	summaries     []services.MusicianProfileSummary
	summariesErr  error
	setInitialErr error
	updateErr     error
	consentErr    error
	toggleErr     error
}

func (s *stubMusicianProfile) GetProfile(_ *pop.Connection, _ int64) (*services.MusicianProfile, error) {
	return s.profile, s.profileErr
}

func (s *stubMusicianProfile) SetInitialProfile(_ *pop.Connection, _ int64, _, _ string, _ *time.Time, _ string) error {
	return s.setInitialErr
}

func (s *stubMusicianProfile) UpdateProfile(_ *pop.Connection, _ int64, _, _, _ string, _ int64, _ *time.Time, _, _, _ string) error {
	return s.updateErr
}

func (s *stubMusicianProfile) ListNonAnonymized(_ *pop.Connection) ([]services.MusicianProfileSummary, error) {
	return s.summaries, s.summariesErr
}

func (s *stubMusicianProfile) ConsentWithdrawal(_ *pop.Connection, _ int64) error {
	return s.consentErr
}

func (s *stubMusicianProfile) ToggleProcessingRestriction(_ *pop.Connection, _ int64) error {
	return s.toggleErr
}

type stubCompliance struct {
	anonymizeErr  error
	retentionList []services.RetentionEntryDTO
	retentionErr  error
}

func (s *stubCompliance) Anonymize(_ *pop.Connection, _ int64) error {
	return s.anonymizeErr
}

func (s *stubCompliance) RetentionReviewList(_ *pop.Connection) ([]services.RetentionEntryDTO, error) {
	return s.retentionList, s.retentionErr
}

type stubInstruments struct {
	instruments models.Instruments
	err         error
}

func (s *stubInstruments) List(_ *pop.Connection) (models.Instruments, error) {
	return s.instruments, s.err
}

func defaultInstruments() models.Instruments {
	return models.Instruments{{ID: 1, Name: "Clarinette"}, {ID: 2, Name: "Trompette"}}
}

func newMusiciansTestApp(h MusiciansHandler, register func(*buffalo.App, MusiciansHandler)) http.Handler {
	return newTestApp(func(a *buffalo.App) {
		register(a, h)
	})
}

// --- Index tests ---

func TestMusiciansHandler_Index_ReturnsList(t *testing.T) {
	summaries := []services.MusicianProfileSummary{
		{AccountID: 1, FirstName: "Alice", LastName: "Martin", MainInstrumentName: "Clarinette", Status: "active"},
	}
	h := MusiciansHandler{
		Membership: &stubMusicianProfile{summaries: summaries},
	}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.GET("/admin/musiciens", h.Index)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/musiciens", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "Alice")
	assert.Contains(t, res.Body.String(), "Clarinette")
}

func TestMusiciansHandler_Index_EmptyList(t *testing.T) {
	h := MusiciansHandler{
		Membership: &stubMusicianProfile{summaries: []services.MusicianProfileSummary{}},
	}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.GET("/admin/musiciens", h.Index)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/musiciens", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "Aucun musicien")
}

// --- Show tests ---

func TestMusiciansHandler_Show_RendersProfile(t *testing.T) {
	account := &services.AccountDTO{ID: 1, Email: "alice@example.com", Status: services.StatusActive}
	profile := &services.MusicianProfile{
		AccountID:          1,
		FirstName:          "Alice",
		LastName:           "Martin",
		MainInstrumentName: "Clarinette",
	}
	h := MusiciansHandler{
		Accounts:    &stubAccountAdmin{account: account},
		Membership:  &stubMusicianProfile{profile: profile},
		FeePayments: &stubFeePaymentManager{},
		Seasons:     stubSeasonManager{},
	}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.GET("/admin/musiciens/{id}", h.Show)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/musiciens/1", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "Alice")
	assert.Contains(t, res.Body.String(), "Clarinette")
}

func TestMusiciansHandler_Show_AccountNotFound_Returns404(t *testing.T) {
	h := MusiciansHandler{
		Accounts: &stubAccountAdmin{accountErr: errors.New("not found")},
	}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.GET("/admin/musiciens/{id}", h.Show)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/musiciens/999", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNotFound, res.Code)
}

// --- Create tests ---

func TestMusiciansHandler_Create_Success_Redirects(t *testing.T) {
	inviteToken := &services.InviteTokenDTO{
		Token:     "tok",
		URL:       "http://localhost:3000/invitation/tok",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	h := MusiciansHandler{
		Accounts:    &stubAccountAdmin{createdID: 7, inviteToken: inviteToken},
		Membership:  &stubMusicianProfile{},
		Instruments: &stubInstruments{instruments: defaultInstruments()},
	}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.POST("/admin/musiciens", h.Create)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("first_name=Alice&last_name=Martin&email=alice%40example.com&main_instrument_id=1&birth_date=&parental_consent_uri=")
	req := httptest.NewRequest(http.MethodPost, "/admin/musiciens", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Contains(t, res.Header().Get("Location"), "/admin/musiciens/7")
}

func TestMusiciansHandler_Create_InvalidInstrumentID_ShowsError(t *testing.T) {
	h := MusiciansHandler{
		Accounts:    &stubAccountAdmin{},
		Membership:  &stubMusicianProfile{},
		Instruments: &stubInstruments{instruments: defaultInstruments()},
	}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.POST("/admin/musiciens", h.Create)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("first_name=Alice&last_name=Martin&email=alice%40example.com&main_instrument_id=notanumber")
	req := httptest.NewRequest(http.MethodPost, "/admin/musiciens", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
	assert.Contains(t, res.Body.String(), "Instrument invalide")
}

func TestMusiciansHandler_Create_ParentalConsentRequired_ShowsError(t *testing.T) {
	h := MusiciansHandler{
		Accounts:    &stubAccountAdmin{createdID: 7},
		Membership:  &stubMusicianProfile{setInitialErr: services.ErrParentalConsentRequired},
		Instruments: &stubInstruments{instruments: defaultInstruments()},
	}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.POST("/admin/musiciens", h.Create)
	})

	res := httptest.NewRecorder()
	// Under-15 birth date, no consent URI
	body := strings.NewReader("first_name=Lea&last_name=Martin&email=lea%40example.com&main_instrument_id=1&birth_date=2018-01-01&parental_consent_uri=")
	req := httptest.NewRequest(http.MethodPost, "/admin/musiciens", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
	assert.Contains(t, res.Body.String(), "accord parental")
}

// --- Delete tests ---

func TestMusiciansHandler_Delete_Pending_Redirects(t *testing.T) {
	h := MusiciansHandler{
		Accounts: &stubAccountAdmin{deleteErr: nil},
	}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.DELETE("/admin/musiciens/{id}", h.Delete)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/admin/musiciens/1", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/admin/musiciens", res.Header().Get("Location"))
}

func TestMusiciansHandler_Delete_NotPending_ShowsFlash(t *testing.T) {
	h := MusiciansHandler{
		Accounts: &stubAccountAdmin{deleteErr: services.ErrAccountNotPending},
	}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.DELETE("/admin/musiciens/{id}", h.Delete)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/admin/musiciens/1", nil)
	app.ServeHTTP(res, req)

	// Flash and redirect to show page
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Contains(t, res.Header().Get("Location"), "/admin/musiciens/1")
}

// --- Anonymize tests ---

func TestMusiciansHandler_Anonymize_ConfirmationMissing_DoesNotAnonymize(t *testing.T) {
	compliance := &stubCompliance{}
	h := MusiciansHandler{
		Compliance: compliance,
	}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.POST("/admin/musiciens/{id}/anonymiser", h.Anonymize)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("confirmed=wrong")
	req := httptest.NewRequest(http.MethodPost, "/admin/musiciens/1/anonymiser", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	// Should redirect back with danger flash, not call Anonymize
	assert.Equal(t, http.StatusSeeOther, res.Code)
}

func TestMusiciansHandler_Anonymize_WithConfirmation_Succeeds(t *testing.T) {
	compliance := &stubCompliance{}
	h := MusiciansHandler{
		Compliance: compliance,
	}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.POST("/admin/musiciens/{id}/anonymiser", h.Anonymize)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("confirmed=ANONYMISER")
	req := httptest.NewRequest(http.MethodPost, "/admin/musiciens/1/anonymiser", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Contains(t, res.Header().Get("Location"), "/admin/musiciens/1")
}

// --- Edit tests ---

func TestMusiciansHandler_Edit_RendersForm(t *testing.T) {
	account := &services.AccountDTO{ID: 1, Email: "alice@example.com", Status: services.StatusActive}
	profile := &services.MusicianProfile{
		AccountID:          1,
		FirstName:          "Alice",
		LastName:           "Martin",
		MainInstrumentID:   1,
		MainInstrumentName: "Clarinette",
	}
	h := MusiciansHandler{
		Accounts:    &stubAccountAdmin{account: account},
		Membership:  &stubMusicianProfile{profile: profile},
		Instruments: &stubInstruments{instruments: defaultInstruments()},
	}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.GET("/admin/musiciens/{id}/modifier", h.Edit)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/musiciens/1/modifier", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	body := res.Body.String()
	assert.Contains(t, body, "Alice")
	assert.Contains(t, body, "alice@example.com")
	assert.Contains(t, body, "Clarinette")
}

func TestMusiciansHandler_Edit_AnonymizedAccount_Redirects(t *testing.T) {
	account := &services.AccountDTO{ID: 1, Email: "", Status: services.StatusAnonymized}
	h := MusiciansHandler{
		Accounts: &stubAccountAdmin{account: account},
	}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.GET("/admin/musiciens/{id}/modifier", h.Edit)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/musiciens/1/modifier", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Contains(t, res.Header().Get("Location"), "/admin/musiciens/1")
}

// --- Update tests ---

func TestMusiciansHandler_Update_Success_Redirects(t *testing.T) {
	account := &services.AccountDTO{ID: 1, Email: "alice@example.com", Status: services.StatusActive}
	profile := &services.MusicianProfile{AccountID: 1}
	h := MusiciansHandler{
		Accounts:    &stubAccountAdmin{account: account},
		Membership:  &stubMusicianProfile{profile: profile},
		Instruments: &stubInstruments{instruments: defaultInstruments()},
	}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.PUT("/admin/musiciens/{id}", h.Update)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("first_name=Alice&last_name=Martin&email=alice%40example.com&main_instrument_id=1")
	req := httptest.NewRequest(http.MethodPut, "/admin/musiciens/1", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Contains(t, res.Header().Get("Location"), "/admin/musiciens/1")
}

func TestMusiciansHandler_Update_ParentalConsentRequired_ShowsError(t *testing.T) {
	account := &services.AccountDTO{ID: 1, Email: "lea@example.com", Status: services.StatusActive}
	profile := &services.MusicianProfile{AccountID: 1}
	h := MusiciansHandler{
		Accounts:    &stubAccountAdmin{account: account},
		Membership:  &stubMusicianProfile{profile: profile, updateErr: services.ErrParentalConsentRequired},
		Instruments: &stubInstruments{instruments: defaultInstruments()},
	}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.PUT("/admin/musiciens/{id}", h.Update)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("first_name=Lea&last_name=Martin&email=lea%40example.com&main_instrument_id=1&birth_date=2018-01-01&parental_consent_uri=")
	req := httptest.NewRequest(http.MethodPut, "/admin/musiciens/1", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
	assert.Contains(t, res.Body.String(), "accord parental")
}

// --- GrantAdmin / RevokeAdmin tests ---

func TestMusiciansHandler_GrantAdmin_Success_Redirects(t *testing.T) {
	h := MusiciansHandler{Accounts: &stubAccountAdmin{}}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.POST("/admin/musiciens/{id}/role-admin", h.GrantAdmin)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/musiciens/1/role-admin", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Contains(t, res.Header().Get("Location"), "/admin/musiciens/1")
}

func TestMusiciansHandler_RevokeAdmin_Success_Redirects(t *testing.T) {
	h := MusiciansHandler{Accounts: &stubAccountAdmin{}}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.DELETE("/admin/musiciens/{id}/role-admin", h.RevokeAdmin)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/admin/musiciens/1/role-admin", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Contains(t, res.Header().Get("Location"), "/admin/musiciens/1")
}

func TestMusiciansHandler_RevokeAdmin_LastAdmin_RedirectsWithFlash(t *testing.T) {
	h := MusiciansHandler{Accounts: &stubAccountAdmin{revokeErr: services.ErrLastAdmin}}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.DELETE("/admin/musiciens/{id}/role-admin", h.RevokeAdmin)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/admin/musiciens/1/role-admin", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Contains(t, res.Header().Get("Location"), "/admin/musiciens/1")
}

// --- WithdrawConsent test ---

func TestMusiciansHandler_WithdrawConsent_Success_Redirects(t *testing.T) {
	h := MusiciansHandler{Membership: &stubMusicianProfile{}}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.DELETE("/admin/musiciens/{id}/consentement", h.WithdrawConsent)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/admin/musiciens/1/consentement", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Contains(t, res.Header().Get("Location"), "/admin/musiciens/1")
}

// --- ToggleProcessingRestriction test ---

func TestMusiciansHandler_ToggleProcessingRestriction_Success_Redirects(t *testing.T) {
	h := MusiciansHandler{Membership: &stubMusicianProfile{}}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.POST("/admin/musiciens/{id}/restriction", h.ToggleProcessingRestriction)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/musiciens/1/restriction", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Contains(t, res.Header().Get("Location"), "/admin/musiciens/1")
}

// --- GenerateInviteLink / GenerateResetLink tests ---

func TestMusiciansHandler_GenerateInviteLink_Success_Redirects(t *testing.T) {
	inviteToken := &services.InviteTokenDTO{
		Token:     "tok",
		URL:       "http://localhost:3000/invitation/tok",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	h := MusiciansHandler{Accounts: &stubAccountAdmin{inviteToken: inviteToken}}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.POST("/admin/musiciens/{id}/invitation", h.GenerateInviteLink)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/musiciens/1/invitation", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Contains(t, res.Header().Get("Location"), "/admin/musiciens/1")
}

func TestMusiciansHandler_GenerateResetLink_Success_Redirects(t *testing.T) {
	resetToken := &services.PasswordResetTokenDTO{
		Token:     "rst",
		URL:       "http://localhost:3000/reinitialiser-mot-de-passe/rst",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	h := MusiciansHandler{Accounts: &stubAccountAdmin{resetToken: resetToken}}
	app := newMusiciansTestApp(h, func(a *buffalo.App, h MusiciansHandler) {
		a.POST("/admin/musiciens/{id}/reinitialisation", h.GenerateResetLink)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/musiciens/1/reinitialisation", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Contains(t, res.Header().Get("Location"), "/admin/musiciens/1")
}
