package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

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
	return NewAdminHandler(st, metrics.NewCollector()), st
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
