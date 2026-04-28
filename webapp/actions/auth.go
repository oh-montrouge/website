package actions

import (
	"encoding/gob"
	"errors"
	"net/http"

	"ohmontrouge/webapp/services"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
)

func init() {
	gob.Register(int64(0))
}

// AuthHandler handles login and logout routes.
type AuthHandler struct {
	Accounts services.AccountAuthenticator
	Sessions services.SessionRepository // nil in unit tests
	DB       *pop.Connection            // nil in unit tests; passed to Sessions.BindAccount
}

func (h AuthHandler) Form(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("auth/login.plush.html"))
}

func (h AuthHandler) Submit(c buffalo.Context) error {
	email := c.Request().FormValue("email")
	password := c.Request().FormValue("password")

	tx := c.Value("tx").(*pop.Connection)
	account, err := h.Accounts.Authenticate(tx, email, password)
	if err != nil {
		if errors.Is(err, services.ErrAccountNotFound) || errors.Is(err, services.ErrInvalidPassword) {
			c.Flash().Add("danger", "Email ou mot de passe incorrect.")
			return c.Redirect(http.StatusSeeOther, "/connexion")
		}
		return err
	}

	session := c.Session()
	session.Set("account_id", account.ID)
	if err := session.Save(); err != nil {
		return err
	}

	// Bind session to account outside Pop's transaction — avoids row-lock
	// contention with pgstore on follow-up requests.
	if h.Sessions != nil {
		if err := h.Sessions.BindAccount(h.DB, session.Session.ID, account.ID); err != nil {
			return err
		}
	}

	return c.Redirect(http.StatusSeeOther, "/tableau-de-bord")
}

func (h AuthHandler) Logout(c buffalo.Context) error {
	session := c.Session()
	session.Session.Options.MaxAge = -1
	if err := session.Save(); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/connexion")
}
