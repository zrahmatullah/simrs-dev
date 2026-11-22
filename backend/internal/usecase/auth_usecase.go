package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"simrs-asoy/backend/internal/config"
	"simrs-asoy/backend/internal/domain"
	"simrs-asoy/backend/internal/repository"
	"simrs-asoy/backend/internal/security"
)

type AuthUsecase struct {
	userRepo  *repository.UserRepository
	auditRepo *repository.AuditRepository
	tokenRepo *repository.TokenRepository
	cfg       *config.Config
}

type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	RoleID   int    `json:"role_id"`
	RoleCode string `json:"role_code"`
	RoleName string `json:"role_name"`
	PoliID   *int   `json:"poli_id,omitempty"`
	jwt.RegisteredClaims
}

func NewAuthUsecase(userRepo *repository.UserRepository, auditRepo *repository.AuditRepository, tokenRepo *repository.TokenRepository, cfg *config.Config) *AuthUsecase {
	return &AuthUsecase{
		userRepo:  userRepo,
		auditRepo: auditRepo,
		tokenRepo: tokenRepo,
		cfg:       cfg,
	}
}

// generateAccessToken creates a short-lived (15 minutes) signed JWT access token.
func (u *AuthUsecase) generateAccessToken(user *domain.User) (string, error) {
	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		FullName: user.FullName,
		RoleID:   user.RoleID,
		RoleCode: user.RoleCode,
		RoleName: user.RoleName,
		PoliID:   user.PoliID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(security.AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "simrs-asoy",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(u.cfg.JWTSecret))
}

func (u *AuthUsecase) Login(req domain.LoginRequest, ipAddress, userAgent string) (*domain.LoginResponse, error) {
	// 1. Sanitize username input
	sanitizedUsername := security.SanitizePlainText(req.Username)

	user, err := u.userRepo.FindByUsername(sanitizedUsername)
	if err != nil {
		// Log failed login attempt
		_ = u.auditRepo.Log(&domain.AuditLog{
			Action:     "LOGIN_FAILED",
			EntityType: "user",
			Details:    fmt.Sprintf(`{"reason": "User not found", "attempted_username": "%s"}`, sanitizedUsername),
			IPAddress:  ipAddress,
			UserAgent:  userAgent,
		})
		return nil, errors.New("username atau password salah")
	}

	if !user.StatusAktif {
		return nil, errors.New("akun Anda sedang dinonaktifkan")
	}

	// 2. Constant-time password hash comparison
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		_ = u.auditRepo.Log(&domain.AuditLog{
			UserID:     &user.ID,
			Username:   user.Username,
			Action:     "LOGIN_FAILED",
			EntityType: "user",
			EntityID:   user.ID,
			Details:    `{"reason": "Invalid password"}`,
			IPAddress:  ipAddress,
			UserAgent:  userAgent,
		})
		return nil, errors.New("username atau password salah")
	}

	// 3. Generate short-lived Access Token (15 mins)
	accessToken, err := u.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat token akses: %w", err)
	}

	// 4. Generate cryptographically random Refresh Token (7 days)
	refreshToken, err := security.GenerateRandomToken(32)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat token refresh: %w", err)
	}

	// 5. Store hashed refresh token in repository
	refreshExpires := time.Now().Add(security.RefreshTokenDuration)
	_ = u.tokenRepo.StoreRefreshToken(user.ID, refreshToken, refreshExpires, ipAddress, userAgent)

	// 6. Audit Log successful login
	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &user.ID,
		Username:   user.Username,
		Action:     "LOGIN",
		EntityType: "user",
		EntityID:   user.ID,
		Details:    `{"message": "User logged in successfully (Dual Token: Access + Refresh)"}`,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	})

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(security.AccessTokenDuration.Seconds()),
		Token:        accessToken, // backward compatibility
		User:         user,
	}, nil
}

// RefreshToken implements OWASP Refresh Token Rotation (RTR).
// It revokes the old refresh token, validates against token reuse/theft, and issues a fresh pair.
func (u *AuthUsecase) RefreshToken(rawRefreshToken, ipAddress, userAgent string) (*domain.RefreshTokenResponse, error) {
	if rawRefreshToken == "" {
		return nil, errors.New("refresh token wajib disertakan")
	}

	// 1. Lookup token record by raw token hash
	rec, err := u.tokenRepo.FindByToken(rawRefreshToken)
	if err != nil {
		return nil, errors.New("refresh token tidak valid")
	}

	// 2. Check for token reuse / theft detection
	if rec.RevokedAt != nil {
		// ALARM: A revoked refresh token is being used! This indicates token interception/theft.
		// Revoke ALL user sessions for safety as recommended by OWASP.
		_ = u.tokenRepo.RevokeAllUserTokens(rec.UserID)
		_ = u.auditRepo.Log(&domain.AuditLog{
			UserID:     &rec.UserID,
			Action:     "SECURITY_ALERT_TOKEN_REUSE",
			EntityType: "auth",
			Details:    `{"warning": "Re-use of revoked refresh token detected. Invaliding all user sessions."}`,
			IPAddress:  ipAddress,
			UserAgent:  userAgent,
		})
		return nil, errors.New("terdeteksi penggunaan refresh token kedaluwarsa/tercabut. Sesi telah direset demi keamanan. Silakan login kembali.")
	}

	// 3. Check expiration
	if time.Now().After(rec.ExpiresAt) {
		return nil, errors.New("refresh token telah kedaluwarsa. Silakan login kembali.")
	}

	// 4. Fetch user
	user, err := u.userRepo.FindByID(rec.UserID)
	if err != nil || !user.StatusAktif {
		return nil, errors.New("pengguna tidak ditemukan atau telah dinonaktifkan")
	}

	// 5. Rotate Token: Revoke the old refresh token
	_ = u.tokenRepo.RevokeToken(rawRefreshToken)

	// 6. Issue new short-lived Access Token & new Refresh Token
	newAccessToken, err := u.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := security.GenerateRandomToken(32)
	if err != nil {
		return nil, err
	}

	// 7. Store new hashed refresh token
	newExpires := time.Now().Add(security.RefreshTokenDuration)
	_ = u.tokenRepo.StoreRefreshToken(user.ID, newRefreshToken, newExpires, ipAddress, userAgent)

	return &domain.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(security.AccessTokenDuration.Seconds()),
		Token:        newAccessToken,
	}, nil
}

// Logout terminates the session and revokes the refresh token.
func (u *AuthUsecase) Logout(rawRefreshToken, userID, ipAddress, userAgent string) error {
	if rawRefreshToken != "" {
		_ = u.tokenRepo.RevokeToken(rawRefreshToken)
	} else if userID != "" {
		_ = u.tokenRepo.RevokeAllUserTokens(userID)
	}

	var uidPtr *string
	if userID != "" {
		uidPtr = &userID
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     uidPtr,
		Action:     "LOGOUT",
		EntityType: "user",
		Details:    `{"message": "User session terminated and refresh token revoked"}`,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	})

	return nil
}
