package admin

import (
	"testing"
	"time"

	"github.com/example/ase/internal/config"
	"golang.org/x/crypto/bcrypt"
)

func TestSessionSigner_roundTrip(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		AdminUsername:       "u1",
		AdminPasswordBcrypt: string(hash),
		AdminSessionSecret:  "sixteencharslong",
		AdminSessionTTL:     time.Hour,
	}
	if !cfg.AdminUIEnabled() {
		t.Fatal("fixture should enable admin")
	}
	s := NewSessionSigner(cfg)
	now := time.Unix(1_700_000_000, 0)
	tok, err := s.Issue(now)
	if err != nil {
		t.Fatal(err)
	}
	if !s.Verify(tok, now.Add(30*time.Minute)) {
		t.Fatal("verify should succeed before exp")
	}
	if s.Verify(tok, now.Add(2*time.Hour)) {
		t.Fatal("verify should fail after exp")
	}
}
