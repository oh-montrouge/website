package actions

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"ohmontrouge/webapp/services"
)

// MusiciansHandler handles all admin musician management routes.
type MusiciansHandler struct {
	Accounts    services.AccountAdminManager
	Membership  services.MusicianProfileManager
	Compliance  services.ComplianceManager
	Instruments services.InstrumentRepository
	FeePayments services.FeePaymentManager
	Seasons     services.SeasonManager
	BaseURL     string
}

func (h MusiciansHandler) Index(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	musicians, err := h.Membership.ListNonAnonymized(tx)
	if err != nil {
		return err
	}
	c.Set("musicians", musicians)
	return c.Render(http.StatusOK, r.HTML("admin/musicians/index.plush.html"))
}

func (h MusiciansHandler) New(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	instruments, err := h.Instruments.List(tx)
	if err != nil {
		return err
	}
	c.Set("instruments", instruments)
	c.Set("formError", "")
	return c.Render(http.StatusOK, r.HTML("admin/musicians/new.plush.html"))
}

func (h MusiciansHandler) Create(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	req := c.Request()
	firstName := req.FormValue("first_name")
	lastName := req.FormValue("last_name")
	email := req.FormValue("email")
	instrumentIDStr := req.FormValue("main_instrument_id")
	birthDateStr := req.FormValue("birth_date")
	parentalConsentURI := req.FormValue("parental_consent_uri")

	instrumentID, err := strconv.ParseInt(instrumentIDStr, 10, 64)
	if err != nil {
		return h.renderNewWithError(c, tx, "Instrument invalide.")
	}

	birthDate, err := parseOptionalDate(birthDateStr)
	if err != nil {
		return h.renderNewWithError(c, tx, "Date de naissance invalide.")
	}

	accountID, err := h.Accounts.CreatePending(tx, email, instrumentID)
	if err != nil {
		return h.renderNewWithError(c, tx, "Erreur lors de la création du compte : "+err.Error())
	}

	if err := h.Membership.SetInitialProfile(tx, accountID, firstName, lastName, birthDate, parentalConsentURI); err != nil {
		if errors.Is(err, services.ErrParentalConsentRequired) {
			return h.renderNewWithError(c, tx, err.Error())
		}
		return err
	}

	if _, err := h.Accounts.GenerateInviteToken(tx, accountID, h.BaseURL); err != nil {
		return err
	}

	c.Flash().Add("success", "Compte musicien créé. Copiez le lien d'invitation sur la fiche du musicien.")
	return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", accountID)
}

func (h MusiciansHandler) Show(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	id, err := parseID(c)
	if err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	account, err := h.Accounts.GetByID(tx, id)
	if err != nil {
		return c.Error(http.StatusNotFound, errors.New("musicien introuvable"))
	}

	profile, err := h.Membership.GetProfile(tx, id)
	if err != nil {
		return err
	}

	isAdmin, err := h.Accounts.IsAdmin(tx, id)
	if err != nil {
		return err
	}

	inviteToken, err := h.Accounts.GetActiveInviteToken(tx, id, h.BaseURL)
	if err != nil {
		return err
	}

	resetToken, err := h.Accounts.GetActivePasswordResetToken(tx, id, h.BaseURL)
	if err != nil {
		return err
	}

	feePayments, err := h.FeePayments.ListByAccount(tx, id)
	if err != nil {
		return err
	}

	firstInscriptionDate, err := h.FeePayments.GetFirstInscriptionDate(tx, id)
	if err != nil {
		return err
	}

	seasons, err := h.Seasons.List(tx)
	if err != nil {
		return err
	}

	c.Set("account", account)
	c.Set("accountStatus", string(account.Status))
	c.Set("profile", profile)
	c.Set("birthDateStr", safeDateStr(profile.BirthDate))
	c.Set("isAdmin", isAdmin)
	c.Set("hasInviteToken", inviteToken != nil)
	c.Set("inviteTokenURL", safeTokenURL(inviteToken))
	c.Set("inviteTokenExpiry", safeTokenExpiry(inviteToken))
	c.Set("hasResetToken", resetToken != nil)
	c.Set("resetTokenURL", safeResetURL(resetToken))
	c.Set("resetTokenExpiry", safeResetExpiry(resetToken))
	c.Set("feePayments", feePayments)
	c.Set("firstInscriptionDate", safeDateStr(firstInscriptionDate))
	c.Set("seasons", seasons)
	return c.Render(http.StatusOK, r.HTML("admin/musicians/show.plush.html"))
}

