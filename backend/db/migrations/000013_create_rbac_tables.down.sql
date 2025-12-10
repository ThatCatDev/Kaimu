-- Rollback RBAC tables

-- Drop invitations table
DROP TABLE IF EXISTS invitations;

-- Drop project_members table
DROP TABLE IF EXISTS project_members;

-- Remove role_id from organization_members
DROP INDEX IF EXISTS idx_org_members_role_id;
ALTER TABLE organization_members DROP COLUMN IF EXISTS role_id;

-- Drop role_permissions junction table
DROP TABLE IF EXISTS role_permissions;

-- Drop roles table
DROP TABLE IF EXISTS roles;

-- Drop permissions table
DROP TABLE IF EXISTS permissions;
