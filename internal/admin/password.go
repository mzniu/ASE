package admin

import (
	"crypto/subtle"
	"strings"

	"github.com/example/ase/internal/config"
	"golang.org/x/crypto/bcrypt"
)

// CheckPassword verifies username/password against config (bcrypt preferred, then constant-time plain for dev).
func CheckPassword(cfg config.Config, username, password string) bool {
	if !cfg.AdminUIEnabled() {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(username), []byte(cfg.AdminUsername)) != 1 {
		return false
	}
	if cfg.AdminPasswordBcrypt != "" {
		return bcrypt.CompareHashAndPassword([]byte(cfg.AdminPasswordBcrypt), []byte(password)) == nil
	}
	p := strings.TrimSpace(cfg.AdminPasswordPlain)
	if p == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(password), []byte(p)) == 1
}
