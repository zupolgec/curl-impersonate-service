package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// adminSessionCookie holds a signed admin session: "<expiry unix>.<hmac>".
// The HMAC is keyed with the admin token, so rotating ADMIN_TOKEN logs out
// every session without any server-side state.
const (
	adminSessionCookie = "impersonate_admin"
	adminSessionTTL    = 30 * 24 * time.Hour
)

func signSession(adminToken string, expiry int64) string {
	exp := strconv.FormatInt(expiry, 10)
	mac := hmac.New(sha256.New, []byte(adminToken))
	mac.Write([]byte("admin-session:" + exp))
	return exp + "." + hex.EncodeToString(mac.Sum(nil))
}

// SetAdminSession issues a new admin session cookie.
func SetAdminSession(w http.ResponseWriter, r *http.Request, adminToken string) {
	expiry := time.Now().Add(adminSessionTTL)
	http.SetCookie(w, &http.Cookie{
		Name:     adminSessionCookie,
		Value:    signSession(adminToken, expiry.Unix()),
		Path:     "/",
		Expires:  expiry,
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearAdminSession removes the admin session cookie.
func ClearAdminSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: adminSessionCookie, Value: "", Path: "/", MaxAge: -1})
}

// CheckAdminPassword reports whether pass equals the admin token.
func CheckAdminPassword(adminToken, pass string) bool {
	return subtle.ConstantTimeCompare([]byte(pass), []byte(adminToken)) == 1
}

// IsAdmin reports whether the request carries a valid admin session cookie or
// Basic auth credentials (kept for scripts and curl).
func IsAdmin(r *http.Request, adminToken string) bool {
	if _, pass, ok := r.BasicAuth(); ok && CheckAdminPassword(adminToken, pass) {
		return true
	}
	c, err := r.Cookie(adminSessionCookie)
	if err != nil {
		return false
	}
	exp, _, ok := strings.Cut(c.Value, ".")
	if !ok {
		return false
	}
	expiry, err := strconv.ParseInt(exp, 10, 64)
	if err != nil || time.Now().Unix() > expiry {
		return false
	}
	return hmac.Equal([]byte(c.Value), []byte(signSession(adminToken, expiry)))
}

// AdminAuthMiddleware protects the admin UI. Unauthenticated requests are sent
// to the login page, remembering where they were headed.
func AdminAuthMiddleware(adminToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !IsAdmin(r, adminToken) {
				http.Redirect(w, r, "/admin/login?next="+url.QueryEscape(r.URL.RequestURI()), http.StatusSeeOther)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
