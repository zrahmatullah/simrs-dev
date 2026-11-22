package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

const (
	// AccessTokenDuration sets short-lived access tokens (15 minutes as per OWASP recommendation)
	AccessTokenDuration = 15 * time.Minute
	// RefreshTokenDuration sets 7 days session duration
	RefreshTokenDuration = 7 * 24 * time.Hour
)

// GenerateRandomToken produces a cryptographically secure random string with 256 bits of entropy.
func GenerateRandomToken(byteLen int) (string, error) {
	if byteLen <= 0 {
		byteLen = 32 // 256 bits
	}
	bytes := make([]byte, byteLen)
	if _, err := rand.Read(bytes); err != nil {
		return "", errors.New("gagal menghasilkan token acak kriptografis: " + err.Error())
	}
	return hex.EncodeToString(bytes), nil
}

// HashToken computes the SHA-256 checksum of a raw refresh token.
// Only this hash is stored in persistent storage to protect sessions in case of DB leakage.
func HashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
