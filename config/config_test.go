package config

import (
	"os"
	"testing"
)

func TestLoad_MissingToken(t *testing.T) {
	os.Clearenv()

	_, err := Load()
	if err == nil {
		t.Error("Load() should fail when TOKEN is missing")
	}
}

func TestLoad_WithToken(t *testing.T) {
	os.Clearenv()
	os.Setenv("TOKEN", "test-token")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Token != "test-token" {
		t.Errorf("Token = %q, want %q", cfg.Token, "test-token")
	}

	// Check defaults
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8080")
	}

	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "info")
	}

	if cfg.MaxTimeout != 120 {
		t.Errorf("MaxTimeout = %d, want %d", cfg.MaxTimeout, 120)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	os.Clearenv()
	os.Setenv("TOKEN", "test-token")
	os.Setenv("PORT", "9090")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("MAX_TIMEOUT", "60")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want %q", cfg.Port, "9090")
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}

	if cfg.MaxTimeout != 60 {
		t.Errorf("MaxTimeout = %d, want %d", cfg.MaxTimeout, 60)
	}
}