func (h MusiciansHandler) Edit(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	id, err := parseID(c)
	if err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	account, err := h.Accounts.GetByID(tx, id)
	if err != nil {
		return c.Error(http.StatusNotFound, errors.New("musicien introuvable"))
	}
	if account.Status == services.StatusAnonymized {
		c.Flash().Add("danger", "Les comptes anonymisés ne peuvent pas être modifiés.")
		return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", id)
	}

	profile, err := h.Membership.GetProfile(tx, id)
	if err != nil {
		return err
	}

	instruments, err := h.Instruments.List(tx)
	if err != nil {
		return err
	}

	c.Set("account", account)
	c.Set("accountStatus", string(account.Status))
	c.Set("profile", profile)
	c.Set("birthDateStr", safeDateInput(profile.BirthDate))
	c.Set("instruments", instruments)
	c.Set("formError", "")
	return c.Render(http.StatusOK, r.HTML("admin/musicians/edit.plush.html"))
}

func (h MusiciansHandler) Update(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	id, err := parseID(c)
	if err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	profile, err := h.Membership.GetProfile(tx, id)
	if err != nil {
		return err
	}

	req := c.Request()
	firstName := req.FormValue("first_name")
	lastName := req.FormValue("last_name")
	email := req.FormValue("email")
	instrumentIDStr := req.FormValue("main_instrument_id")
	birthDateStr := req.FormValue("birth_date")
	parentalConsentURI := req.FormValue("parental_consent_uri")

	instrumentID, err := strconv.ParseInt(instrumentIDStr, 10, 64)
	if err != nil {
		return h.renderEditWithError(c, tx, id, "Instrument invalide.")
	}

	birthDate, err := parseOptionalDate(birthDateStr)
	if err != nil {
		return h.renderEditWithError(c, tx, id, "Date de naissance invalide.")
	}

	// Only allow editing phone/address if consent is given
	phone := profile.Phone
	address := profile.Address
	if profile.PhoneAddressConsent {
		phone = req.FormValue("phone")
		address = req.FormValue("address")
	}

	if err := h.Membership.UpdateProfile(tx, id, firstName, lastName, email, instrumentID, birthDate, parentalConsentURI, phone, address); err != nil {
		if errors.Is(err, services.ErrParentalConsentRequired) {
			return h.renderEditWithError(c, tx, id, err.Error())
		}
		return h.renderEditWithError(c, tx, id, "Erreur lors de la mise à jour : "+err.Error())
	}

	c.Flash().Add("success", "Profil mis à jour.")
	return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", id)
}

func (h MusiciansHandler) Delete(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	id, err := parseID(c)
	if err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := h.Accounts.DeletePending(tx, id); err != nil {
		if errors.Is(err, services.ErrLastAdmin) || errors.Is(err, services.ErrAccountNotPending) {
			c.Flash().Add("danger", err.Error())
			return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", id)
		}
		return err
	}

	c.Flash().Add("success", "Compte supprimé.")
	return c.Redirect(http.StatusSeeOther, "/admin/musiciens")
}

func (h MusiciansHandler) Anonymize(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	id, err := parseID(c)
	if err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if c.Request().FormValue("confirmed") != "ANONYMISER" {
		c.Flash().Add("danger", "Confirmation requise.")
		return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", id)
	}

	if err := h.Compliance.Anonymize(tx, id); err != nil {
		if errors.Is(err, services.ErrLastAdmin) {
			c.Flash().Add("danger", err.Error())
			return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", id)
		}
		return err
	}

	c.Flash().Add("success", "Compte anonymisé.")
	return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", id)
}

