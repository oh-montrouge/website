package actions

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"ohmontrouge/webapp/services"
)

// SeasonsHandler handles admin season management.
type SeasonsHandler struct {
	Seasons services.SeasonManager
}

// Index renders the season list.
func (h SeasonsHandler) Index(c buffalo.Context) error {
	return h.renderIndex(c, "", false)
}

// Create handles season creation form submissions.
func (h SeasonsHandler) Create(c buffalo.Context) error {
	label := strings.TrimSpace(c.Request().FormValue("label"))
	startDateStr := c.Request().FormValue("start_date")
	endDateStr := c.Request().FormValue("end_date")

	if label == "" {
		return h.renderIndex(c, "Le libellé est requis.", true)
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return h.renderIndex(c, "La date de début est invalide.", true)
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return h.renderIndex(c, "La date de fin est invalide.", true)
	}

	if !endDate.After(startDate) {
		return h.renderIndex(c, "La date de fin doit être postérieure à la date de début.", true)
	}

	tx := c.Value("tx").(*pop.Connection)
	if err := h.Seasons.Create(tx, label, startDate, endDate); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/admin/saisons")
}

// DesignateCurrent transfers the current-season designation to the given season.
func (h SeasonsHandler) DesignateCurrent(c buffalo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.Error(http.StatusBadRequest, err)
	}

	tx := c.Value("tx").(*pop.Connection)
	if err := h.Seasons.DesignateCurrent(tx, id); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/admin/saisons")
}

func (h SeasonsHandler) renderIndex(c buffalo.Context, createError string, createOpen bool) error {
	tx := c.Value("tx").(*pop.Connection)
	seasons, err := h.Seasons.List(tx)
	if err != nil {
		return err
	}
	c.Set("seasons", seasons)
	c.Set("seasonsEmpty", len(seasons) == 0)
	c.Set("createError", createError)
	c.Set("createOpen", createOpen)
	status := http.StatusOK
	if createError != "" {
		status = http.StatusUnprocessableEntity
	}
	return c.Render(status, r.HTML("admin/seasons/index.plush.html"))
}
