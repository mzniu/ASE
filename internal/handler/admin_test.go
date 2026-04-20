package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/ase/internal/adapter/noopindex"
	"github.com/example/ase/internal/admin"
	"github.com/example/ase/internal/config"
	"github.com/example/ase/internal/handler"
	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

func TestAdmin_login_session_config(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret1"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		AdminUsername:       "adm",
		AdminPasswordBcrypt: string(hash),
		AdminSessionSecret:  "sixteencharslong",
		AdminSessionTTL:     time.Hour,
	}
	if !cfg.AdminUIEnabled() {
		t.Fatal("admin should be enabled")
	}
	signer := admin.NewSessionSigner(cfg)
	r := chi.NewRouter()
	handler.RegisterAdmin(r, cfg, signer, noopindex.Repo{})

	// login
	body, _ := json.Marshal(map[string]string{"username": "adm", "password": "secret1"})
	req := httptest.NewRequest(http.MethodPost, "/admin/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("login status = %d body=%s", rr.Code, rr.Body.String())
	}
	var cookies []*http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == "ase_admin" {
			cookies = append(cookies, c)
		}
	}
	if len(cookies) != 1 || cookies[0].Value == "" {
		t.Fatal("expected session cookie")
	}

	req2 := httptest.NewRequest(http.MethodGet, "/admin/api/config", nil)
	req2.AddCookie(cookies[0])
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("config status = %d %s", rr2.Code, rr2.Body.String())
	}
	var snap map[string]any
	if err := json.Unmarshal(rr2.Body.Bytes(), &snap); err != nil {
		t.Fatal(err)
	}
	if snap["http_addr"] == nil {
		t.Fatal("expected config snapshot")
	}

	req3 := httptest.NewRequest(http.MethodGet, "/admin/api/opensearch/documents", nil)
	req3.AddCookie(cookies[0])
	rr3 := httptest.NewRecorder()
	r.ServeHTTP(rr3, req3)
	if rr3.Code != http.StatusServiceUnavailable {
		t.Fatalf("opensearch documents without cluster: status = %d want 503 body=%s", rr3.Code, rr3.Body.String())
	}

	req4 := httptest.NewRequest(http.MethodGet, "/admin/api/opensearch/hits?q=test", nil)
	req4.AddCookie(cookies[0])
	rr4 := httptest.NewRecorder()
	r.ServeHTTP(rr4, req4)
	if rr4.Code != http.StatusServiceUnavailable {
		t.Fatalf("opensearch hits without cluster: status = %d want 503 body=%s", rr4.Code, rr4.Body.String())
	}

	req5 := httptest.NewRequest(http.MethodGet, "/admin/api/opensearch/hits", nil)
	req5.AddCookie(cookies[0])
	rr5 := httptest.NewRecorder()
	r.ServeHTTP(rr5, req5)
	if rr5.Code != http.StatusBadRequest {
		t.Fatalf("opensearch hits without q: status = %d want 400 body=%s", rr5.Code, rr5.Body.String())
	}
}

func TestAdmin_openSearch_hits_noop_index(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret1"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		AdminUsername:       "adm",
		AdminPasswordBcrypt: string(hash),
		AdminSessionSecret:  "sixteencharslong",
		AdminSessionTTL:     time.Hour,
		OpenSearchURLs:      []string{"http://127.0.0.1:9200"},
		OpenSearchIndex:     "idx",
	}
	signer := admin.NewSessionSigner(cfg)
	r := chi.NewRouter()
	handler.RegisterAdmin(r, cfg, signer, noopindex.Repo{})

	body, _ := json.Marshal(map[string]string{"username": "adm", "password": "secret1"})
	req := httptest.NewRequest(http.MethodPost, "/admin/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body.String())
	}
	var cookie *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == "ase_admin" {
			cookie = c
			break
		}
	}
	if cookie == nil {
		t.Fatal("no session cookie")
	}

	req2 := httptest.NewRequest(http.MethodGet, "/admin/api/opensearch/hits?q=hello", nil)
	req2.AddCookie(cookie)
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("hits status = %d %s", rr2.Code, rr2.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rr2.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out["query"] != "hello" {
		t.Fatalf("query field: %v", out["query"])
	}
	hits, _ := out["hits"].([]any)
	if hits == nil {
		t.Fatal("expected hits array")
	}
	if len(hits) != 0 {
		t.Fatalf("noop index: want 0 hits got %d", len(hits))
	}
}
