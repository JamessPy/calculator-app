package config

import (
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
	if cfg.ReadTimeout != 5*time.Second {
		t.Errorf("ReadTimeout = %v, want 5s", cfg.ReadTimeout)
	}
	if len(cfg.AllowedOrigins) == 0 {
		t.Error("AllowedOrigins must not be empty")
	}
}

func TestLoad_PortFromEnvironment(t *testing.T) {
	t.Setenv("PORT", "9090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want 9090", cfg.Port)
	}
	if cfg.Addr() != ":9090" {
		t.Errorf("Addr() = %q, want :9090", cfg.Addr())
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port string
	}{
		{name: "not a number", port: "http"},
		{name: "zero", port: "0"},
		{name: "above range", port: "70000"},
		{name: "negative", port: "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PORT", tt.port)

			if _, err := Load(); err == nil {
				t.Errorf("Load() succeeded with PORT=%q, want an error", tt.port)
			}
		})
	}
}

func TestLoad_EmptyEnvironmentUsesDefault(t *testing.T) {
	t.Setenv("PORT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
}
