package actions

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/stretchr/testify/assert"
	"ohmontrouge/webapp/services"
)

// stubMiddlewareAuth allows controlling IsAdmin independently of account/err,
// which stubAuth (auth_test.go) does not support.
type stubMiddlewareAuth struct {
	account  *services.AccountDTO
	err      error
	isAdmin  bool
	adminErr error
}

func (s stubMiddlewareAuth) Authenticate(_ *pop.Connection, _, _ string) (*services.AccountDTO, error) {
	return s.account, s.err
}
func (s stubMiddlewareAuth) GetByID(_ *pop.Connection, _ int64) (*services.AccountDTO, error) {
	return s.account, s.err
}
func (s stubMiddlewareAuth) IsAdmin(_ *pop.Connection, _ int64) (bool, error) {
	return s.isAdmin, s.adminErr
}

// injectSession injects a session value before the rest of the middleware chain runs.
func injectSession(key string, value any) buffalo.MiddlewareFunc {
	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			c.Session().Set(key, value)
			return next(c)
		}
	}
}

// injectContextValue injects a context value before the rest of the middleware chain runs.
func injectContextValue(key string, value any) buffalo.MiddlewareFunc {
	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			c.Set(key, value)
			return next(c)
		}
	}
}

// sentinel is a terminal handler that returns 200 to confirm the middleware chain passed through.
func sentinel(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.String("ok"))
}

// --- RequireActiveAccount ---

func TestRequireActiveAccount_NoSession_RedirectsToLogin(t *testing.T) {
	app := newTestApp(func(a *buffalo.App) {
		a.Use(RequireActiveAccount(stubMiddlewareAuth{}))
		a.GET("/protected", sentinel)
	})

	res := httptest.NewRecorder()
	app.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/protected", nil))

	assert.Equal(t, http.StatusFound, res.Code)
	assert.Equal(t, "/connexion", res.Header().Get("Location"))
}

func TestRequireActiveAccount_SessionWrongType_RedirectsToLogin(t *testing.T) {
	app := newTestApp(func(a *buffalo.App) {
		a.Use(injectSession("account_id", "not-an-int64"))
		a.Use(RequireActiveAccount(stubMiddlewareAuth{}))
		a.GET("/protected", sentinel)
	})

	res := httptest.NewRecorder()
	app.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/protected", nil))

	assert.Equal(t, http.StatusFound, res.Code)
	assert.Equal(t, "/connexion", res.Header().Get("Location"))
}

func TestRequireActiveAccount_AccountNotFound_RedirectsToLogin(t *testing.T) {
	svc := stubMiddlewareAuth{err: sql.ErrNoRows}
	app := newTestApp(func(a *buffalo.App) {
		a.Use(injectSession("account_id", int64(42)))
		a.Use(RequireActiveAccount(svc))
		a.GET("/protected", sentinel)
	})

	res := httptest.NewRecorder()
	app.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/protected", nil))

	assert.Equal(t, http.StatusFound, res.Code)
	assert.Equal(t, "/connexion", res.Header().Get("Location"))
}

func TestRequireActiveAccount_DBError_Returns500(t *testing.T) {
	svc := stubMiddlewareAuth{err: errors.New("db failure")}
	app := newTestApp(func(a *buffalo.App) {
		a.Use(injectSession("account_id", int64(42)))
		a.Use(RequireActiveAccount(svc))
		a.GET("/protected", sentinel)
	})

	res := httptest.NewRecorder()
	app.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/protected", nil))

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestRequireActiveAccount_AccountNotActive_RedirectsToLogin(t *testing.T) {
	svc := stubMiddlewareAuth{account: &services.AccountDTO{ID: 42, Status: services.StatusPending}}
	app := newTestApp(func(a *buffalo.App) {
		a.Use(injectSession("account_id", int64(42)))
		a.Use(RequireActiveAccount(svc))
		a.GET("/protected", sentinel)
	})

	res := httptest.NewRecorder()
	app.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/protected", nil))

	assert.Equal(t, http.StatusFound, res.Code)
	assert.Equal(t, "/connexion", res.Header().Get("Location"))
}

func TestRequireActiveAccount_ActiveAccount_PassesThrough(t *testing.T) {
	svc := stubMiddlewareAuth{account: &services.AccountDTO{ID: 42, Status: services.StatusActive}}
	app := newTestApp(func(a *buffalo.App) {
		a.Use(injectSession("account_id", int64(42)))
		a.Use(RequireActiveAccount(svc))
		a.GET("/protected", sentinel)
	})

	res := httptest.NewRecorder()
	app.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/protected", nil))

	assert.Equal(t, http.StatusOK, res.Code)
}

// --- RequireAdmin ---

func TestRequireAdmin_NoCurrentAccount_Returns403(t *testing.T) {
	app := newTestApp(func(a *buffalo.App) {
		a.Use(RequireAdmin(stubMiddlewareAuth{}))
		a.GET("/admin", sentinel)
	})

	res := httptest.NewRecorder()
	app.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/admin", nil))

	assert.Equal(t, http.StatusForbidden, res.Code)
}

func TestRequireAdmin_NotAdmin_Returns403(t *testing.T) {
	svc := stubMiddlewareAuth{isAdmin: false}
	app := newTestApp(func(a *buffalo.App) {
		a.Use(injectContextValue("current_account", &services.AccountDTO{ID: 42}))
		a.Use(RequireAdmin(svc))
		a.GET("/admin", sentinel)
	})

	res := httptest.NewRecorder()
	app.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/admin", nil))

	assert.Equal(t, http.StatusForbidden, res.Code)
}

func TestRequireAdmin_IsAdminDBError_Returns500(t *testing.T) {
	svc := stubMiddlewareAuth{adminErr: errors.New("db failure")}
	app := newTestApp(func(a *buffalo.App) {
		a.Use(injectContextValue("current_account", &services.AccountDTO{ID: 42}))
		a.Use(RequireAdmin(svc))
		a.GET("/admin", sentinel)
	})

	res := httptest.NewRecorder()
	app.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/admin", nil))

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestRequireAdmin_IsAdmin_PassesThrough(t *testing.T) {
	svc := stubMiddlewareAuth{isAdmin: true}
	app := newTestApp(func(a *buffalo.App) {
		a.Use(injectContextValue("current_account", &services.AccountDTO{ID: 42}))
		a.Use(RequireAdmin(svc))
		a.GET("/admin", sentinel)
	})

	res := httptest.NewRecorder()
	app.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/admin", nil))

	assert.Equal(t, http.StatusOK, res.Code)
}
