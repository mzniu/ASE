package handler

import (
	"archive/zip"
	"bytes"
	"fmt"
	"net/http"

	"github.com/example/ase/internal/agentskill"
)

// SkillSKILLMD serves GET /skills/ase-search-api/SKILL.md (embedded; no GitHub).
func SkillSKILLMD(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="SKILL.md"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(agentskill.SKILLMD)
}

// SkillReferenceMD serves GET /skills/ase-search-api/reference.md
func SkillReferenceMD(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="reference.md"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(agentskill.ReferenceMD)
}

// SkillBundleZIP serves GET /skills/ase-search-api/bundle.zip (SKILL.md + reference.md).
func SkillBundleZIP(w http.ResponseWriter, _ *http.Request) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, ent := range []struct {
		name string
		data []byte
	}{
		{"ase-search-api/SKILL.md", agentskill.SKILLMD},
		{"ase-search-api/reference.md", agentskill.ReferenceMD},
	} {
		f, err := zw.Create(ent.name)
		if err != nil {
			http.Error(w, fmt.Sprintf("zip: %v", err), http.StatusInternalServerError)
			return
		}
		if _, err := f.Write(ent.data); err != nil {
			http.Error(w, fmt.Sprintf("zip: %v", err), http.StatusInternalServerError)
			return
		}
	}
	if err := zw.Close(); err != nil {
		http.Error(w, fmt.Sprintf("zip: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="ase-search-api-skill.zip"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}
