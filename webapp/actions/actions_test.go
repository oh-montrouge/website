package actions

import (
	"net/http"

	"ohmontrouge/webapp/locales"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/middleware/i18n"
	"github.com/gobuffalo/pop/v6"
)

// newTestApp builds a minimal Buffalo app for handler unit tests.
// It registers the i18n and CSRF helpers that production templates require,
// and injects a nil *pop.Connection so handlers can extract "tx" without
// panicking. Stub repositories must not dereference the connection.
func newTestApp(register func(*buffalo.App)) http.Handler {
	a := buffalo.New(buffalo.Options{Env: "test"})

	// Templates use t() — register the translator without the production
	// app.Stop path.
	if translator, err := i18n.New(locales.FS(), "en-US"); err == nil {
		a.Use(translator.Middleware())
	}

	a.Use(func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			c.Set("authenticity_token", "") // layout requires this CSRF placeholder
			c.Set("sheet_music_url", "")    // layout requires this; tests override with injectContextValue
			c.Set("tx", (*pop.Connection)(nil))
			return next(c)
		}
	})

	register(a)
	return a
}
