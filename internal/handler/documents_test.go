package handler_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/ase/internal/adapter/noopindex"
	"github.com/example/ase/internal/config"
	"github.com/example/ase/internal/handler"
)

func TestDocuments_POST_withoutAuth_returns401(t *testing.T) {
	t.Setenv("DEV_API_KEY", "secret-key")
	cfg := config.Load()
	h := handler.NewDocuments(cfg, noopindex.Repo{})

	req := httptest.NewRequest(http.MethodPost, "/v1/documents", bytes.NewBufferString(`{"id":"a","title":"t","body_text":"b"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Handle(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestDocuments_POST_noOpenSearch_returns501(t *testing.T) {
	t.Setenv("DEV_API_KEY", "secret-key")
	cfg := config.Load()
	h := handler.NewDocuments(cfg, noopindex.Repo{})

	req := httptest.NewRequest(http.MethodPost, "/v1/documents", bytes.NewBufferString(`{"id":"a","title":"t","body_text":"b"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer secret-key")
	rr := httptest.NewRecorder()

	h.Handle(rr, req)

	if rr.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusNotImplemented, rr.Body.String())
	}
}

func TestDocuments_emptyID_returns400(t *testing.T) {
	t.Setenv("DEV_API_KEY", "secret-key")
	cfg := config.Load()
	h := handler.NewDocuments(cfg, noopindex.Repo{})

	req := httptest.NewRequest(http.MethodPost, "/v1/documents", bytes.NewBufferString(`{"id":"  ","title":"t","body_text":"b"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer secret-key")
	rr := httptest.NewRecorder()

	h.Handle(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}
