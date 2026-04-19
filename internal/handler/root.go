package handler

import (
	"encoding/json"
	"net/http"
)

// Root serves GET / with a small JSON discovery response so browsers and uptime checks
// hitting the site root do not see 404 (the API has no HTML UI).
func Root(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"service": "ase",
		"summary": "AI agent search API — use POST /v1/search with Bearer token",
		"links": map[string]string{
			"health":    "/health",
			"metrics":   "/metrics",
			"search":    "/v1/search",
			"documents": "/v1/documents",
		},
	})
}
