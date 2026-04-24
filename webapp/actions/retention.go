package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"ohmontrouge/webapp/services"
)

// RetentionHandler handles the admin retention review list.
type RetentionHandler struct {
	Compliance services.ComplianceManager
}

func (h RetentionHandler) Index(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	entries, err := h.Compliance.RetentionReviewList(tx)
	if err != nil {
		return err
	}
	c.Set("entries", entries)
	return c.Render(http.StatusOK, r.HTML("admin/retention/index.plush.html"))
}
