package tests

import (
	"testing"

	"github.com/joseph/ollama-logging-proxy/internal/config"
)

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv(config.EnvListen, "")
	t.Setenv(config.EnvTarget, "")
	t.Setenv(config.EnvMaxBodyBytes, "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Listen != config.DefaultListen {
		t.Fatalf("expected default listen %q, got %q", config.DefaultListen, cfg.Listen)
	}
	if cfg.Target.String() != config.DefaultTarget {
		t.Fatalf("expected default target %q, got %q", config.DefaultTarget, cfg.Target.String())
	}
	if cfg.MaxBodyBytes != config.DefaultMaxBodyBytes {
		t.Fatalf("expected default max body bytes %d, got %d", config.DefaultMaxBodyBytes, cfg.MaxBodyBytes)
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	t.Setenv(config.EnvListen, "127.0.0.1:18080")
	t.Setenv(config.EnvTarget, "http://127.0.0.1:19090")
	t.Setenv(config.EnvMaxBodyBytes, "2048")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Listen != "127.0.0.1:18080" {
		t.Fatalf("expected overridden listen, got %q", cfg.Listen)
	}
	if cfg.Target.String() != "http://127.0.0.1:19090" {
		t.Fatalf("expected overridden target, got %q", cfg.Target.String())
	}
	if cfg.MaxBodyBytes != 2048 {
		t.Fatalf("expected overridden max body bytes, got %d", cfg.MaxBodyBytes)
	}
}

func TestLoadConfigInvalidTarget(t *testing.T) {
	t.Setenv(config.EnvTarget, "127.0.0.1:19090")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for target without scheme, got nil")
	}
}

func TestLoadConfigInvalidMaxBodyBytes(t *testing.T) {
	t.Setenv(config.EnvMaxBodyBytes, "-1")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for invalid max body bytes, got nil")
	}
}
