package actions

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/stretchr/testify/assert"
	"ohmontrouge/webapp/services"
)

func TestRetentionHandler_Index_RendersList(t *testing.T) {
	endDate := time.Date(2018, 6, 30, 0, 0, 0, 0, time.UTC)
	entries := []services.RetentionEntryDTO{
		{
			AccountID:          3,
			FirstName:          "Jean",
			LastName:           "Valjean",
			MainInstrumentName: "Tuba",
			LastSeasonLabel:    "2017-2018",
			LastSeasonEndDate:  endDate,
		},
	}
	h := RetentionHandler{Compliance: &stubCompliance{retentionList: entries}}
	app := newTestApp(func(a *buffalo.App) {
		a.GET("/admin/retention", h.Index)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/retention", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	body := res.Body.String()
	assert.Contains(t, body, "Valjean")
	assert.Contains(t, body, "Tuba")
	assert.Contains(t, body, "2017-2018")
}

func TestRetentionHandler_Index_EmptyList(t *testing.T) {
	h := RetentionHandler{Compliance: &stubCompliance{retentionList: []services.RetentionEntryDTO{}}}
	app := newTestApp(func(a *buffalo.App) {
		a.GET("/admin/retention", h.Index)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/retention", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "Aucun compte")
}
