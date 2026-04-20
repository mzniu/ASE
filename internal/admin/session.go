package admin

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/example/ase/internal/config"
)

// SessionSigner issues and verifies HMAC-signed session tokens for the admin cookie.
type SessionSigner struct {
	secret []byte
	user   string
	ttl    time.Duration
}

// NewSessionSigner returns nil if admin UI is not fully configured.
func NewSessionSigner(cfg config.Config) *SessionSigner {
	if !cfg.AdminUIEnabled() {
		return nil
	}
	return &SessionSigner{
		secret: []byte(cfg.AdminSessionSecret),
		user:   cfg.AdminUsername,
		ttl:    cfg.AdminSessionTTL,
	}
}

type tokenClaims struct {
	U   string `json:"u"`
	Exp int64  `json:"exp"`
}

// Issue returns a compact signed token for the configured admin user.
func (s *SessionSigner) Issue(now time.Time) (string, error) {
	c := tokenClaims{U: s.user, Exp: now.Add(s.ttl).Unix()}
	raw, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	sig := hmac.New(sha256.New, s.secret)
	_, _ = sig.Write(raw)
	sum := sig.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(raw) + "." + base64.RawURLEncoding.EncodeToString(sum), nil
}

// Verify checks token and returns true if valid and not expired.
func (s *SessionSigner) Verify(token string, now time.Time) bool {
	if s == nil || token == "" {
		return false
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	want, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	sig := hmac.New(sha256.New, s.secret)
	_, _ = sig.Write(raw)
	sum := sig.Sum(nil)
	if subtle.ConstantTimeCompare(sum, want) != 1 {
		return false
	}
	var c tokenClaims
	if err := json.Unmarshal(raw, &c); err != nil {
		return false
	}
	if c.U != s.user || c.Exp <= now.Unix() {
		return false
	}
	return true
}
