package actions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"ohmontrouge/webapp/services"
)

func runProfileShow(t *testing.T, h ProfileHandler, account *services.AccountDTO) *httptest.ResponseRecorder {
	t.Helper()
	return serveGET(t, "/profil", h.Show, injectContextValue("current_account", account))
}

// --- Show tests ---

func TestProfileHandler_Show_RendersProfile(t *testing.T) {
	profile := &services.MusicianProfile{
		AccountID:          1,
		FirstName:          "Alice",
		LastName:           "Martin",
		MainInstrumentName: "Clarinette",
	}
	account := &services.AccountDTO{ID: 1, Email: "alice@example.com", Status: services.StatusActive}
	res := runProfileShow(t, ProfileHandler{Membership: &stubMusicianProfile{profile: profile}}, account)

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
	res := runProfileShow(t, ProfileHandler{Membership: &stubMusicianProfile{profile: profile}}, account)

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
	res := runProfileShow(t, ProfileHandler{Membership: &stubMusicianProfile{profile: profile}}, account)

	assert.Equal(t, http.StatusOK, res.Code)
	body := res.Body.String()
	assert.NotContains(t, body, "07 00 00 00 00", "phone must not be shown when consent is false")
	assert.NotContains(t, body, "1 avenue Victor Hugo", "address must not be shown when consent is false")
}
