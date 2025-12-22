package models

import (
	"encoding/json"
	"fmt"
	"os"
)

type BrowserInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	OS      string `json:"os"`
	Device  string `json:"device,omitempty"`
}

type BrowserConfig struct {
	Name          string      `json:"name"`
	Browser       BrowserInfo `json:"browser"`
	Binary        string      `json:"binary"`
	WrapperScript string      `json:"wrapper_script"`
}

type BrowsersData struct {
	Browsers []BrowserConfig `json:"browsers"`
}

var (
	browsersCache  map[string]BrowserConfig
	browserAliases = map[string]string{
		"chrome-latest":  "chrome116",
		"firefox-latest": "ff117",
		"edge-latest":    "edge101",
		"safari-latest":  "safari15_5",
	}
	defaultBrowser = "chrome116"
)

// LoadBrowsers loads and caches browsers.json
func LoadBrowsers(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read browsers.json: %w", err)
	}

	var browsersData BrowsersData
	if err := json.Unmarshal(data, &browsersData); err != nil {
		return fmt.Errorf("failed to parse browsers.json: %w", err)
	}

	browsersCache = make(map[string]BrowserConfig)
	for _, browser := range browsersData.Browsers {
		browsersCache[browser.Name] = browser
	}

	return nil
}

// ResolveBrowserName resolves aliases and returns the actual browser name
func ResolveBrowserName(name string) string {
	if name == "" {
		return defaultBrowser
	}
	if alias, ok := browserAliases[name]; ok {
		return alias
	}
	return name
}

// GetBrowserConfig returns the browser configuration
func GetBrowserConfig(name string) (BrowserConfig, error) {
	resolved := ResolveBrowserName(name)
	config, ok := browsersCache[resolved]
	if !ok {
		return BrowserConfig{}, fmt.Errorf("unknown browser: %s", name)
	}
	return config, nil
}

// GetAllBrowsers returns all available browsers
func GetAllBrowsers() []BrowserConfig {
	browsers := make([]BrowserConfig, 0, len(browsersCache))
	for _, browser := range browsersCache {
		browsers = append(browsers, browser)
	}
	return browsers
}

// GetAliases returns all browser aliases
func GetAliases() map[string]string {
	return browserAliases
}

// GetDefaultBrowser returns the default browser name
func GetDefaultBrowser() string {
	return defaultBrowser
}
