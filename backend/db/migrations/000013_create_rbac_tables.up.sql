-- RBAC (Role-Based Access Control) Tables

-- Create permissions table
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    resource_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_permissions_resource_type ON permissions(resource_type);
CREATE INDEX idx_permissions_code ON permissions(code);

-- Seed default permissions
INSERT INTO permissions (code, name, description, resource_type) VALUES
-- Organization permissions
('org:view', 'View Organization', 'Can view organization details', 'organization'),
('org:manage', 'Manage Organization', 'Can edit organization settings', 'organization'),
('org:delete', 'Delete Organization', 'Can delete the organization', 'organization'),
('org:invite', 'Invite Members', 'Can invite new members to organization', 'organization'),
('org:remove_members', 'Remove Members', 'Can remove members from organization', 'organization'),
('org:manage_roles', 'Manage Roles', 'Can create and edit custom roles', 'organization'),

-- Project permissions
('project:view', 'View Project', 'Can view project details', 'project'),
('project:create', 'Create Project', 'Can create new projects', 'project'),
('project:manage', 'Manage Project', 'Can edit project settings', 'project'),
('project:delete', 'Delete Project', 'Can delete projects', 'project'),
('project:manage_members', 'Manage Project Members', 'Can add/remove project members', 'project'),

-- Board permissions
('board:view', 'View Board', 'Can view board and columns', 'board'),
('board:create', 'Create Board', 'Can create new boards', 'board'),
('board:manage', 'Manage Board', 'Can edit board settings and columns', 'board'),
('board:delete', 'Delete Board', 'Can delete boards', 'board'),

-- Card permissions
('card:view', 'View Cards', 'Can view cards on boards', 'card'),
('card:create', 'Create Cards', 'Can create new cards', 'card'),
('card:edit', 'Edit Cards', 'Can edit card details', 'card'),
('card:move', 'Move Cards', 'Can move cards between columns', 'card'),
('card:delete', 'Delete Cards', 'Can delete cards', 'card'),
('card:assign', 'Assign Cards', 'Can assign cards to users', 'card');

-- Create roles table
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    scope VARCHAR(50) NOT NULL DEFAULT 'organization',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_role_name_per_org UNIQUE (organization_id, name)
);

CREATE INDEX idx_roles_org_id ON roles(organization_id);
CREATE INDEX idx_roles_is_system ON roles(is_system);

-- Insert system roles with fixed UUIDs
INSERT INTO roles (id, organization_id, name, description, is_system, scope) VALUES
('00000000-0000-0000-0000-000000000001', NULL, 'Owner', 'Full access to everything. Cannot be removed or demoted.', TRUE, 'organization'),
('00000000-0000-0000-0000-000000000002', NULL, 'Admin', 'Administrative access to manage organization and projects.', TRUE, 'organization'),
('00000000-0000-0000-0000-000000000003', NULL, 'Member', 'Standard member with ability to contribute to projects.', TRUE, 'organization'),
('00000000-0000-0000-0000-000000000004', NULL, 'Viewer', 'Read-only access to view content.', TRUE, 'organization');

-- Create role_permissions junction table
CREATE TABLE role_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_role_permission UNIQUE (role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);

-- Owner gets all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000000-0000-0000-0000-000000000001', id FROM permissions;

-- Admin gets all permissions except org:delete and org:manage_roles
INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000000-0000-0000-0000-000000000002', id FROM permissions
WHERE code NOT IN ('org:delete', 'org:manage_roles');

-- Member gets view + create + edit permissions (not delete/manage)
INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000000-0000-0000-0000-000000000003', id FROM permissions
WHERE code IN (
    'org:view',
    'project:view', 'project:create',
    'board:view', 'board:create',
    'card:view', 'card:create', 'card:edit', 'card:move', 'card:assign'
);

-- Viewer gets read-only permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000000-0000-0000-0000-000000000004', id FROM permissions
WHERE code IN ('org:view', 'project:view', 'board:view', 'card:view');

-- Add role_id column to organization_members
ALTER TABLE organization_members ADD COLUMN role_id UUID REFERENCES roles(id) ON DELETE SET NULL;
CREATE INDEX idx_org_members_role_id ON organization_members(role_id);

-- Migrate existing role strings to role_id
UPDATE organization_members SET role_id = '00000000-0000-0000-0000-000000000001' WHERE role = 'owner';
UPDATE organization_members SET role_id = '00000000-0000-0000-0000-000000000003' WHERE role = 'member';

-- Create project_members table for project-level role assignments
CREATE TABLE project_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_project_member UNIQUE (project_id, user_id)
);

CREATE INDEX idx_project_members_project_id ON project_members(project_id);
CREATE INDEX idx_project_members_user_id ON project_members(user_id);
CREATE INDEX idx_project_members_role_id ON project_members(role_id);

-- Create invitations table
CREATE TABLE invitations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
    invited_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    accepted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_pending_invitation UNIQUE (organization_id, email)
);

CREATE INDEX idx_invitations_org_id ON invitations(organization_id);
CREATE INDEX idx_invitations_token ON invitations(token);
CREATE INDEX idx_invitations_email ON invitations(email);
