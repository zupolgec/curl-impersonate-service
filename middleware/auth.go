package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/zupolgec/curl-impersonate-service/models"
)

// TokenValidator validates an API token value and returns the associated token
// name and whether it is valid.
type TokenValidator func(token string) (name string, ok bool)

type contextKey string

const tokenNameKey contextKey = "tokenName"

// TokenName returns the authenticated token name stored in the request context.
func TokenName(ctx context.Context) string {
	if v, ok := ctx.Value(tokenNameKey).(string); ok {
		return v
	}
	return ""
}

// AuthMiddleware authenticates API requests via a Bearer token or a `token`
// query parameter, validating against the provided validator.
func AuthMiddleware(validate TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, provided := extractToken(r)
			if !provided {
				models.WriteJSONError(w, http.StatusUnauthorized, "auth", "missing authentication token")
				return
			}

			name, ok := validate(token)
			if !ok {
				models.WriteJSONError(w, http.StatusUnauthorized, "auth", "invalid authentication token")
				return
			}

			ctx := context.WithValue(r.Context(), tokenNameKey, name)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractToken pulls the token from the query string or Authorization header.
func extractToken(r *http.Request) (string, bool) {
	if q := r.URL.Query().Get("token"); q != "" {
		return q, true
	}
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", false
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", false
	}
	return parts[1], true
}
