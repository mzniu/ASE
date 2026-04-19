package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/ase/internal/handler"
)

func TestSkillSKILLMD_GET(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/skills/ase-search-api/SKILL.md", nil)
	rr := httptest.NewRecorder()
	handler.SkillSKILLMD(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	if len(rr.Body.Bytes()) < 100 {
		t.Fatal("body too short")
	}
}

func TestSkillBundleZIP_GET(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/skills/ase-search-api/bundle.zip", nil)
	rr := httptest.NewRecorder()
	handler.SkillBundleZIP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/zip" {
		t.Fatalf("Content-Type = %q", ct)
	}
	if len(rr.Body.Bytes()) < 200 {
		t.Fatal("zip too small")
	}
}
