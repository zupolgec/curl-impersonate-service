package models

import (
	"testing"
)

func TestResolveBrowserName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty returns default", "", "chrome136"},
		{"chrome-latest alias", "chrome-latest", "chrome136"},
		{"firefox-latest alias", "firefox-latest", "firefox135"},
		{"edge-latest alias", "edge-latest", "edge101"},
		{"safari-latest alias", "safari-latest", "safari260"},
		{"tor-latest alias", "tor-latest", "tor145"},
		{"direct browser name", "chrome99", "chrome99"},
		{"unknown browser passthrough", "unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveBrowserName(tt.input)
			if result != tt.expected {
				t.Errorf("ResolveBrowserName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetAliases(t *testing.T) {
	aliases := GetAliases()

	if len(aliases) == 0 {
		t.Error("GetAliases() returned empty map")
	}

	if aliases["chrome-latest"] != "chrome136" {
		t.Errorf("chrome-latest alias incorrect, got %q", aliases["chrome-latest"])
	}
}

func TestGetDefaultBrowser(t *testing.T) {
	defaultBrowser := GetDefaultBrowser()
	if defaultBrowser != "chrome136" {
		t.Errorf("GetDefaultBrowser() = %q, want %q", defaultBrowser, "chrome136")
	}
}
