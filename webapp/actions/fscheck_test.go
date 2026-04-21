package actions

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"ohmontrouge/webapp/public"
)

func TestCSSServed(t *testing.T) {
	// Simulate the file server exactly as app.go sets it up
	srv := httptest.NewServer(http.FileServer(http.FS(public.FS())))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/assets/application.css")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("closing response body: %v", err)
		}
	}()
	body, _ := io.ReadAll(resp.Body)
	t.Logf("status=%d  len=%d  content-type=%s  first100=%q",
		resp.StatusCode, len(body), resp.Header.Get("Content-Type"), string(body[:min(100, len(body))]))
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if len(body) < 1000 {
		t.Errorf("CSS too short: %d bytes", len(body))
	}
}
