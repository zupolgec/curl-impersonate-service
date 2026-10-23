package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zupolgec/curl-impersonate-service/metrics"
	"github.com/zupolgec/curl-impersonate-service/store"
)

func newTestAdmin(t *testing.T) (http.Handler, *store.Store) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "admin.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return NewAdminHandler(st, metrics.NewCollector(), true, 72), st
}

func TestAdminDashboardRenders(t *testing.T) {
	h, _ := newTestAdmin(t)
	req := httptest.NewRequest(http.MethodGet, "/admin/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Dashboard") {
		t.Fatal("dashboard body missing nav")
	}
}

func TestAdminCreatesToken(t *testing.T) {
	h, st := newTestAdmin(t)

	form := url.Values{"name": {"my-token"}}
	req := httptest.NewRequest(http.MethodPost, "/admin/tokens", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", w.Code)
	}
	toks, _ := st.ListTokens()
	if len(toks) != 1 || toks[0].Name != "my-token" {
		t.Fatalf("token not created: %+v", toks)
	}
}

func TestAdminCORSProviderReflectsSetting(t *testing.T) {
	h, st := newTestAdmin(t)
	provider := CORSOriginProvider(st, []string{"*"})

	form := url.Values{"origins": {"https://a.com, https://b.com"}}
	req := httptest.NewRequest(http.MethodPost, "/admin/cors", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(httptest.NewRecorder(), req)

	got := provider()
	if len(got) != 2 || got[0] != "https://a.com" || got[1] != "https://b.com" {
		t.Fatalf("provider() = %v, want [https://a.com https://b.com]", got)
	}
}

func TestFormatUptime(t *testing.T) {
	cases := map[int64]string{
		0:       "0s",
		42:      "42s",
		61:      "1m",
		3720:    "1h 2m",
		90061:   "1d 1h",
		2592000: "30d 0h",
	}
	for in, want := range cases {
		if got := formatUptime(in); got != want {
			t.Errorf("formatUptime(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestAdminLogin(t *testing.T) {
	const admin = "admin-secret"
	h := NewAdminLoginHandler(admin)

	// Login page renders.
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin/login", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `type="password"`) {
		t.Fatalf("login page: status %d", w.Code)
	}

	post := func(pass, next string) *httptest.ResponseRecorder {
		form := url.Values{"password": {pass}, "next": {next}}
		req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		return w
	}

	// Wrong password: no cookie, error shown.
	w = post("nope", "")
	if w.Code != http.StatusUnauthorized || len(w.Result().Cookies()) != 0 {
		t.Fatalf("wrong password: status %d, cookies %d", w.Code, len(w.Result().Cookies()))
	}

	// Right password: cookie set, redirected to next.
	w = post(admin, "/admin/logs")
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/admin/logs" {
		t.Fatalf("login: status %d, location %q", w.Code, w.Header().Get("Location"))
	}
	if len(w.Result().Cookies()) != 1 {
		t.Fatal("login did not set a session cookie")
	}

	// Open redirects are refused.
	if loc := post(admin, "//evil.example").Header().Get("Location"); loc != "/admin/" {
		t.Fatalf("open redirect: location %q", loc)
	}

	// Logout clears the cookie.
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/admin/logout", nil))
	c := w.Result().Cookies()
	if w.Code != http.StatusSeeOther || len(c) != 1 || c[0].MaxAge >= 0 {
		t.Fatalf("logout: status %d, cookies %+v", w.Code, c)
	}
}

// Dashboard counters come from the stored history, so they survive restarts
// (a fresh collector here stands in for a restarted process).
func TestAdminDashboardShowsStoredHistory(t *testing.T) {
	h, st := newTestAdmin(t)
	for _, ok := range []bool{true, true, false} {
		if err := st.AddLog(store.LogEntry{TokenName: "a", Browser: "chrome152", Method: "GET", Success: ok, DurationMs: 30}); err != nil {
			t.Fatal(err)
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin/", nil))
	body := w.Body.String()
	for _, want := range []string{
		`<div class="n">3</div><div class="l">Requests</div>`,
		`<div class="n ok">2</div>`,
		`<div class="n bad">1</div>`,
		"Last 30 days",
		time.Now().UTC().Format("2006-01-02"),
	} {
		if !strings.Contains(body, want) {
			t.Errorf("dashboard missing %q", want)
		}
	}
}
