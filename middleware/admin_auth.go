package middleware

import (
	"crypto/subtle"
	"net/http"
)

// AdminAuthMiddleware protects the admin UI with HTTP Basic auth. Any username
// is accepted; the password must equal the fixed admin token. Basic auth keeps
// the browser flow simple (native login prompt, cached credentials) without
// sessions or cookies.
func AdminAuthMiddleware(adminToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, pass, ok := r.BasicAuth()
			if !ok || subtle.ConstantTimeCompare([]byte(pass), []byte(adminToken)) != 1 {
				w.Header().Set("WWW-Authenticate", `Basic realm="impersonate-admin"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
