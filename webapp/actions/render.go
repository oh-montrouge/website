package actions

import (
	"time"

	"ohmontrouge/webapp/public"
	"ohmontrouge/webapp/templates"

	"github.com/gobuffalo/buffalo/render"
)

var r *render.Engine

func init() {
	r = render.New(render.Options{
		HTMLLayout:  "layouts/application.plush.html",
		TemplatesFS: templates.FS(),
		AssetsFS:    public.FS(),
		Helpers: render.Helpers{
			"currentYear": func() int { return time.Now().Year() },
		},
	})
}
