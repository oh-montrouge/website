package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"ohmontrouge/webapp/services"
)

// ProfileHandler handles the musician's own profile view.
type ProfileHandler struct {
	Membership services.MusicianProfileManager
}

func (h ProfileHandler) Show(c buffalo.Context) error {
	account := c.Value("current_account").(*services.AccountDTO)
	tx := c.Value("tx").(*pop.Connection)

	profile, err := h.Membership.GetProfile(tx, account.ID)
	if err != nil {
		return err
	}

	c.Set("profile", profile)
	c.Set("birthDateStr", safeDateStr(profile.BirthDate))
	c.Set("account", account)
	return c.Render(http.StatusOK, r.HTML("profile/show.plush.html"))
}
