package actions

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ohmontrouge/webapp/services"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/stretchr/testify/assert"
)

// stubTokenManager is a stub AccountTokenManager for handler unit tests.
type stubTokenManager struct {
	inviteCtx   *services.InviteContextDTO
	inviteErr   error
	resetCtx    *services.PasswordResetContextDTO
	resetErr    error
	completeErr error
}

func (s stubTokenManager) ValidateInviteToken(_ *pop.Connection, _ string) (*services.InviteContextDTO, error) {
	return s.inviteCtx, s.inviteErr
}

func (s stubTokenManager) CompleteInvite(_ *pop.Connection, _, _ int64, _ string, _ bool) error {
	return s.completeErr
}

func (s stubTokenManager) ValidatePasswordResetToken(_ *pop.Connection, _ string) (*services.PasswordResetContextDTO, error) {
	return s.resetCtx, s.resetErr
}

func (s stubTokenManager) CompletePasswordReset(_ *pop.Connection, _, _ int64, _ string) error {
	return s.completeErr
}

func validInviteCtx() *services.InviteContextDTO {
	return &services.InviteContextDTO{
		TokenID:        1,
		AccountID:      5,
		FirstName:      "Alice",
		LastName:       "Dupont",
		Email:          "alice@example.com",
		InstrumentName: "Clarinette",
	}
}

func newTokensTestApp(h TokensHandler, register func(*buffalo.App, TokensHandler)) http.Handler {
	return newTestApp(func(a *buffalo.App) {
		register(a, h)
	})
}

// --- invite GET ---

