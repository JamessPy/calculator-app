// Package config loads runtime configuration from environment variables.
//
// Environment variables are used rather than config files because the
// service is meant to run in a container, where the environment is the
// standard injection point (12-factor app, factor III).
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the runtime configuration of the service.
type Config struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	AllowedOrigins  []string
}

// Load reads the configuration from the environment, falling back to
// defaults that are sensible for local development.
func Load() (Config, error) {
	cfg := Config{
		Port:            getString("PORT", "8080"),
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 15 * time.Second,
		AllowedOrigins:  []string{"http://localhost:5173", "http://localhost:3000"},
	}

	if err := validatePort(cfg.Port); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Addr returns the address the HTTP server should listen on.
// The host part is left empty so the server binds to all interfaces,
// which is required inside a container.
func (c Config) Addr() string {
	return ":" + c.Port
}

func getString(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func validatePort(port string) error {
	n, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("PORT must be a number, got %q", port)
	}
	if n < 1 || n > 65535 {
		return fmt.Errorf("PORT must be between 1 and 65535, got %d", n)
	}
	return nil
}
