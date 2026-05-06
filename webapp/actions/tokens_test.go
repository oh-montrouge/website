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

func runInvitePost(t *testing.T, accounts stubTokenManager, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	h := TokensHandler{Accounts: accounts}
	app := newTokensTestApp(h, func(a *buffalo.App, h TokensHandler) {
		a.POST("/invitation/{token}", h.InviteSubmit)
	})
	req := httptest.NewRequest(http.MethodPost, "/invitation/"+token, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	return res
}

func runResetPost(t *testing.T, accounts stubTokenManager, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	h := TokensHandler{Accounts: accounts}
	app := newTokensTestApp(h, func(a *buffalo.App, h TokensHandler) {
		a.POST("/reinitialiser-mot-de-passe/{token}", h.ResetSubmit)
	})
	req := httptest.NewRequest(http.MethodPost, "/reinitialiser-mot-de-passe/"+token, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	return res
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
	res := runInvitePost(t, stubTokenManager{inviteCtx: validInviteCtx()}, "goodtoken",
		"password=ValidPassword1!ExtraLong&password_confirm=ValidPassword1!ExtraLong&privacy_consent=1")
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/evenements", res.Header().Get("Location"))
}

func TestTokensHandler_InviteSubmit_InvalidToken(t *testing.T) {
	res := runInvitePost(t, stubTokenManager{inviteErr: services.ErrInvalidToken}, "badtoken",
		"password=ValidPassword1!ExtraLong&password_confirm=ValidPassword1!ExtraLong")
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "expiré")
}

func TestTokensHandler_InviteSubmit_PasswordMismatch(t *testing.T) {
	res := runInvitePost(t, stubTokenManager{inviteCtx: validInviteCtx()}, "goodtoken",
		"password=ValidPassword1!ExtraLong&password_confirm=Different1!Password")
	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
	assert.Contains(t, res.Body.String(), "correspondent pas")
	assert.Contains(t, res.Body.String(), "Alice")
}

func TestTokensHandler_InviteSubmit_MissingPrivacyConsent(t *testing.T) {
	res := runInvitePost(t, stubTokenManager{inviteCtx: validInviteCtx()}, "goodtoken",
		"password=ValidPassword1!ExtraLong&password_confirm=ValidPassword1!ExtraLong")
	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
	assert.Contains(t, res.Body.String(), "politique de confidentialité")
	assert.Contains(t, res.Body.String(), "Alice")
}

func TestTokensHandler_InviteSubmit_WeakPassword(t *testing.T) {
	res := runInvitePost(t, stubTokenManager{inviteCtx: validInviteCtx()}, "goodtoken",
		"password=short&password_confirm=short")
	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
	assert.Contains(t, res.Body.String(), "Alice")
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
	res := runResetPost(t, stubTokenManager{resetCtx: &services.PasswordResetContextDTO{TokenID: 1, AccountID: 5}}, "goodtoken",
		"password=ValidPassword1!ExtraLong&password_confirm=ValidPassword1!ExtraLong")
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/connexion", res.Header().Get("Location"))
}

func TestTokensHandler_ResetSubmit_InvalidToken(t *testing.T) {
	res := runResetPost(t, stubTokenManager{resetErr: services.ErrInvalidToken}, "badtoken",
		"password=ValidPassword1!ExtraLong&password_confirm=ValidPassword1!ExtraLong")
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "expiré")
}

func TestTokensHandler_ResetSubmit_PasswordMismatch(t *testing.T) {
	res := runResetPost(t, stubTokenManager{resetCtx: &services.PasswordResetContextDTO{TokenID: 1, AccountID: 5}}, "badtoken",
		"password=ValidPassword1!ExtraLong&password_confirm=Different1!Password")
	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
	assert.Contains(t, res.Body.String(), "correspondent pas")
}

func TestTokensHandler_ResetForm_DBError(t *testing.T) {
	h := TokensHandler{Accounts: stubTokenManager{resetErr: errors.New("db error")}}
	app := newTokensTestApp(h, func(a *buffalo.App, h TokensHandler) {
		a.GET("/reinitialiser-mot-de-passe/{token}", h.ResetForm)
	})

	res := httptest.NewRecorder()
	app.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/reinitialiser-mot-de-passe/tok", nil))

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestTokensHandler_ResetSubmit_WeakPassword(t *testing.T) {
	res := runResetPost(t, stubTokenManager{resetCtx: &services.PasswordResetContextDTO{TokenID: 1, AccountID: 5}}, "goodtoken",
		"password=short&password_confirm=short")
	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
	assert.Contains(t, res.Body.String(), "caractères")
}

func TestTokensHandler_ResetSubmit_DBError(t *testing.T) {
	res := runResetPost(t, stubTokenManager{resetErr: errors.New("db error")}, "badtoken",
		"password=ValidPassword1!ExtraLong&password_confirm=ValidPassword1!ExtraLong")
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestTokensHandler_ResetSubmit_CompleteError(t *testing.T) {
	res := runResetPost(t, stubTokenManager{
		resetCtx:    &services.PasswordResetContextDTO{TokenID: 1, AccountID: 5},
		completeErr: errors.New("db"),
	}, "goodtoken", "password=ValidPassword1!ExtraLong&password_confirm=ValidPassword1!ExtraLong")
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestTokensHandler_InviteSubmit_CompleteError(t *testing.T) {
	res := runInvitePost(t, stubTokenManager{
		inviteCtx:   validInviteCtx(),
		completeErr: errors.New("db"),
	}, "goodtoken", "password=ValidPassword1!ExtraLong&password_confirm=ValidPassword1!ExtraLong&privacy_consent=1")
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}
