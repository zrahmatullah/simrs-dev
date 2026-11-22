package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"simrs-asoy/backend/internal/config"
	"simrs-asoy/backend/internal/domain"
	"simrs-asoy/backend/internal/usecase"
)

type contextKey string

const (
	UserContextKey contextKey = "currentUser"
)

func Auth(cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondUnauthorized(w, "Token otentikasi tidak ditemukan")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondUnauthorized(w, "Format header otentikasi tidak valid")
				return
			}

			tokenString := parts[1]
			claims := &usecase.JWTClaims{}

			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				// OWASP: Enforce HMAC-SHA256 method to eliminate 'none' algorithm and key confusion attacks
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("metode penandatanganan token tidak valid")
				}
				if token.Header["alg"] != "HS256" {
					return nil, errors.New("algoritma token harus HS256")
				}
				return []byte(cfg.JWTSecret), nil
			})

			if err != nil || !token.Valid {
				respondUnauthorized(w, "Sesi Anda telah kedaluwarsa atau token tidak valid. Silakan gunakan refresh token atau login ulang.")
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetCurrentUser(r *http.Request) *usecase.JWTClaims {
	claims, ok := r.Context().Value(UserContextKey).(*usecase.JWTClaims)
	if !ok {
		return nil
	}
	return claims
}

func respondUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(domain.ResponseEnvelope{
		Success: false,
		Message: message,
		Data:    nil,
		Errors:  map[string]string{"auth": message},
	})
}
