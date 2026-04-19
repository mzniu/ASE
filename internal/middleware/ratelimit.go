package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sync"

	"github.com/example/ase/internal/config"
	"github.com/example/ase/internal/httpx"
	"golang.org/x/time/rate"
)

// RateLimit enforces global and per-credential QPS (REQ-F-011).
func RateLimit(cfg config.Config) func(http.Handler) http.Handler {
	global := rate.NewLimiter(rate.Limit(cfg.RateLimitGlobal), cfg.RateLimitGlobalBurst)
	var mu sync.Mutex
	per := make(map[string]*rate.Limiter)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !global.Allow() {
				httpx.WriteProblem(w, http.StatusTooManyRequests, "rate limit", "global limit exceeded")
				return
			}
			key := clientKey(r)
			lim := getLimiter(&mu, per, key, rate.Limit(cfg.RateLimitPerKey), cfg.RateLimitBurst)
			if !lim.Allow() {
				httpx.WriteProblem(w, http.StatusTooManyRequests, "rate limit", "per-key limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientKey(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "ip:" + r.RemoteAddr
	}
	sum := sha256.Sum256([]byte(h))
	return "auth:" + hex.EncodeToString(sum[:8])
}

func getLimiter(mu *sync.Mutex, m map[string]*rate.Limiter, key string, every rate.Limit, burst int) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()
	if lim, ok := m[key]; ok {
		return lim
	}
	lim := rate.NewLimiter(every, burst)
	m[key] = lim
	return lim
}