func TestTokensHandler_InviteForm_ValidToken(t *testing.T) {
	h := TokensHandler{Accounts: stubTokenManager{inviteCtx: validInviteCtx()}}
	app := newTokensTestApp(h, func(a *buffalo.App, h TokensHandler) {
		a.GET("/invitation/{token}", h.InviteForm)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/invitation/sometoken", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "Alice")
	assert.Contains(t, res.Body.String(), "Clarinette")
}

func TestTokensHandler_InviteForm_InvalidToken(t *testing.T) {
	h := TokensHandler{Accounts: stubTokenManager{inviteErr: services.ErrInvalidToken}}
	app := newTokensTestApp(h, func(a *buffalo.App, h TokensHandler) {
		a.GET("/invitation/{token}", h.InviteForm)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/invitation/badtoken", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "expiré")
}

func TestTokensHandler_InviteForm_DBError(t *testing.T) {
	h := TokensHandler{Accounts: stubTokenManager{inviteErr: errors.New("db error")}}
	app := newTokensTestApp(h, func(a *buffalo.App, h TokensHandler) {
		a.GET("/invitation/{token}", h.InviteForm)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/invitation/tok", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

// --- invite POST ---

func TestTokensHandler_InviteSubmit_Success(t *testing.T) {
	h := TokensHandler{Accounts: stubTokenManager{inviteCtx: validInviteCtx()}}
	app := newTokensTestApp(h, func(a *buffalo.App, h TokensHandler) {
		a.POST("/invitation/{token}", h.InviteSubmit)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("password=ValidPassword1!ExtraLong&password_confirm=ValidPassword1!ExtraLong&privacy_consent=1")
	req := httptest.NewRequest(http.MethodPost, "/invitation/goodtoken", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/evenements", res.Header().Get("Location"))
}

func TestTokensHandler_InviteSubmit_InvalidToken(t *testing.T) {
	h := TokensHandler{Accounts: stubTokenManager{inviteErr: services.ErrInvalidToken}}
	app := newTokensTestApp(h, func(a *buffalo.App, h TokensHandler) {
		a.POST("/invitation/{token}", h.InviteSubmit)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("password=ValidPassword1!ExtraLong&password_confirm=ValidPassword1!ExtraLong")
	req := httptest.NewRequest(http.MethodPost, "/invitation/badtoken", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "expiré")
}

func TestTokensHandler_InviteSubmit_PasswordMismatch(t *testing.T) {
	h := TokensHandler{Accounts: stubTokenManager{inviteCtx: validInviteCtx()}}
	app := newTokensTestApp(h, func(a *buffalo.App, h TokensHandler) {
		a.POST("/invitation/{token}", h.InviteSubmit)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("password=ValidPassword1!ExtraLong&password_confirm=Different1!Password")
	req := httptest.NewRequest(http.MethodPost, "/invitation/goodtoken", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
	assert.Contains(t, res.Body.String(), "correspondent pas")
	assert.Contains(t, res.Body.String(), "Alice") // account info card re-rendered
}

func TestTokensHandler_InviteSubmit_MissingPrivacyConsent(t *testing.T) {
	h := TokensHandler{Accounts: stubTokenManager{inviteCtx: validInviteCtx()}}
	app := newTokensTestApp(h, func(a *buffalo.App, h TokensHandler) {
		a.POST("/invitation/{token}", h.InviteSubmit)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("password=ValidPassword1!ExtraLong&password_confirm=ValidPassword1!ExtraLong")
	req := httptest.NewRequest(http.MethodPost, "/invitation/goodtoken", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
	assert.Contains(t, res.Body.String(), "politique de confidentialité")
	assert.Contains(t, res.Body.String(), "Alice") // account info card re-rendered
}

func TestTokensHandler_InviteSubmit_WeakPassword(t *testing.T) {
	h := TokensHandler{Accounts: stubTokenManager{inviteCtx: validInviteCtx()}}
	app := newTokensTestApp(h, func(a *buffalo.App, h TokensHandler) {
		a.POST("/invitation/{token}", h.InviteSubmit)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("password=short&password_confirm=short")
	req := httptest.NewRequest(http.MethodPost, "/invitation/goodtoken", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
	assert.Contains(t, res.Body.String(), "Alice") // re-rendered form with account info
}

// --- reset GET ---

func TestTokensHandler_ResetForm_ValidToken(t *testing.T) {
	h := TokensHandler{Accounts: stubTokenManager{resetCtx: &services.PasswordResetContextDTO{TokenID: 1, AccountID: 5}}}
	app := newTokensTestApp(h, func(a *buffalo.App, h TokensHandler) {
		a.GET("/reinitialiser-mot-de-passe/{token}", h.ResetForm)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/reinitialiser-mot-de-passe/goodtoken", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "Réinitialiser")
}

func TestTokensHandler_ResetForm_InvalidToken(t *testing.T) {
	h := TokensHandler{Accounts: stubTokenManager{resetErr: services.ErrInvalidToken}}
	app := newTokensTestApp(h, func(a *buffalo.App, h TokensHandler) {
		a.GET("/reinitialiser-mot-de-passe/{token}", h.ResetForm)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/reinitialiser-mot-de-passe/badtoken", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "expiré")
}

// --- reset POST ---

func TestTokensHandler_ResetSubmit_Success(t *testing.T) {
	h := TokensHandler{Accounts: stubTokenManager{resetCtx: &services.PasswordResetContextDTO{TokenID: 1, AccountID: 5}}}
	app := newTokensTestApp(h, func(a *buffalo.App, h TokensHandler) {
		a.POST("/reinitialiser-mot-de-passe/{token}", h.ResetSubmit)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("password=ValidPassword1!ExtraLong&password_confirm=ValidPassword1!ExtraLong")
	req := httptest.NewRequest(http.MethodPost, "/reinitialiser-mot-de-passe/goodtoken", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/connexion", res.Header().Get("Location"))
}

func TestTokensHandler_ResetSubmit_InvalidToken(t *testing.T) {
	h := TokensHandler{Accounts: stubTokenManager{resetErr: services.ErrInvalidToken}}
	app := newTokensTestApp(h, func(a *buffalo.App, h TokensHandler) {
		a.POST("/reinitialiser-mot-de-passe/{token}", h.ResetSubmit)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("password=ValidPassword1!ExtraLong&password_confirm=ValidPassword1!ExtraLong")
	req := httptest.NewRequest(http.MethodPost, "/reinitialiser-mot-de-passe/badtoken", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "expiré")
}

func TestTokensHandler_ResetSubmit_PasswordMismatch(t *testing.T) {
	h := TokensHandler{Accounts: stubTokenManager{resetCtx: &services.PasswordResetContextDTO{TokenID: 1, AccountID: 5}}}
	app := newTokensTestApp(h, func(a *buffalo.App, h TokensHandler) {
		a.POST("/reinitialiser-mot-de-passe/{token}", h.ResetSubmit)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("password=ValidPassword1!ExtraLong&password_confirm=Different1!Password")
	req := httptest.NewRequest(http.MethodPost, "/reinitialiser-mot-de-passe/badtoken", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
	assert.Contains(t, res.Body.String(), "correspondent pas")
}
