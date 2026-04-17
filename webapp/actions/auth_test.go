package actions

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/stretchr/testify/assert"
	"ohmontrouge/webapp/services"
)

type stubAuth struct {
	account *services.AccountDTO
	err     error
}

func (s stubAuth) Authenticate(_ *pop.Connection, _, _ string) (*services.AccountDTO, error) {
	return s.account, s.err
}

func (s stubAuth) GetByID(_ *pop.Connection, _ int64) (*services.AccountDTO, error) {
	return s.account, s.err
}

func (s stubAuth) IsAdmin(_ *pop.Connection, _ int64) (bool, error) {
	return false, nil
}

func TestAuthHandler_Form_ReturnsLoginPage(t *testing.T) {
	h := AuthHandler{Accounts: stubAuth{}}
	app := newTestApp(func(a *buffalo.App) {
		a.GET("/connexion", h.Form)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/connexion", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "connexion")
}

func TestAuthHandler_Submit_InvalidCredentials_RedirectsToLogin(t *testing.T) {
	h := AuthHandler{Accounts: stubAuth{err: services.ErrInvalidPassword}}
	app := newTestApp(func(a *buffalo.App) {
		a.POST("/connexion", h.Submit)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("email=user%40example.com&password=wrong")
	req := httptest.NewRequest(http.MethodPost, "/connexion", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/connexion", res.Header().Get("Location"))
}

func TestAuthHandler_Submit_DBError_Returns500(t *testing.T) {
	h := AuthHandler{Accounts: stubAuth{err: errors.New("db failure")}}
	app := newTestApp(func(a *buffalo.App) {
		a.POST("/connexion", h.Submit)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("email=user%40example.com&password=pass")
	req := httptest.NewRequest(http.MethodPost, "/connexion", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestAuthHandler_Submit_Success_RedirectsToHome(t *testing.T) {
	account := &services.AccountDTO{ID: 1, Email: "admin@example.com", Status: services.StatusActive}
	h := AuthHandler{Accounts: stubAuth{account: account}}
	app := newTestApp(func(a *buffalo.App) {
		a.POST("/connexion", h.Submit)
	})

	res := httptest.NewRecorder()
	body := strings.NewReader("email=admin%40example.com&password=secret")
	req := httptest.NewRequest(http.MethodPost, "/connexion", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/", res.Header().Get("Location"))
}

func TestAuthHandler_Logout_RedirectsToLogin(t *testing.T) {
	h := AuthHandler{Accounts: stubAuth{}}
	app := newTestApp(func(a *buffalo.App) {
		a.POST("/deconnexion", h.Logout)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/deconnexion", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/connexion", res.Header().Get("Location"))
}
