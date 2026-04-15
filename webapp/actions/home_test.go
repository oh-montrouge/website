package actions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/stretchr/testify/assert"
)

func TestHomeHandler_ReturnsOK(t *testing.T) {
	app := newTestApp(func(a *buffalo.App) {
		a.GET("/", HomeHandler)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	app.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
}
