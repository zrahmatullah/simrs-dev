-- ==============================================================================
-- MIGRATION: OWASP SECURITY HARDENING & TOKEN MANAGEMENT
-- Table: user_refresh_tokens
-- Purpose: Secure persistent storage of hashed refresh tokens for session rotation
-- ==============================================================================

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

-- Performance & security indexes
CREATE INDEX IF NOT EXISTS idx_refresh_token_hash ON user_refresh_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_user_id ON user_refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_expires_at ON user_refresh_tokens(expires_at);

-- Comments for documentation & audit
COMMENT ON TABLE user_refresh_tokens IS 'Tabel pencatatan refresh token sesi pengguna dengan hash SHA-256 sesuai standar OWASP ASVS V3.';
COMMENT ON COLUMN user_refresh_tokens.token_hash IS 'Hash SHA-256 dari token acak kriptografis 256-bit (token mentah tidak pernah disimpan di database).';
COMMENT ON COLUMN user_refresh_tokens.revoked_at IS 'Waktu pencabutan token saat rotasi atau logout. Jika token revoked digunakan kembali, sistem mendeteksi token theft.';
