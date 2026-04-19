package auth

import (
	"testing"

	"github.com/example/ase/internal/config"
)

func TestValidateAPIKey_authList(t *testing.T) {
	cfg := config.Config{AuthValidKeys: []string{"a", "b"}}
	if err := ValidateAPIKey("a", cfg); err != nil {
		t.Fatal(err)
	}
	if err := ValidateAPIKey("x", cfg); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateAPIKey_devKey(t *testing.T) {
	cfg := config.Config{DevAPIKey: "secret"}
	if err := ValidateAPIKey("secret", cfg); err != nil {
		t.Fatal(err)
	}
	if err := ValidateAPIKey("wrong", cfg); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateAPIKey_openDev(t *testing.T) {
	cfg := config.Config{}
	if err := ValidateAPIKey("anything", cfg); err != nil {
		t.Fatal(err)
	}
}
