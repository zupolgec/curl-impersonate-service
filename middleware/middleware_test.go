package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestAuthMiddleware(t *testing.T) {
	const token = "secret-token"
	validate := func(t string) (string, bool) {
		if t == token {
			return "test", true
		}
		return "", false
	}
	h := AuthMiddleware(validate)(okHandler())

	cases := []struct {
		name       string
		setup      func(*http.Request)
		wantStatus int
	}{
		{"valid bearer", func(r *http.Request) { r.Header.Set("Authorization", "Bearer "+token) }, http.StatusOK},
		{"valid query", func(r *http.Request) { r.URL.RawQuery = "token=" + token }, http.StatusOK},
		{"wrong bearer", func(r *http.Request) { r.Header.Set("Authorization", "Bearer nope") }, http.StatusUnauthorized},
		{"wrong query", func(r *http.Request) { r.URL.RawQuery = "token=nope" }, http.StatusUnauthorized},
		{"missing", func(r *http.Request) {}, http.StatusUnauthorized},
		{"malformed header", func(r *http.Request) { r.Header.Set("Authorization", "Basic abc") }, http.StatusUnauthorized},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "http://x/", nil)
			tc.setup(r)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tc.wantStatus {
				t.Fatalf("got %d, want %d", w.Code, tc.wantStatus)
			}
		})
	}
}

func TestResolveAllowedOrigin(t *testing.T) {
	cases := []struct {
		name    string
		allowed []string
		origin  string
		want    string
	}{
		{"wildcard", []string{"*"}, "https://a.com", "*"},
		{"match", []string{"https://a.com"}, "https://a.com", "https://a.com"},
		{"no match", []string{"https://a.com"}, "https://b.com", ""},
		{"empty origin non-wildcard", []string{"https://a.com"}, "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveAllowedOrigin(tc.allowed, tc.origin); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCORSMiddlewarePreflight(t *testing.T) {
	h := CORSMiddleware(func() []string { return []string{"*"} })(okHandler())
	r := httptest.NewRequest(http.MethodOptions, "http://x/", nil)
	r.Header.Set("Origin", "https://a.com")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent {
		t.Fatalf("preflight got %d, want 204", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("ACAO got %q, want *", got)
	}
}
