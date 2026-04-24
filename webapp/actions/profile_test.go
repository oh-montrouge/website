package actions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/stretchr/testify/assert"
	"ohmontrouge/webapp/services"
)

// --- Show tests ---

func TestProfileHandler_Show_RendersProfile(t *testing.T) {
	profile := &services.MusicianProfile{
		AccountID:          1,
		FirstName:          "Alice",
		LastName:           "Martin",
		MainInstrumentName: "Clarinette",
	}
	account := &services.AccountDTO{ID: 1, Email: "alice@example.com", Status: services.StatusActive}

	h := ProfileHandler{Membership: &stubMusicianProfile{profile: profile}}
	app := newTestApp(func(a *buffalo.App) {
		a.Use(injectContextValue("current_account", account))
		a.GET("/profil", h.Show)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/profil", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "Alice")
	assert.Contains(t, res.Body.String(), "Clarinette")
	assert.Contains(t, res.Body.String(), "alice@example.com")
}

// --- AC-M7: Consent flag gates phone/address display ---

func TestProfileHandler_Show_PhoneAddress_ShownWhenConsentTrue(t *testing.T) {
	profile := &services.MusicianProfile{
		AccountID:           1,
		FirstName:           "Bob",
		LastName:            "Dupont",
		MainInstrumentName:  "Trompette",
		Phone:               "06 12 34 56 78",
		Address:             "12 rue de la Paix",
		PhoneAddressConsent: true,
	}
	account := &services.AccountDTO{ID: 1, Email: "bob@example.com", Status: services.StatusActive}

	h := ProfileHandler{Membership: &stubMusicianProfile{profile: profile}}
	app := newTestApp(func(a *buffalo.App) {
		a.Use(injectContextValue("current_account", account))
		a.GET("/profil", h.Show)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/profil", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	body := res.Body.String()
	assert.Contains(t, body, "06 12 34 56 78", "phone should be shown when consent is true")
	assert.Contains(t, body, "12 rue de la Paix", "address should be shown when consent is true")
}

func TestProfileHandler_Show_PhoneAddress_HiddenWhenConsentFalse(t *testing.T) {
	profile := &services.MusicianProfile{
		AccountID:           1,
		FirstName:           "Claire",
		LastName:            "Leblanc",
		MainInstrumentName:  "Hautbois",
		Phone:               "07 00 00 00 00",
		Address:             "1 avenue Victor Hugo",
		PhoneAddressConsent: false,
	}
	account := &services.AccountDTO{ID: 1, Email: "claire@example.com", Status: services.StatusActive}

	h := ProfileHandler{Membership: &stubMusicianProfile{profile: profile}}
	app := newTestApp(func(a *buffalo.App) {
		a.Use(injectContextValue("current_account", account))
		a.GET("/profil", h.Show)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/profil", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	body := res.Body.String()
	assert.NotContains(t, body, "07 00 00 00 00", "phone must not be shown when consent is false")
	assert.NotContains(t, body, "1 avenue Victor Hugo", "address must not be shown when consent is false")
}
