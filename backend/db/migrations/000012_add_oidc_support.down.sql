-- Rollback OIDC support

-- Drop OIDC identities table
DROP TABLE IF EXISTS oidc_identities;

-- Drop users indexes
DROP INDEX IF EXISTS idx_users_email;

-- Remove OIDC fields from users
ALTER TABLE users
    DROP COLUMN IF EXISTS email,
    DROP COLUMN IF EXISTS display_name,
    DROP COLUMN IF EXISTS avatar_url;

-- Make password_hash required again (will fail if there are OIDC-only users!)
-- Note: You may need to delete OIDC-only users before running this migration down
ALTER TABLE users
    ALTER COLUMN password_hash SET NOT NULL;
