package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
)

// HomeHandler serves the public homepage and privacy notice.
type HomeHandler struct{}

func (h HomeHandler) Index(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("home/index.plush.html"))
}

func (h HomeHandler) Privacy(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("privacy/index.plush.html"))
}
