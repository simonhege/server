package server

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/simonhege/server/ip"
	"golang.org/x/time/rate"
)

type rateLimiter struct {
	keys  map[string]*rate.Limiter
	mu    *sync.RWMutex
	rate  rate.Limit
	burst int
}

func newRateLimiter(r rate.Limit, b int) *rateLimiter {
	i := &rateLimiter{
		keys:  make(map[string]*rate.Limiter),
		mu:    &sync.RWMutex{},
		rate:  r,
		burst: b,
	}

	return i
}

// Add creates a new rate limiter for a key and adds it to the internal map
func (i *rateLimiter) Add(key string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.rate, i.burst)

	i.keys[key] = limiter

	return limiter
}

// GetLimiter returns the rate limiter for the provided key if it exists.
// Otherwise calls Add to add key to the map
func (i *rateLimiter) GetLimiter(key string) *rate.Limiter {
	i.mu.Lock()
	limiter, exists := i.keys[key]

	if !exists {
		i.mu.Unlock()
		return i.Add(key)
	}

	i.mu.Unlock()

	return limiter
}

// RateLimiter is a middleware that limits the number of requests from a single IP address.
func RateLimiter(r rate.Limit, b int) func(next http.Handler) http.Handler {

	var limiter = newRateLimiter(r, b)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			caller_ip := ip.Get(r)
			limiter := limiter.GetLimiter(caller_ip)
			if !limiter.Allow() {
				slog.WarnContext(r.Context(), "Too Many Requests", "method", r.Method, "url", r.URL.String(), "ip", caller_ip)
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

}
