package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/zupolgec/curl-impersonate-service/models"
)

// tokensEqual compares two tokens in constant time to avoid leaking the token
// through response-timing differences.
func tokensEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func AuthMiddleware(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check query parameter first
			queryToken := r.URL.Query().Get("token")
			if queryToken != "" {
				if tokensEqual(queryToken, token) {
					next.ServeHTTP(w, r)
					return
				}
				models.WriteJSONError(w, http.StatusUnauthorized, "auth", "invalid authentication token")
				return
			}

			// Check Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				models.WriteJSONError(w, http.StatusUnauthorized, "auth", "missing authentication token")
				return
			}

			// Check Bearer token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				models.WriteJSONError(w, http.StatusUnauthorized, "auth", "invalid authorization header format")
				return
			}

			if !tokensEqual(parts[1], token) {
				models.WriteJSONError(w, http.StatusUnauthorized, "auth", "invalid authentication token")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