func (h MusiciansHandler) GenerateInviteLink(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	id, err := parseID(c)
	if err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if _, err := h.Accounts.GenerateInviteToken(tx, id, h.BaseURL); err != nil {
		return err
	}

	c.Flash().Add("success", "Nouveau lien d'invitation généré.")
	return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", id)
}

func (h MusiciansHandler) GenerateResetLink(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	id, err := parseID(c)
	if err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if _, err := h.Accounts.GeneratePasswordResetToken(tx, id, h.BaseURL); err != nil {
		return err
	}

	c.Flash().Add("success", "Nouveau lien de réinitialisation généré.")
	return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", id)
}

func (h MusiciansHandler) GrantAdmin(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	id, err := parseID(c)
	if err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := h.Accounts.GrantAdmin(tx, id); err != nil {
		c.Flash().Add("danger", "Erreur lors de l'attribution du rôle admin.")
		return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", id)
	}

	c.Flash().Add("success", "Rôle administrateur accordé.")
	return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", id)
}

func (h MusiciansHandler) RevokeAdmin(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	id, err := parseID(c)
	if err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := h.Accounts.RevokeAdmin(tx, id); err != nil {
		if errors.Is(err, services.ErrLastAdmin) {
			c.Flash().Add("danger", err.Error())
			return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", id)
		}
		return err
	}

	c.Flash().Add("success", "Rôle administrateur révoqué.")
	return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", id)
}

func (h MusiciansHandler) WithdrawConsent(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	id, err := parseID(c)
	if err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := h.Membership.ConsentWithdrawal(tx, id); err != nil {
		return err
	}

	c.Flash().Add("success", "Consentement retiré. Téléphone et adresse effacés.")
	return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", id)
}

func (h MusiciansHandler) ToggleProcessingRestriction(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	id, err := parseID(c)
	if err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := h.Membership.ToggleProcessingRestriction(tx, id); err != nil {
		return err
	}

	c.Flash().Add("success", "Restriction de traitement mise à jour.")
	return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", id)
}

func (h MusiciansHandler) renderNewWithError(c buffalo.Context, tx *pop.Connection, msg string) error {
	instruments, err := h.Instruments.List(tx)
	if err != nil {
		return err
	}
	c.Set("instruments", instruments)
	c.Set("formError", msg)
	return c.Render(http.StatusUnprocessableEntity, r.HTML("admin/musicians/new.plush.html"))
}

func (h MusiciansHandler) renderEditWithError(c buffalo.Context, tx *pop.Connection, id int64, msg string) error {
	account, err := h.Accounts.GetByID(tx, id)
	if err != nil {
		return err
	}
	profile, err := h.Membership.GetProfile(tx, id)
	if err != nil {
		return err
	}
	instruments, err := h.Instruments.List(tx)
	if err != nil {
		return err
	}
	c.Set("account", account)
	c.Set("accountStatus", string(account.Status))
	c.Set("profile", profile)
	c.Set("birthDateStr", safeDateInput(profile.BirthDate))
	c.Set("instruments", instruments)
	c.Set("formError", msg)
	return c.Render(http.StatusUnprocessableEntity, r.HTML("admin/musicians/edit.plush.html"))
}

// safeDateStr formats a date for display (DD/MM/YYYY).
func safeDateStr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("02/01/2006")
}

// safeDateInput formats a date for an HTML date input (YYYY-MM-DD).
func safeDateInput(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func safeTokenURL(tok *services.InviteTokenDTO) string {
	if tok == nil {
		return ""
	}
	return tok.URL
}

func safeTokenExpiry(tok *services.InviteTokenDTO) string {
	if tok == nil {
		return ""
	}
	return tok.ExpiresAt.Format("02/01/2006 à 15:04")
}

func safeResetURL(tok *services.PasswordResetTokenDTO) string {
	if tok == nil {
		return ""
	}
	return tok.URL
}

func safeResetExpiry(tok *services.PasswordResetTokenDTO) string {
	if tok == nil {
		return ""
	}
	return tok.ExpiresAt.Format("02/01/2006 à 15:04")
}

func parseID(c buffalo.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

func parseOptionalDate(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
