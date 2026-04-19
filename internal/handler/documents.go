package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/example/ase/internal/auth"
	"github.com/example/ase/internal/config"
	"github.com/example/ase/internal/httpx"
	"github.com/example/ase/internal/port"
)

// Documents handles POST /v1/documents (index ingest when OpenSearch is configured).
type Documents struct {
	Cfg   config.Config
	Index port.IndexRepository
}

// NewDocuments constructs the handler.
func NewDocuments(cfg config.Config, idx port.IndexRepository) *Documents {
	return &Documents{Cfg: cfg, Index: idx}
}

// Handle validates JSON, authenticates, and indexes title + body_text.
func (d *Documents) Handle(w http.ResponseWriter, r *http.Request) {
	const maxBody = 1 << 20

	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	defer r.Body.Close()

	rid := reqID(r)

	if ct := r.Header.Get("Content-Type"); ct != "" && !strings.HasPrefix(ct, "application/json") {
		slog.Warn("documents rejected", "request_id", rid, "reason", "unsupported_media_type")
		httpx.WriteProblem(w, http.StatusUnsupportedMediaType, "unsupported media type", "Content-Type must be application/json")
		return
	}

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Warn("documents rejected", "request_id", rid, "reason", "read_body", "err", err)
		httpx.WriteProblem(w, http.StatusBadRequest, "invalid body", err.Error())
		return
	}

	var req documentIndexRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		slog.Warn("documents rejected", "request_id", rid, "reason", "invalid_json", "err", err)
		httpx.WriteProblem(w, http.StatusBadRequest, "invalid json", err.Error())
		return
	}
	if strings.TrimSpace(req.ID) == "" {
		slog.Warn("documents rejected", "request_id", rid, "reason", "empty_id")
		httpx.WriteProblem(w, http.StatusBadRequest, "validation error", "id is required")
		return
	}

	token, err := auth.TokenFromRequest(r)
	if err != nil {
		slog.Warn("documents rejected", "request_id", rid, "reason", "missing_or_invalid_auth")
		httpx.WriteProblem(w, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}
	if err := auth.ValidateAPIKey(token, d.Cfg); err != nil {
		slog.Warn("documents rejected", "request_id", rid, "reason", "invalid_api_key")
		httpx.WriteProblem(w, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	if err := d.Index.IndexDocument(r.Context(), req.ID, req.Title, req.BodyText); err != nil {
		if errors.Is(err, port.ErrIndexingDisabled) {
			httpx.WriteProblem(w, http.StatusNotImplemented, "indexing not available",
				"OpenSearch is not configured; set OPENSEARCH_URLS and OPENSEARCH_INDEX")
			return
		}
		slog.Error("index document failed", "request_id", rid, "doc_id", req.ID, "err", err)
		httpx.WriteProblem(w, http.StatusServiceUnavailable, "dependency unavailable", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
