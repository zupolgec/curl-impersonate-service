package models

import (
	"os"
	"path/filepath"
	"testing"
)

const fixtureBrowsers = `{"browsers":[
 {"name":"chrome99","browser":{"name":"chrome","version":"99.0.4844.51","os":"win10"}},
 {"name":"chrome100","browser":{"name":"chrome","version":"100.0.4896.75","os":"win10"}},
 {"name":"chrome152","browser":{"name":"chrome","version":"152","os":"win10"}},
 {"name":"chrome152_macos","browser":{"name":"chrome","version":"152","os":"macos"}},
 {"name":"chrome160_android","browser":{"name":"chrome","version":"160.0.0.0","os":"android","device":"pixel6"}},
 {"name":"firefox147","browser":{"name":"firefox","version":"147.0","os":"macos"}},
 {"name":"firefox155","browser":{"name":"firefox","version":"155","os":"win10"}},
 {"name":"safari2601","browser":{"name":"safari","version":"26.0.1","os":"macos"}},
 {"name":"safari270","browser":{"name":"safari","version":"27.0","os":"macos"}},
 {"name":"safari280_ios","browser":{"name":"safari","version":"28.0","os":"ios","device":"iphone"}},
 {"name":"edge101","browser":{"name":"edge","version":"101.0.1210.47","os":"win10"}}
]}`

func loadFixture(t *testing.T) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "browsers.json")
	if err := os.WriteFile(path, []byte(fixtureBrowsers), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := LoadBrowsers(path); err != nil {
		t.Fatal(err)
	}
}

// Aliases follow browsers.json: the highest desktop version of each browser,
// preferring the plain profile over an OS-suffixed one of the same version.
func TestAliasesDerivedFromBrowsers(t *testing.T) {
	loadFixture(t)
	want := map[string]string{
		"chrome-latest":  "chrome152",
		"firefox-latest": "firefox155",
		"safari-latest":  "safari270",
		"edge-latest":    "edge101",
	}
	got := GetAliases()
	if len(got) != len(want) {
		t.Fatalf("aliases = %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, want %q", k, got[k], v)
		}
	}
	if d := GetDefaultBrowser(); d != "chrome152" {
		t.Errorf("default = %q, want chrome152", d)
	}
}

func TestResolveBrowserName(t *testing.T) {
	loadFixture(t)
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty returns default", "", "chrome152"},
		{"chrome-latest alias", "chrome-latest", "chrome152"},
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
