package actions

import (
	"errors"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"ohmontrouge/webapp/services"
)

// TokensHandler handles the public invite and password reset flows.
type TokensHandler struct {
	Accounts services.AccountTokenManager
	Sessions services.SessionRepository // nil in unit tests
	DB       *pop.Connection            // nil in unit tests; used by Sessions.BindAccount
}

// InviteForm renders the invite form for a valid token, or the invalid-token page.
func (h TokensHandler) InviteForm(c buffalo.Context) error {
	token := c.Param("token")
	tx := c.Value("tx").(*pop.Connection)

	ctx, err := h.Accounts.ValidateInviteToken(tx, token)
	if err != nil {
		if errors.Is(err, services.ErrInvalidToken) {
			return c.Render(http.StatusOK, r.HTML("tokens/invite_invalid.plush.html"))
		}
		return err
	}
	c.Set("invite", ctx)
	c.Set("formAction", "/invitation/"+token)
	c.Set("passwordError", "")
	return c.Render(http.StatusOK, r.HTML("tokens/invite.plush.html"))
}

// InviteSubmit processes the invite form POST.
func (h TokensHandler) InviteSubmit(c buffalo.Context) error {
	token := c.Param("token")
	tx := c.Value("tx").(*pop.Connection)

	ctx, err := h.Accounts.ValidateInviteToken(tx, token)
	if err != nil {
		if errors.Is(err, services.ErrInvalidToken) {
			return c.Render(http.StatusOK, r.HTML("tokens/invite_invalid.plush.html"))
		}
		return err
	}

	password := c.Request().FormValue("password")
	confirm := c.Request().FormValue("password_confirm")

	if err := services.ValidatePasswordStrength(password, confirm); err != nil {
		c.Set("invite", ctx)
		c.Set("formAction", "/invitation/"+token)
		c.Set("passwordError", err.Error())
		return c.Render(http.StatusUnprocessableEntity, r.HTML("tokens/invite.plush.html"))
	}

	if c.Request().FormValue("privacy_consent") != "1" {
		c.Set("invite", ctx)
		c.Set("formAction", "/invitation/"+token)
		c.Set("passwordError", "Vous devez accepter la politique de confidentialité.")
		return c.Render(http.StatusUnprocessableEntity, r.HTML("tokens/invite.plush.html"))
	}

	hash, err := services.HashPassword(password)
	if err != nil {
		return err
	}

	phoneAddressConsent := c.Request().FormValue("phone_address_consent") == "1"
	if err := h.Accounts.CompleteInvite(tx, ctx.TokenID, ctx.AccountID, hash, phoneAddressConsent); err != nil {
		return err
	}

	session := c.Session()
	session.Set("account_id", ctx.AccountID)
	if err := session.Save(); err != nil {
		return err
	}
	if h.Sessions != nil {
		if err := h.Sessions.BindAccount(h.DB, session.Session.ID, ctx.AccountID); err != nil {
			return err
		}
	}

	return c.Redirect(http.StatusSeeOther, "/evenements")
}

// ResetForm renders the password reset form for a valid token, or the invalid-token page.
func (h TokensHandler) ResetForm(c buffalo.Context) error {
	token := c.Param("token")
	tx := c.Value("tx").(*pop.Connection)

	_, err := h.Accounts.ValidatePasswordResetToken(tx, token)
	if err != nil {
		if errors.Is(err, services.ErrInvalidToken) {
			return c.Render(http.StatusOK, r.HTML("tokens/reset_invalid.plush.html"))
		}
		return err
	}
	c.Set("formAction", "/reinitialiser-mot-de-passe/"+token)
	c.Set("passwordError", "")
	return c.Render(http.StatusOK, r.HTML("tokens/reset.plush.html"))
}

// ResetSubmit processes the password reset form POST.
func (h TokensHandler) ResetSubmit(c buffalo.Context) error {
	token := c.Param("token")
	tx := c.Value("tx").(*pop.Connection)

	ctx, err := h.Accounts.ValidatePasswordResetToken(tx, token)
	if err != nil {
		if errors.Is(err, services.ErrInvalidToken) {
			return c.Render(http.StatusOK, r.HTML("tokens/reset_invalid.plush.html"))
		}
		return err
	}

	password := c.Request().FormValue("password")
	confirm := c.Request().FormValue("password_confirm")

	if err := services.ValidatePasswordStrength(password, confirm); err != nil {
		c.Set("formAction", "/reinitialiser-mot-de-passe/"+token)
		c.Set("passwordError", err.Error())
		return c.Render(http.StatusUnprocessableEntity, r.HTML("tokens/reset.plush.html"))
	}

	hash, err := services.HashPassword(password)
	if err != nil {
		return err
	}

	if err := h.Accounts.CompletePasswordReset(tx, ctx.TokenID, ctx.AccountID, hash); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/connexion")
}
