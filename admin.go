package server

import (
	"crypto/subtle"
	"log/slog"
	"net/http"

	"github.com/simonhege/server/ip"
)

// Admin is a middleware that checks if the request has a valid admin API key.
func Admin(adminKey string) func(next http.Handler) http.Handler {

	adminKeyBytes := []byte(adminKey)

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			apiKeyHeader := []byte(r.Header.Get("X-Api-Key"))
			if subtle.ConstantTimeCompare(apiKeyHeader, adminKeyBytes) == 0 {
				slog.WarnContext(r.Context(), "Invalid API key", "method", r.Method, "url", r.URL.String(), "ip", ip.Get(r))
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
