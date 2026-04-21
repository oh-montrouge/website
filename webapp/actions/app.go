package actions

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"ohmontrouge/webapp/locales"
	"ohmontrouge/webapp/models"
	"ohmontrouge/webapp/public"
	"ohmontrouge/webapp/services"

	"github.com/antonlindstrom/pgstore"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo-pop/v3/pop/popmw"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/middleware/csrf"
	"github.com/gobuffalo/middleware/forcessl"
	"github.com/gobuffalo/middleware/i18n"
	"github.com/gobuffalo/middleware/paramlogger"
	"github.com/unrolled/secure"
)

var ENV = envy.Get("GO_ENV", "development")

var (
	app     *buffalo.App
	appOnce sync.Once
	T       *i18n.Translator
)

func App() *buffalo.App {
	appOnce.Do(func() {
		sessionSecret, err := envy.MustGet("SESSION_SECRET")
		if err != nil {
			panic(fmt.Sprintf("SESSION_SECRET not set: %v", err))
		}
		dbURL, err := envy.MustGet("DATABASE_URL")
		if err != nil {
			panic(fmt.Sprintf("DATABASE_URL not set: %v", err))
		}
		store, err := pgstore.NewPGStore(dbURL, []byte(sessionSecret))
		if err != nil {
			panic(fmt.Sprintf("pgstore: %v", err))
		}
		store.StopCleanup(store.Cleanup(time.Hour))

		app = buffalo.New(buffalo.Options{
			Env:          ENV,
			SessionName:  "_ohm_session",
			SessionStore: store,
		})

		app.Use(forceSSL())
		app.Use(paramlogger.ParameterLogger)
		app.Use(csrf.New)
		app.Use(popmw.Transaction(models.DB))
		app.Use(translations())

		authSvc := services.AccountService{
			Accounts: models.AccountStore{},
			Roles:    models.AccountRoleStore{},
		}

		// Inject sheet_music_url into every request context (empty string when unset).
		sheetMusicURL := envy.Get("SHEET_MUSIC_URL", "")
		app.Use(func(next buffalo.Handler) buffalo.Handler {
			return func(c buffalo.Context) error {
				c.Set("sheet_music_url", sheetMusicURL)
				return next(c)
			}
		})

		// Optionally load the authenticated account; sets "current_account" when valid.
		app.Use(LoadCurrentAccount(authSvc))

		auth := AuthHandler{Accounts: authSvc, Sessions: models.HTTPSessionStore{}, DB: models.DB}
		home := HomeHandler{}

		app.GET("/", home.Index)
		app.GET("/politique-de-confidentialite", home.Privacy)
		app.GET("/connexion", auth.Form)
		app.POST("/connexion", auth.Submit)
		app.GET("/deconnexion", auth.Logout)

		admin := app.Group("/admin")
		admin.Use(RequireActiveAccount(authSvc))
		admin.Use(RequireAdmin(authSvc))
		admin.GET("/", func(c buffalo.Context) error {
			return c.Render(http.StatusOK, r.HTML("admin/index.plush.html"))
		})

		app.ServeFiles("/", http.FS(public.FS()))
	})

	return app
}

func translations() buffalo.MiddlewareFunc {
	var err error
	if T, err = i18n.New(locales.FS(), "en-US"); err != nil {
		_ = app.Stop(err)
	}
	return T.Middleware()
}

func forceSSL() buffalo.MiddlewareFunc {
	return forcessl.Middleware(secure.Options{
		SSLRedirect:     ENV == "production",
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	})
}
