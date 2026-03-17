package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Fatalf("expected default port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Database.Port != 5432 {
		t.Fatalf("expected default db port 5432, got %d", cfg.Database.Port)
	}
	if cfg.BootstrapAdmin.Email != "" {
		t.Fatalf("expected empty bootstrap admin email, got %q", cfg.BootstrapAdmin.Email)
	}
}

func TestLoadBootstrapAdmin(t *testing.T) {
	clearEnv(t)
	t.Setenv("BOOTSTRAP_ADMIN_EMAIL", "admin@example.com")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.BootstrapAdmin.Email != "admin@example.com" {
		t.Fatalf("expected bootstrap email admin@example.com, got %q", cfg.BootstrapAdmin.Email)
	}
	if cfg.BootstrapAdmin.Password != "secret" {
		t.Fatalf("expected bootstrap password secret, got %q", cfg.BootstrapAdmin.Password)
	}
}

func TestLoadInvalidPort(t *testing.T) {
	clearEnv(t)
	t.Setenv("PORT", "bad")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid PORT")
	}
}

func clearEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"PORT",
		"DB_HOST",
		"DB_PORT",
		"DB_NAME",
		"DB_USER",
		"DB_PASSWORD",
		"DB_SSLMODE",
		"JWT_SECRET",
		"JWT_ACCESS_TTL_MINUTES",
		"BOOTSTRAP_ADMIN_EMAIL",
		"BOOTSTRAP_ADMIN_PASSWORD",
	}
	for _, k := range keys {
		_ = os.Unsetenv(k)
	}
}

