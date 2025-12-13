-- Create refresh tokens table for JWT refresh token rotation
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE,
    replaced_by UUID REFERENCES refresh_tokens(id),
    user_agent TEXT,
    ip_address VARCHAR(45)
);

-- Index for token lookup (we store hash, lookup by hash)
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);

-- Index for user's tokens (for listing/revoking all user sessions)
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- Index for cleanup of expired tokens
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- Index for finding active (non-revoked) tokens
CREATE INDEX idx_refresh_tokens_revoked_at ON refresh_tokens(revoked_at) WHERE revoked_at IS NULL;
