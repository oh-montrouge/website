package actions

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/stretchr/testify/assert"
	"ohmontrouge/webapp/services"
)

var home = HomeHandler{}

// --- AC-M1: public routes require no authentication ---

func TestHomeIndex_Returns200WithoutSession(t *testing.T) {
	app := newTestApp(func(a *buffalo.App) {
		a.GET("/", home.Index)
	})

	res := httptest.NewRecorder()
	app.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusOK, res.Code)
}

func TestPrivacy_Returns200WithoutSession(t *testing.T) {
	app := newTestApp(func(a *buffalo.App) {
		a.GET("/politique-de-confidentialite", home.Privacy)
	})

	res := httptest.NewRecorder()
	app.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/politique-de-confidentialite", nil))

	assert.Equal(t, http.StatusOK, res.Code)
}

// --- AC-M2: sheet music link gated by SHEET_MUSIC_URL ---

func TestHomeIndex_AuthenticatedWithSheetMusicURL_ContainsPartitions(t *testing.T) {
	const sheetURL = "https://drive.example.com/partitions"
	app := newTestApp(func(a *buffalo.App) {
		a.Use(injectContextValue("current_account", &services.AccountDTO{ID: 1, Email: "test@example.com", Status: services.StatusActive}))
		a.Use(injectContextValue("sheet_music_url", sheetURL))
		a.GET("/", home.Index)
	})

	res := httptest.NewRecorder()
	app.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusOK, res.Code)
	body := res.Body.String()
	assert.True(t, strings.Contains(body, "Partitions"), "nav should contain Partitions link when SHEET_MUSIC_URL is set")
	assert.True(t, strings.Contains(body, sheetURL), "nav should contain the configured sheet music URL")
}

func TestHomeIndex_AuthenticatedWithoutSheetMusicURL_NoPartitions(t *testing.T) {
	app := newTestApp(func(a *buffalo.App) {
		a.Use(injectContextValue("current_account", &services.AccountDTO{ID: 1, Email: "test@example.com", Status: services.StatusActive}))
		a.Use(injectContextValue("sheet_music_url", ""))
		a.GET("/", home.Index)
	})

	res := httptest.NewRecorder()
	app.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusOK, res.Code)
	assert.False(t, strings.Contains(res.Body.String(), "Partitions"), "nav should not contain Partitions link when SHEET_MUSIC_URL is unset")
}

// --- AC-M3: privacy notice link present on all pages ---

func TestHomeIndex_ContainsPrivacyLink(t *testing.T) {
	res := serveGET(t, "/", home.Index, injectContextValue("sheet_music_url", ""))
	assert.True(t, strings.Contains(res.Body.String(), "/politique-de-confidentialite"),
		"homepage should contain a link to the privacy notice")
}

func TestLoginForm_ContainsPrivacyLink(t *testing.T) {
	res := serveGET(t, "/connexion", AuthHandler{}.Form, injectContextValue("sheet_music_url", ""))
	assert.True(t, strings.Contains(res.Body.String(), "/politique-de-confidentialite"),
		"login page should contain a link to the privacy notice")
}

func TestAuthenticatedPage_ContainsPrivacyLink(t *testing.T) {
	res := serveGET(t, "/", home.Index,
		injectContextValue("current_account", &services.AccountDTO{ID: 1, Email: "test@example.com", Status: services.StatusActive}),
		injectContextValue("sheet_music_url", ""))
	assert.True(t, strings.Contains(res.Body.String(), "/politique-de-confidentialite"),
		"authenticated page should contain a link to the privacy notice")
}
