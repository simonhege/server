package server

import (
	"context"
	"log/slog"
	"math/rand"
	"net/http"
)

// Define a custom key type to avoid context key collisions.
type ctxKeyRequestID struct{}

// requestIDKey is the key used to set/retrieve the request ID.
var requestIDKey = ctxKeyRequestID{}

// generateRequestID creates a random 6-character alphanumeric string.
func generateRequestID() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// RequestID is a middleware that generates a unique request ID for each incoming HTTP request.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		requestId := generateRequestID()
		ctx := context.WithValue(r.Context(), requestIDKey, requestId)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Wrap is a function that wraps a slog.Handler to add request ID information to log records.
func Wrap(h slog.Handler) slog.Handler {
	return &requestIDHandler{
		handler: h,
	}
}

// requestIDHandler enriches log records with the request ID from the context.
type requestIDHandler struct {
	handler slog.Handler
}

// Enabled implements slog.Handler.
func (h *requestIDHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle implements slog.Handler.
func (h *requestIDHandler) Handle(ctx context.Context, r slog.Record) error {
	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		r.Add("request_id", reqID)
	}
	return h.handler.Handle(ctx, r)
}

// WithAttrs implements slog.Handler.
func (h *requestIDHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &requestIDHandler{handler: h.handler.WithAttrs(attrs)}
}

// WithGroup implements slog.Handler.
func (h *requestIDHandler) WithGroup(name string) slog.Handler {
	return &requestIDHandler{handler: h.handler.WithGroup(name)}
}

var _ slog.Handler = (*requestIDHandler)(nil)
