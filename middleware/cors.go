package middleware

import "net/http"

// OriginProvider returns the currently allowed CORS origins. Returning a slice
// containing "*" allows any origin. It is a function so the source can be
// static config today and admin-managed settings later.
type OriginProvider func() []string

// CORSMiddleware adds CORS headers, restricting the allowed origin to the
// configured list.
func CORSMiddleware(origins OriginProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if allowed := resolveAllowedOrigin(origins(), r.Header.Get("Origin")); allowed != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowed)
				if allowed != "*" {
					w.Header().Add("Vary", "Origin")
				}
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// resolveAllowedOrigin returns the value to send in Access-Control-Allow-Origin,
// or "" when the request origin is not allowed.
func resolveAllowedOrigin(allowed []string, reqOrigin string) string {
	for _, o := range allowed {
		if o == "*" {
			return "*"
		}
		if reqOrigin != "" && o == reqOrigin {
			return reqOrigin
		}
	}
	return ""
}
