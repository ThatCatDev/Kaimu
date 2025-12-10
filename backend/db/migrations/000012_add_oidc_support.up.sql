-- Add OIDC support to Pulse

-- Add OIDC-related fields to users table
ALTER TABLE users
    ADD COLUMN email VARCHAR(255),
    ADD COLUMN display_name VARCHAR(255),
    ADD COLUMN avatar_url TEXT;

-- Make password_hash nullable (OIDC users won't have passwords)
ALTER TABLE users
    ALTER COLUMN password_hash DROP NOT NULL;

-- Create index on email for lookups
CREATE INDEX idx_users_email ON users(email);

-- Create OIDC identities table (allows multiple OIDC providers per user)
CREATE TABLE oidc_identities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    issuer VARCHAR(512) NOT NULL,
    subject VARCHAR(512) NOT NULL,
    email VARCHAR(255),
    email_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Unique constraint: one identity per issuer+subject combination
    CONSTRAINT unique_oidc_identity UNIQUE (issuer, subject)
);

-- Index for fast lookup by user
CREATE INDEX idx_oidc_identities_user_id ON oidc_identities(user_id);

-- Index for OIDC login lookup (issuer + subject)
CREATE INDEX idx_oidc_identities_issuer_subject ON oidc_identities(issuer, subject);
