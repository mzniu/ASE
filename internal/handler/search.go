package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/example/ase/internal/auth"
	"github.com/example/ase/internal/config"
	"github.com/example/ase/internal/httpx"
	"github.com/example/ase/internal/metrics"
	"github.com/example/ase/internal/orchestrator"
)

// MarkdownSearcher is satisfied by *orchestrator.Service (tests may use fakes).
type MarkdownSearcher interface {
	SearchMarkdown(ctx context.Context, query string, providers []string, deepSearch *bool) (string, error)
}

// Search serves POST /v1/search.
type Search struct {
	Cfg  config.Config
	Orch MarkdownSearcher
}

// NewSearch constructs a handler with orchestrator wiring.
func NewSearch(cfg config.Config, orch MarkdownSearcher) *Search {
	return &Search{Cfg: cfg, Orch: orch}
}

// Handle validates the request, runs search orchestration, writes Markdown or Problem Details.
func (s *Search) Handle(w http.ResponseWriter, r *http.Request) {
	const maxBody = 1 << 20 // 1 MiB

	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	defer r.Body.Close()

	rid := reqID(r)

	if ct := r.Header.Get("Content-Type"); ct != "" && !strings.HasPrefix(ct, "application/json") {
		slog.Warn("search rejected", "request_id", rid, "reason", "unsupported_media_type")
		httpx.WriteProblem(w, http.StatusUnsupportedMediaType, "unsupported media type", "Content-Type must be application/json")
		return
	}

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Warn("search rejected", "request_id", rid, "reason", "read_body", "err", err)
		httpx.WriteProblem(w, http.StatusBadRequest, "invalid body", err.Error())
		return
	}

	var req searchRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		slog.Warn("search rejected", "request_id", rid, "reason", "invalid_json", "err", err)
		httpx.WriteProblem(w, http.StatusBadRequest, "invalid json", err.Error())
		return
	}
	if strings.TrimSpace(req.Query) == "" {
		slog.Warn("search rejected", "request_id", rid, "reason", "empty_query")
		httpx.WriteProblem(w, http.StatusBadRequest, "validation error", "query is required")
		return
	}
	if utf8.RuneCountInString(req.Query) > s.Cfg.MaxQueryRunes {
		slog.Warn("search rejected", "request_id", rid, "reason", "query_too_long", "query_runes", utf8.RuneCountInString(req.Query), "max", s.Cfg.MaxQueryRunes)
		httpx.WriteProblem(w, http.StatusBadRequest, "validation error", "query too long")
		return
	}

	token, err := auth.TokenFromRequest(r)
	if err != nil {
		slog.Warn("search rejected", "request_id", rid, "reason", "missing_or_invalid_auth")
		httpx.WriteProblem(w, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}
	if err := auth.ValidateAPIKey(token, s.Cfg); err != nil {
		slog.Warn("search rejected", "request_id", rid, "reason", "invalid_api_key")
		httpx.WriteProblem(w, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	md, err := s.Orch.SearchMarkdown(r.Context(), req.Query, req.Providers, req.DeepSearch)
	metrics.RecordSearchOrchestration(err)
	if err != nil {
		if errors.Is(err, orchestrator.ErrBadRequest) {
			slog.Warn("search rejected", "request_id", rid, "reason", "invalid_providers", "err", err)
		} else {
			slog.Error("search failed", "request_id", rid, "query_runes", utf8.RuneCountInString(req.Query), "err", err)
		}
		httpx.WriteSearchFailure(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(md))
}

func reqID(r *http.Request) string {
	id := middleware.GetReqID(r.Context())
	if id == "" {
		return "-"
	}
	return id
}
