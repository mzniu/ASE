package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/ase/internal/adapter/noopindex"
	"github.com/example/ase/internal/adapter/stubprovider"
	"github.com/example/ase/internal/config"
	"github.com/example/ase/internal/handler"
	"github.com/example/ase/internal/httpx"
	"github.com/example/ase/internal/orchestrator"
	"github.com/example/ase/internal/port"
	"github.com/go-chi/chi/v5"
)

func testSearch(t *testing.T, cfg config.Config) *handler.Search {
	t.Helper()
	prov := stubprovider.Fixed{
		Result: port.ProviderResult{
			Items: []port.ProviderItem{{Title: "stub", Snippet: "fallback line for tests"}},
		},
	}
	orch := &orchestrator.Service{
		Index: noopindex.Repo{},
		Registry: map[string]port.SearchProvider{
			"stub": prov,
		},
		DefaultNames: []string{"stub"},
		Config:       cfg,
	}
	return handler.NewSearch(cfg, orch)
}

func TestSearch_POST_withoutAuth_returns401(t *testing.T) {
	t.Setenv("DEV_API_KEY", "secret-key")
	cfg := config.Load()
	h := testSearch(t, cfg)

	req := httptest.NewRequest(http.MethodPost, "/v1/search", bytes.NewBufferString(`{"query":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Handle(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestSearch_POST_withValidKey_returns200Markdown(t *testing.T) {
	t.Setenv("DEV_API_KEY", "secret-key")
	cfg := config.Load()
	h := testSearch(t, cfg)

	req := httptest.NewRequest(http.MethodPost, "/v1/search", bytes.NewBufferString(`{"query":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer secret-key")
	rr := httptest.NewRecorder()

	h.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/markdown; charset=utf-8" {
		t.Fatalf("Content-Type = %q", ct)
	}
}

func TestSearch_emptyQuery_returns400(t *testing.T) {
	cfg := config.Config{
		MaxQueryRunes:    4096,
		MaxResponseRunes: 16000,
		RequestDeadline:  time.Minute,
		MinHitCount:      1,
		MinTotalTextLen:  1,
		MinSimilarity:    0,
		RateLimitGlobal:  1000,
		RateLimitPerKey:  1000,
		RateLimitBurst:   2000,
		RateLimitGlobalBurst: 2000,
	}
	h := testSearch(t, cfg)

	req := httptest.NewRequest(http.MethodPost, "/v1/search", bytes.NewBufferString(`{"query":"  "}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer any")
	rr := httptest.NewRecorder()

	h.Handle(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

type fakeSearcher struct {
	md  string
	err error
}

func (f *fakeSearcher) SearchMarkdown(_ context.Context, _ string, _ []string, _ *bool) (string, error) {
	return f.md, f.err
}

type deepSearchSpy struct {
	got *bool
}

func (d *deepSearchSpy) SearchMarkdown(_ context.Context, _ string, _ []string, deep *bool) (string, error) {
	d.got = deep
	return "ok", nil
}

func TestSearch_deepsearch_passedToOrchestrator(t *testing.T) {
	t.Setenv("DEV_API_KEY", "secret-key")
	cfg := config.Load()
	spy := &deepSearchSpy{}
	h := handler.NewSearch(cfg, spy)

	body := map[string]any{"query": "hi", "deepsearch": true}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/search", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer secret-key")
	rr := httptest.NewRecorder()
	h.Handle(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	if spy.got == nil || !*spy.got {
		t.Fatalf("deepSearch = %v", spy.got)
	}
}

func TestSearch_orchestratorErr_mapsProblemDetails(t *testing.T) {
	t.Setenv("DEV_API_KEY", "secret-key")
	cfg := config.Load()
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantTitle  string
	}{
		{"bad_request", orchestrator.ErrBadRequest, 400, "validation error"},
		{"deadline", context.DeadlineExceeded, 504, "gateway timeout"},
		{"canceled", context.Canceled, 503, "dependency unavailable"},
		{"generic", errors.New("opensearch down"), 503, "dependency unavailable"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewSearch(cfg, &fakeSearcher{err: tc.err})
			req := httptest.NewRequest(http.MethodPost, "/v1/search", bytes.NewBufferString(`{"query":"hello"}`))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer secret-key")
			rr := httptest.NewRecorder()
			h.Handle(rr, req)
			if rr.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d body=%s", rr.Code, tc.wantStatus, rr.Body.String())
			}
			var p httpx.ProblemDetail
			if err := json.Unmarshal(rr.Body.Bytes(), &p); err != nil {
				t.Fatal(err)
			}
			if p.Title != tc.wantTitle {
				t.Fatalf("title = %q, want %q", p.Title, tc.wantTitle)
			}
		})
	}
}

func TestSearch_GET_returns405(t *testing.T) {
	cfg := config.Config{MaxQueryRunes: 4096, MaxResponseRunes: 16000, RequestDeadline: time.Minute,
		MinHitCount: 1, MinTotalTextLen: 1, MinSimilarity: 0,
		RateLimitGlobal: 1000, RateLimitPerKey: 1000, RateLimitBurst: 2000, RateLimitGlobalBurst: 2000}
	h := testSearch(t, cfg)

	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Post("/search", h.Handle)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/search", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}
