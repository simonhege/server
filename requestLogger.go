package server

import (
	"log/slog"
	"net/http"

	"github.com/felixge/httpsnoop"
	"github.com/simonhege/server/ip"
)

// RequestLogger is a middleware that logs details of each HTTP request.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		caller_ip := ip.Get(r)

		metrics := httpsnoop.CaptureMetrics(next, w, r)

		slog.InfoContext(r.Context(), "Request executed",
			"method", r.Method,
			"url", r.URL.String(),
			"code", metrics.Code,
			"written", metrics.Written,
			"duration", metrics.Duration.Seconds(),
			"ip", caller_ip)

	})
}
