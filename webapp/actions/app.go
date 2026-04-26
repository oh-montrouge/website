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
			Accounts:     models.AccountStore{},
			Roles:        models.AccountRoleStore{},
			InviteTokens: models.InviteTokenStore{},
			ResetTokens:  models.PasswordResetTokenStore{},
			Events:       models.RSVPStore{},
		}

		seasonSvc := services.SeasonService{
			Seasons: models.SeasonStore{},
		}

		feePaymentSvc := services.FeePaymentService{
			FeePayments: models.FeePaymentStore{},
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
		tokens := TokensHandler{Accounts: authSvc, Sessions: models.HTTPSessionStore{}, DB: models.DB}
		seasons := SeasonsHandler{Seasons: seasonSvc}

		app.GET("/", home.Index)
		app.GET("/politique-de-confidentialite", home.Privacy)
		app.GET("/connexion", auth.Form)
		app.POST("/connexion", auth.Submit)
		app.GET("/deconnexion", auth.Logout)
		app.GET("/invitation/{token}", tokens.InviteForm)
		app.POST("/invitation/{token}", tokens.InviteSubmit)
		app.GET("/reinitialiser-mot-de-passe/{token}", tokens.ResetForm)
		app.POST("/reinitialiser-mot-de-passe/{token}", tokens.ResetSubmit)

		admin := app.Group("/admin")
		admin.Use(RequireActiveAccount(authSvc))
		admin.Use(RequireAdmin(authSvc))
		admin.GET("/", func(c buffalo.Context) error {
			return c.Render(http.StatusOK, r.HTML("admin/index.plush.html"))
		})

		membershipSvc := services.MembershipService{
			Membership: models.AccountStore{},
		}
		complianceSvc := services.ComplianceService{
			Accounts:     models.AccountStore{},
			Membership:   models.AccountStore{},
			Roles:        models.AccountRoleStore{},
			InviteTokens: models.InviteTokenStore{},
			ResetTokens:  models.PasswordResetTokenStore{},
			Sessions:     models.HTTPSessionStore{},
			Events:       models.RSVPStore{},
		}

		baseURL := envy.Get("APP_BASE_URL", "http://localhost:3000")

		musicians := MusiciansHandler{
			Accounts:    authSvc,
			Membership:  membershipSvc,
			Compliance:  complianceSvc,
			Instruments: models.InstrumentStore{},
			FeePayments: feePaymentSvc,
			Seasons:     seasonSvc,
			BaseURL:     baseURL,
		}
		profileH := ProfileHandler{Membership: membershipSvc}
		retentionH := RetentionHandler{Compliance: complianceSvc}

		// Musician admin routes
		admin.GET("/musiciens", musicians.Index)
		admin.GET("/musiciens/nouveau", musicians.New)
		admin.POST("/musiciens", musicians.Create)
		admin.GET("/musiciens/{id}", musicians.Show)
		admin.GET("/musiciens/{id}/modifier", musicians.Edit)
		admin.PUT("/musiciens/{id}", musicians.Update)
		admin.DELETE("/musiciens/{id}", musicians.Delete)
		admin.POST("/musiciens/{id}/anonymiser", musicians.Anonymize)
		admin.POST("/musiciens/{id}/invitation", musicians.GenerateInviteLink)
		admin.POST("/musiciens/{id}/reinitialisation", musicians.GenerateResetLink)
		admin.POST("/musiciens/{id}/role-admin", musicians.GrantAdmin)
		admin.DELETE("/musiciens/{id}/role-admin", musicians.RevokeAdmin)
		admin.DELETE("/musiciens/{id}/consentement", musicians.WithdrawConsent)
		admin.POST("/musiciens/{id}/restriction", musicians.ToggleProcessingRestriction)

		// Fee payment routes
		feePayments := FeePaymentsHandler{FeePayments: feePaymentSvc}
		admin.POST("/musiciens/{account_id}/cotisations", feePayments.Create)
		admin.GET("/cotisations/{id}/modifier", feePayments.EditForm)
		admin.PUT("/cotisations/{id}", feePayments.Update)
		admin.DELETE("/cotisations/{id}", feePayments.Delete)

		// Retention review
		admin.GET("/retention", retentionH.Index)

		eventSvc := services.EventService{
			Events: models.EventStore{},
			RSVPs:  models.RSVPStore{},
		}
		eventsH := EventsHandler{
			Events:      eventSvc,
			Instruments: models.InstrumentStore{},
			Membership:  membershipSvc,
		}

		// Admin event routes
		admin.GET("/evenements/nouveau", eventsH.New)
		admin.POST("/evenements", eventsH.Create)
		admin.GET("/evenements/{id}/modifier", eventsH.Edit)
		admin.PUT("/evenements/{id}", eventsH.Update)
		admin.DELETE("/evenements/{id}", eventsH.Delete)
		admin.POST("/evenements/{id}/champs", eventsH.AddField)
		admin.GET("/evenements/{id}/champs/{field_id}/modifier", eventsH.EditFieldForm)
		admin.PUT("/evenements/{id}/champs/{field_id}", eventsH.UpdateField)
		admin.DELETE("/evenements/{id}/champs/{field_id}", eventsH.DeleteField)
		admin.PATCH("/evenements/{id}/rsvp/{musician_id}", eventsH.AdminUpdateRSVP)

		// Profile route (authenticated, not admin-only)
		authenticated := app.Group("")
		authenticated.Use(RequireActiveAccount(authSvc))
		authenticated.GET("/profil", profileH.Show)
		authenticated.GET("/tableau-de-bord", eventsH.Dashboard)
		authenticated.GET("/evenements", eventsH.Index)
		authenticated.GET("/evenements/{id}", eventsH.Show)
		authenticated.POST("/evenements/{id}/rsvp", eventsH.UpdateRSVP)

		admin.GET("/saisons", seasons.Index)
		admin.POST("/saisons", seasons.Create)
		admin.POST("/saisons/{id}/courante", seasons.DesignateCurrent)

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
