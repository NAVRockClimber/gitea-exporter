package config_test

import (
	"gitea-exporter/internal/infrastructure/config"
	"os"
	"path/filepath"
	"testing"
)

// config_test.go — a thumbnail test to run Load against a real file
func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte("servers:\n  git-ext:\n    url: https://example.com\n    tokenEnvName: GIT_EXT_TOKEN\n"), 0o644)

	t.Setenv("GIT_EXT_TOKEN", "secret-token")

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Servers["git-ext"].URL != "https://example.com" {
		t.Fatalf("unexpected URL: %q", cfg.Servers["git-ext"].URL)
	}

	// env-default:"prometheus" applies even when output: is absent from YAML,
	// so this minimal config must validate cleanly.
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: unexpected error %v", err)
	}
}

// TestLoadResolvesTokenFromEnv verifies the tokenEnvName indirection:
// YAML names an env var, and Load resolves that var into Server.Token.
func TestLoadResolvesTokenFromEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte("servers:\n  git-ext:\n    url: https://example.com\n    tokenEnvName: GIT_EXT_TOKEN\n"), 0o644)

	t.Setenv("GIT_EXT_TOKEN", "secret-token")

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := cfg.Servers["git-ext"].Token; got != "secret-token" {
		t.Fatalf("Token = %q, want %q", got, "secret-token")
	}
}
