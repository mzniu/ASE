package handler

import (
	"encoding/json"
	"net/http"

	"github.com/example/ase/internal/config"
	"github.com/example/ase/internal/webcontent"
)

// Root serves GET / with the embedded HTML homepage (project intro + Agent Skill setup).
func Root(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(webcontent.IndexHTML)
}

// ServiceInfo serves GET /api/info — JSON discovery for curl and integrations.
func ServiceInfo(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		links := map[string]string{
			"home":      "/",
			"health":    "/health",
			"metrics":   "/metrics",
			"search":    "/v1/search",
			"documents": "/v1/documents",
			"skill_md":  "/skills/ase-search-api/SKILL.md",
			"skill_ref": "/skills/ase-search-api/reference.md",
			"skill_zip": "/skills/ase-search-api/bundle.zip",
		}
		if cfg.AdminUIEnabled() {
			links["admin"] = "/admin/"
			links["admin_opensearch"] = "/admin/opensearch/"
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"service": "ase",
			"summary": "AI agent search API — POST /v1/search returns Markdown",
			"links":   links,
		})
	}
}
