package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zupolgec/curl-impersonate-service/models"
)

func loadRealBrowsers(t *testing.T) {
	t.Helper()
	if err := models.LoadBrowsers("../browsers.json"); err != nil {
		t.Fatal(err)
	}
}

func get(h http.Handler, path, accept string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Host = "imp.example.org"
	req.Header.Set("X-Forwarded-Proto", "https")
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func TestDocsNegotiatesMarkdown(t *testing.T) {
	loadRealBrowsers(t)
	h := NewDocsHandler(true)

	cases := []struct {
		name, path, accept, wantType string
	}{
		{"browser gets html", "/docs", "text/html,application/xhtml+xml,*/*;q=0.8", "text/html"},
		{"no accept gets html", "/docs", "", "text/html"},
		{"agent asking markdown", "/docs", "text/markdown", "text/markdown"},
		{"markdown preferred among others", "/docs", "text/markdown, text/html;q=0.9", "text/markdown"},
		{"docs.md is always markdown", "/docs.md", "", "text/markdown"},
		{"llms.txt is always markdown", "/llms.txt", "text/html", "text/markdown"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := get(h, tc.path, tc.accept)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d", w.Code)
			}
			if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, tc.wantType) {
				t.Fatalf("Content-Type = %q, want %s", ct, tc.wantType)
			}
			body := w.Body.String()
			// Examples use the real host the docs were served from.
			if !strings.Contains(body, "https://imp.example.org/impersonate") {
				t.Error("examples should use the request host")
			}
			if strings.Contains(body, "localhost") {
				t.Error("examples should not mention localhost")
			}
			if tc.wantType == "text/markdown" && !strings.Contains(body, "chrome-latest") {
				t.Error("markdown docs should list aliases")
			}
		})
	}
	if v := get(h, "/docs", "").Header().Get("Vary"); v != "Accept" {
		t.Errorf("Vary = %q, want Accept", v)
	}
}

func TestHomePage(t *testing.T) {
	h := NewHomeHandler(true, true)

	w := get(h, "/", "text/html")
	body := w.Body.String()
	if w.Code != http.StatusOK || !strings.Contains(body, `href="/docs"`) || !strings.Contains(body, `href="/admin/"`) {
		t.Fatalf("home html: status %d body %q", w.Code, body)
	}

	w = get(h, "/", "text/markdown")
	if !strings.HasPrefix(w.Header().Get("Content-Type"), "text/markdown") ||
		!strings.Contains(w.Body.String(), "https://imp.example.org/docs.md") {
		t.Fatalf("home markdown: %q", w.Body.String())
	}

	// Without docs or admin, their links disappear.
	body = get(NewHomeHandler(false, false), "/", "").Body.String()
	if strings.Contains(body, `href="/docs"`) || strings.Contains(body, `href="/admin/"`) {
		t.Fatal("disabled sections must not be linked")
	}
}
