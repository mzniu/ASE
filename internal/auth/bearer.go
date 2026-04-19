package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/example/ase/internal/config"
)

// TokenFromRequest extracts the Bearer token from Authorization.
func TokenFromRequest(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", errors.New("missing Authorization")
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return "", errors.New("invalid Authorization scheme")
	}
	token := strings.TrimSpace(strings.TrimPrefix(h, prefix))
	if token == "" {
		return "", errors.New("empty bearer token")
	}
	return token, nil
}

// ValidateAPIKey checks the token against AUTH_VALID_API_KEYS or DEV_API_KEY (REQ-F-003).
func ValidateAPIKey(token string, cfg config.Config) error {
	if len(cfg.AuthValidKeys) > 0 {
		for _, k := range cfg.AuthValidKeys {
			if token == k {
				return nil
			}
		}
		return errors.New("invalid api key")
	}
	if cfg.DevAPIKey != "" {
		if token == cfg.DevAPIKey {
			return nil
		}
		return errors.New("invalid api key")
	}
	// No keys configured: accept any non-empty token (local dev only).
	return nil
}
