package actions

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"ohmontrouge/webapp/services"
)

func runRetentionIndex(t *testing.T, h RetentionHandler) *httptest.ResponseRecorder {
	t.Helper()
	return serveGET(t, "/admin/retention", h.Index)
}

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
	res := runRetentionIndex(t, RetentionHandler{Compliance: &stubCompliance{retentionList: entries}})
	assert.Equal(t, http.StatusOK, res.Code)
	body := res.Body.String()
	assert.Contains(t, body, "Valjean")
	assert.Contains(t, body, "Tuba")
	assert.Contains(t, body, "2017-2018")
}

func TestRetentionHandler_Index_EmptyList(t *testing.T) {
	res := runRetentionIndex(t, RetentionHandler{Compliance: &stubCompliance{retentionList: []services.RetentionEntryDTO{}}})
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "Aucun compte")
}
