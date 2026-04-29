package actions

import (
	"database/sql"
	"errors"
	"net/http"
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
	res := serveGET(t, "/protected", sentinel, RequireActiveAccount(stubMiddlewareAuth{}))
	assert.Equal(t, http.StatusFound, res.Code)
	assert.Equal(t, "/connexion", res.Header().Get("Location"))
}

func TestRequireActiveAccount_SessionWrongType_RedirectsToLogin(t *testing.T) {
	res := serveGET(t, "/protected", sentinel,
		injectSession("account_id", "not-an-int64"),
		RequireActiveAccount(stubMiddlewareAuth{}))
	assert.Equal(t, http.StatusFound, res.Code)
	assert.Equal(t, "/connexion", res.Header().Get("Location"))
}

func TestRequireActiveAccount_AccountNotFound_RedirectsToLogin(t *testing.T) {
	res := serveGET(t, "/protected", sentinel,
		injectSession("account_id", int64(42)),
		RequireActiveAccount(stubMiddlewareAuth{err: sql.ErrNoRows}))
	assert.Equal(t, http.StatusFound, res.Code)
	assert.Equal(t, "/connexion", res.Header().Get("Location"))
}

func TestRequireActiveAccount_DBError_Returns500(t *testing.T) {
	res := serveGET(t, "/protected", sentinel,
		injectSession("account_id", int64(42)),
		RequireActiveAccount(stubMiddlewareAuth{err: errors.New("db failure")}))
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestRequireActiveAccount_AccountNotActive_RedirectsToLogin(t *testing.T) {
	res := serveGET(t, "/protected", sentinel,
		injectSession("account_id", int64(42)),
		RequireActiveAccount(stubMiddlewareAuth{account: &services.AccountDTO{ID: 42, Status: services.StatusPending}}))
	assert.Equal(t, http.StatusFound, res.Code)
	assert.Equal(t, "/connexion", res.Header().Get("Location"))
}

func TestRequireActiveAccount_ActiveAccount_PassesThrough(t *testing.T) {
	res := serveGET(t, "/protected", sentinel,
		injectSession("account_id", int64(42)),
		RequireActiveAccount(stubMiddlewareAuth{account: &services.AccountDTO{ID: 42, Status: services.StatusActive}}))
	assert.Equal(t, http.StatusOK, res.Code)
}

// --- RequireAdmin ---

func TestRequireAdmin_NoCurrentAccount_Returns403(t *testing.T) {
	res := serveGET(t, "/admin", sentinel, RequireAdmin(stubMiddlewareAuth{}))
	assert.Equal(t, http.StatusForbidden, res.Code)
}

func TestRequireAdmin_NotAdmin_Returns403(t *testing.T) {
	res := serveGET(t, "/admin", sentinel,
		injectContextValue("current_account", &services.AccountDTO{ID: 42}),
		RequireAdmin(stubMiddlewareAuth{isAdmin: false}))
	assert.Equal(t, http.StatusForbidden, res.Code)
}

func TestRequireAdmin_IsAdminDBError_Returns500(t *testing.T) {
	res := serveGET(t, "/admin", sentinel,
		injectContextValue("current_account", &services.AccountDTO{ID: 42}),
		RequireAdmin(stubMiddlewareAuth{adminErr: errors.New("db failure")}))
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestRequireAdmin_IsAdmin_PassesThrough(t *testing.T) {
	res := serveGET(t, "/admin", sentinel,
		injectContextValue("current_account", &services.AccountDTO{ID: 42}),
		RequireAdmin(stubMiddlewareAuth{isAdmin: true}))
	assert.Equal(t, http.StatusOK, res.Code)
}

// --- LoadCurrentAccount ---

func TestLoadCurrentAccount_NoSession_PassesThroughWithoutAccount(t *testing.T) {
	checked := false
	res := serveGET(t, "/",
		func(c buffalo.Context) error {
			checked = true
			assert.Nil(t, c.Value("current_account"), "current_account should not be set")
			return c.Render(http.StatusOK, r.String("ok"))
		},
		LoadCurrentAccount(stubMiddlewareAuth{}))
	assert.Equal(t, http.StatusOK, res.Code)
	assert.True(t, checked)
}

func TestLoadCurrentAccount_ActiveAccount_SetsCurrentAccount(t *testing.T) {
	res := serveGET(t, "/",
		func(c buffalo.Context) error {
			acct, ok := c.Value("current_account").(*services.AccountDTO)
			assert.True(t, ok)
			assert.Equal(t, int64(42), acct.ID)
			return c.Render(http.StatusOK, r.String("ok"))
		},
		injectSession("account_id", int64(42)),
		LoadCurrentAccount(stubMiddlewareAuth{account: &services.AccountDTO{ID: 42, Status: services.StatusActive}}))
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestLoadCurrentAccount_PendingAccount_DoesNotSetCurrentAccount(t *testing.T) {
	res := serveGET(t, "/",
		func(c buffalo.Context) error {
			assert.Nil(t, c.Value("current_account"), "pending account should not be set as current_account")
			return c.Render(http.StatusOK, r.String("ok"))
		},
		injectSession("account_id", int64(42)),
		LoadCurrentAccount(stubMiddlewareAuth{account: &services.AccountDTO{ID: 42, Status: services.StatusPending}}))
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestLoadCurrentAccount_AdminAccount_SetsIsAdmin(t *testing.T) {
	res := serveGET(t, "/",
		func(c buffalo.Context) error {
			isAdmin, ok := c.Value("is_admin").(bool)
			assert.True(t, ok)
			assert.True(t, isAdmin)
			return c.Render(http.StatusOK, r.String("ok"))
		},
		injectSession("account_id", int64(42)),
		LoadCurrentAccount(stubMiddlewareAuth{account: &services.AccountDTO{ID: 42, Status: services.StatusActive}, isAdmin: true}))
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestLoadCurrentAccount_NonAdminAccount_IsAdminFalse(t *testing.T) {
	res := serveGET(t, "/",
		func(c buffalo.Context) error {
			isAdmin, ok := c.Value("is_admin").(bool)
			assert.True(t, ok)
			assert.False(t, isAdmin)
			return c.Render(http.StatusOK, r.String("ok"))
		},
		injectSession("account_id", int64(42)),
		LoadCurrentAccount(stubMiddlewareAuth{account: &services.AccountDTO{ID: 42, Status: services.StatusActive}, isAdmin: false}))
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestLoadCurrentAccount_AccountNotFound_PassesThrough(t *testing.T) {
	res := serveGET(t, "/", sentinel,
		injectSession("account_id", int64(99)),
		LoadCurrentAccount(stubMiddlewareAuth{err: sql.ErrNoRows}))
	assert.Equal(t, http.StatusOK, res.Code)
}
