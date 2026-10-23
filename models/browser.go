package models

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	browserAliases map[string]string
	defaultBrowser string
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
	browserAliases = latestAliases(browsersData.Browsers)
	defaultBrowser = browserAliases["chrome-latest"]

	return nil
}

// latestAliases maps "<browser>-latest" to the highest desktop (no device)
// version of each browser. On equal versions the plain profile wins over an
// OS-suffixed one (chrome152 over chrome152_macos).
func latestAliases(browsers []BrowserConfig) map[string]string {
	best := map[string]BrowserConfig{}
	for _, b := range browsers {
		if b.Browser.Device != "" {
			continue
		}
		key := b.Browser.Name + "-latest"
		cur, ok := best[key]
		if !ok {
			best[key] = b
			continue
		}
		switch c := compareVersions(b.Browser.Version, cur.Browser.Version); {
		case c > 0, c == 0 && len(b.Name) < len(cur.Name):
			best[key] = b
		}
	}
	aliases := make(map[string]string, len(best))
	for k, b := range best {
		aliases[k] = b.Name
	}
	return aliases
}

// compareVersions compares dotted numeric versions ("26.0.1" vs "27").
func compareVersions(a, b string) int {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var x, y int
		if i < len(pa) {
			x, _ = strconv.Atoi(pa[i])
		}
		if i < len(pb) {
			y, _ = strconv.Atoi(pb[i])
		}
		if x != y {
			if x > y {
				return 1
			}
			return -1
		}
	}
	return 0
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
