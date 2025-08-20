package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	mux     *http.ServeMux
	handler http.Handler
}

// New creates a new Server instance with the provided rate limit and buffer size.
func New(corsAllowCredentials bool, middlewares ...Middleware) *Server {
	serveMux := http.NewServeMux()
	s := &Server{
		mux:     serveMux,
		handler: serveMux,
	}

	// Default middlewares
	s.enableCors(corsAllowCredentials)
	s.handler = apply(s.handler, middlewares...)

	return s
}

func (s *Server) enableCors(allowCredentials bool) {
	s.mux.HandleFunc("OPTIONS /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	next := s.handler
	s.handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOrigins := "*"
		if origin := r.Header.Get("Origin"); origin != "" {
			allowedOrigins = origin
		}
		if allowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, PATCH, DELETE, OPTIONS")
		next.ServeHTTP(w, r)
	})
}

type Middleware func(http.Handler) http.Handler

func DefaultMiddlewares() []Middleware {
	return []Middleware{
		RequestID,
		RequestLogger,
	}
}

func apply(handler http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}

func (s *Server) HandleFunc(pattern string, handler http.HandlerFunc, mws ...Middleware) {
	s.mux.Handle(pattern, apply(handler, mws...))
}

func (s *Server) Handle(pattern string, handler http.Handler, mws ...Middleware) {
	s.mux.Handle(pattern, apply(handler, mws...))
}

func (s *Server) Run(ctx context.Context, addr string) error {

	server := http.Server{
		Addr:         addr,
		Handler:      s.handler,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		ctx := context.Background()
		slog.InfoContext(ctx, "Server started", "addr", addr)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.ErrorContext(ctx, "HTTP server error", "err", err)
			return
		}
		slog.InfoContext(ctx, "Server stopped serving new connections.")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	slog.InfoContext(ctx, "Server graceful shutdown requested.")

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("http shutdown error: %w", err)
	}
	slog.InfoContext(ctx, "Graceful shutdown complete.")
	return nil
}
