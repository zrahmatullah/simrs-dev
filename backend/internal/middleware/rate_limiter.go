package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"simrs-asoy/backend/internal/domain"
)

type clientRecord struct {
	count       int
	windowStart time.Time
}

type IPRateLimiter struct {
	mu           sync.Mutex
	clients      map[string]*clientRecord
	limit        int           // Maximum requests
	window       time.Duration // Time window (e.g. 1 minute)
	cleanupTimer *time.Ticker
}

func NewIPRateLimiter(limit int, window time.Duration) *IPRateLimiter {
	limiter := &IPRateLimiter{
		clients:      make(map[string]*clientRecord),
		limit:        limit,
		window:       window,
		cleanupTimer: time.NewTicker(5 * time.Minute),
	}

	// Periodically purge old IP records to prevent memory leak
	go func() {
		for range limiter.cleanupTimer.C {
			limiter.mu.Lock()
			now := time.Now()
			for ip, rec := range limiter.clients {
				if now.Sub(rec.windowStart) > limiter.window*2 {
					delete(limiter.clients, ip)
				}
			}
			limiter.mu.Unlock()
		}
	}()

	return limiter
}

func (l *IPRateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	rec, exists := l.clients[ip]
	if !exists || now.Sub(rec.windowStart) > l.window {
		l.clients[ip] = &clientRecord{
			count:       1,
			windowStart: now,
		}
		return true
	}

	if rec.count >= l.limit {
		return false
	}

	rec.count++
	return true
}

func extractIP(r *http.Request) string {
	// Check X-Forwarded-For
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	// Fallback to RemoteAddr (strip port)
	remote := r.RemoteAddr
	if idx := strings.LastIndex(remote, ":"); idx != -1 {
		return remote[:idx]
	}
	return remote
}

// RateLimit returns a middleware that limits requests by client IP.
func RateLimit(limiter *IPRateLimiter, customMessage string) func(http.Handler) http.Handler {
	if customMessage == "" {
		customMessage = "Terlalu banyak permintaan (Rate limit exceeded). Silakan coba beberapa saat lagi."
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := extractIP(r)

			if !limiter.Allow(clientIP) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(domain.ResponseEnvelope{
					Success: false,
					Message: customMessage,
					Data:    nil,
					Errors:  map[string]string{"rate_limit": "Batas frekuensi permintaan terlampaui"},
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
