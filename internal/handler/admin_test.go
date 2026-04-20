package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
	handler.RegisterAdmin(r, cfg, signer)

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
}
