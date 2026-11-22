package middleware

import (
	"encoding/json"
	"net/http"

	"simrs-asoy/backend/internal/domain"
)

// RequireRoles enforces RBAC as specified in AGENTS.md Matrix
func RequireRoles(allowedRoles ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetCurrentUser(r)
			if claims == nil {
				respondForbidden(w, "Akses ditolak: pengguna belum diautentikasi")
				return
			}

			// SUPERADMIN has full access to all endpoints
			if claims.RoleCode == "SUPERADMIN" {
				next.ServeHTTP(w, r)
				return
			}

			for _, role := range allowedRoles {
				if claims.RoleCode == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			respondForbidden(w, "Akses ditolak: role Anda ("+claims.RoleCode+") tidak memiliki wewenang untuk tindakan ini")
		})
	}
}

func respondForbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(domain.ResponseEnvelope{
		Success: false,
		Message: message,
		Data:    nil,
		Errors:  map[string]string{"rbac": message},
	})
}
