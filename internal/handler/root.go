package handler

import (
	"encoding/json"
	"net/http"

	"github.com/example/ase/internal/webcontent"
)

// Root serves GET / with the embedded HTML homepage (project intro + Cursor Skill setup).
func Root(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(webcontent.IndexHTML)
}

// ServiceInfo serves GET /api/info — JSON discovery for curl and integrations.
func ServiceInfo(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"service": "ase",
		"summary": "AI agent search API — POST /v1/search returns Markdown",
		"links": map[string]string{
			"home":      "/",
			"health":    "/health",
			"metrics":   "/metrics",
			"search":    "/v1/search",
			"documents": "/v1/documents",
		},
	})
}
