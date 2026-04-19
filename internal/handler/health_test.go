package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/ase/internal/handler"
)

func TestHealth_GET(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.Health(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type=%q", ct)
	}
	if body := rr.Body.String(); body != `{"status":"ok"}` {
		t.Fatalf("body=%q", body)
	}
}
