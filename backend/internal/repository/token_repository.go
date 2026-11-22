package repository

import (
	"database/sql"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"simrs-asoy/backend/internal/security"
)

type RefreshTokenRecord struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	TokenHash string     `db:"token_hash"`
	ExpiresAt time.Time  `db:"expires_at"`
	RevokedAt *time.Time `db:"revoked_at"`
	IPAddress string     `db:"ip_address"`
	UserAgent string     `db:"user_agent"`
	CreatedAt time.Time  `db:"created_at"`
}

type TokenRepository struct {
	db       *sqlx.DB
	memStore map[string]*RefreshTokenRecord // Fallback memory store
	mu       sync.RWMutex
}

func NewTokenRepository(db *sqlx.DB) *TokenRepository {
	repo := &TokenRepository{
		db:       db,
		memStore: make(map[string]*RefreshTokenRecord),
	}
	repo.ensureSchema()
	return repo
}

func (r *TokenRepository) ensureSchema() {
	schema := `
	CREATE TABLE IF NOT EXISTS user_refresh_tokens (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token_hash VARCHAR(64) NOT NULL UNIQUE,
		expires_at TIMESTAMP NOT NULL,
		revoked_at TIMESTAMP NULL,
		ip_address VARCHAR(45) NULL,
		user_agent TEXT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_refresh_token_hash ON user_refresh_tokens(token_hash);
	CREATE INDEX IF NOT EXISTS idx_refresh_user_id ON user_refresh_tokens(user_id);
	`
	if _, err := r.db.Exec(schema); err != nil {
		log.Printf("[SECURITY WARNING] Could not auto-create user_refresh_tokens table: %v. Using in-memory fallback.", err)
	}
}

// StoreRefreshToken saves a new refresh token (hashed) for a user session.
func (r *TokenRepository) StoreRefreshToken(userID, rawToken string, expiresAt time.Time, ip, ua string) error {
	tokenHash := security.HashToken(rawToken)

	query := `
		INSERT INTO user_refresh_tokens (user_id, token_hash, expires_at, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`
	_, err := r.db.Exec(query, userID, tokenHash, expiresAt, ip, ua)
	if err != nil {
		// Store in memory fallback
		r.mu.Lock()
		r.memStore[tokenHash] = &RefreshTokenRecord{
			UserID:    userID,
			TokenHash: tokenHash,
			ExpiresAt: expiresAt,
			IPAddress: ip,
			UserAgent: ua,
			CreatedAt: time.Now(),
		}
		r.mu.Unlock()
		return nil
	}
	return nil
}

// FindByToken retrieves the token record by raw token and verifies it is not revoked or expired.
func (r *TokenRepository) FindByToken(rawToken string) (*RefreshTokenRecord, error) {
	tokenHash := security.HashToken(rawToken)

	query := `
		SELECT id, user_id, token_hash, expires_at, revoked_at, ip_address, user_agent, created_at
		FROM user_refresh_tokens
		WHERE token_hash = $1
		LIMIT 1
	`
	var rec RefreshTokenRecord
	err := r.db.Get(&rec, query, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Check memory fallback
			r.mu.RLock()
			memRec, ok := r.memStore[tokenHash]
			r.mu.RUnlock()
			if ok {
				return memRec, nil
			}
			return nil, errors.New("refresh token tidak ditemukan atau tidak valid")
		}
		// Try memory fallback
		r.mu.RLock()
		memRec, ok := r.memStore[tokenHash]
		r.mu.RUnlock()
		if ok {
			return memRec, nil
		}
		return nil, err
	}

	return &rec, nil
}

// RevokeToken marks a token as revoked (used during token rotation or logout).
func (r *TokenRepository) RevokeToken(rawToken string) error {
	tokenHash := security.HashToken(rawToken)
	now := time.Now()

	query := `UPDATE user_refresh_tokens SET revoked_at = $1 WHERE token_hash = $2`
	_, _ = r.db.Exec(query, now, tokenHash)

	r.mu.Lock()
	if rec, ok := r.memStore[tokenHash]; ok {
		rec.RevokedAt = &now
	}
	r.mu.Unlock()

	return nil
}

// RevokeAllUserTokens invalidates all active sessions for a user (e.g. on password reset or token theft).
func (r *TokenRepository) RevokeAllUserTokens(userID string) error {
	now := time.Now()
	query := `UPDATE user_refresh_tokens SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL`
	_, _ = r.db.Exec(query, now, userID)

	r.mu.Lock()
	for _, rec := range r.memStore {
		if rec.UserID == userID {
			rec.RevokedAt = &now
		}
	}
	r.mu.Unlock()

	return nil
}
