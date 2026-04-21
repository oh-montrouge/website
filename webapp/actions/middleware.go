package actions

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"ohmontrouge/webapp/services"
)

// LoadCurrentAccount is a non-blocking middleware that loads the authenticated
// account into context under "current_account" when a valid active session exists.
// Unlike RequireActiveAccount it does not redirect on failure — it simply skips.
func LoadCurrentAccount(svc services.AccountAuthenticator) buffalo.MiddlewareFunc {
	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			rawID := c.Session().Get("account_id")
			if rawID == nil {
				return next(c)
			}
			accountID, ok := rawID.(int64)
			if !ok {
				return next(c)
			}
			tx := c.Value("tx").(*pop.Connection)
			account, err := svc.GetByID(tx, accountID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return next(c)
				}
				return err
			}
			if account.Status == services.StatusActive {
				c.Set("current_account", account)
			}
			return next(c)
		}
	}
}

// RequireActiveAccount redirects to /connexion if the session carries no valid,
// active account. On success it stores the AccountDTO under "current_account".
func RequireActiveAccount(svc services.AccountAuthenticator) buffalo.MiddlewareFunc {
	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			rawID := c.Session().Get("account_id")
			if rawID == nil {
				return c.Redirect(http.StatusFound, "/connexion")
			}
			accountID, ok := rawID.(int64)
			if !ok {
				return c.Redirect(http.StatusFound, "/connexion")
			}
			tx := c.Value("tx").(*pop.Connection)
			account, err := svc.GetByID(tx, accountID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return c.Redirect(http.StatusFound, "/connexion")
				}
				return err
			}
			if account.Status != services.StatusActive {
				return c.Redirect(http.StatusFound, "/connexion")
			}
			c.Set("current_account", account)
			return next(c)
		}
	}
}

// RequireAdmin returns 403 if the current_account (set by RequireActiveAccount)
// does not hold the admin role.
func RequireAdmin(svc services.AccountAuthenticator) buffalo.MiddlewareFunc {
	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			account, ok := c.Value("current_account").(*services.AccountDTO)
			if !ok {
				return c.Error(http.StatusForbidden, errors.New("forbidden"))
			}
			tx := c.Value("tx").(*pop.Connection)
			isAdmin, err := svc.IsAdmin(tx, account.ID)
			if err != nil {
				return err
			}
			if !isAdmin {
				return c.Error(http.StatusForbidden, errors.New("forbidden"))
			}
			return next(c)
		}
	}
}
