-- Create audit action enum
CREATE TYPE audit_action AS ENUM (
    'created',
    'updated',
    'deleted',
    'card_moved',
    'card_assigned',
    'card_unassigned',
    'sprint_started',
    'sprint_completed',
    'card_added_to_sprint',
    'card_removed_from_sprint',
    'member_invited',
    'member_joined',
    'member_removed',
    'member_role_changed',
    'column_reordered',
    'column_visibility_toggled',
    'user_logged_in',
    'user_logged_out'
);

-- Create entity type enum
CREATE TYPE audit_entity_type AS ENUM (
    'user',
    'organization',
    'project',
    'board',
    'board_column',
    'card',
    'sprint',
    'tag',
    'role',
    'invitation'
);

-- Create audit_events table
CREATE TABLE audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action audit_action NOT NULL,
    entity_type audit_entity_type NOT NULL,
    entity_id UUID NOT NULL,
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    board_id UUID REFERENCES boards(id) ON DELETE SET NULL,
    state_before JSONB,
    state_after JSONB,
    metadata JSONB NOT NULL DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    trace_id TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for common query patterns

-- Activity feed: by organization, ordered by time
CREATE INDEX idx_audit_events_org_time ON audit_events(organization_id, occurred_at DESC);

-- Project activity feed
CREATE INDEX idx_audit_events_project_time ON audit_events(project_id, occurred_at DESC)
    WHERE project_id IS NOT NULL;

-- Board activity feed
CREATE INDEX idx_audit_events_board_time ON audit_events(board_id, occurred_at DESC)
    WHERE board_id IS NOT NULL;

-- Entity history: all events for a specific entity
CREATE INDEX idx_audit_events_entity ON audit_events(entity_type, entity_id, occurred_at DESC);

-- User activity: all events by a specific actor
CREATE INDEX idx_audit_events_actor ON audit_events(actor_id, occurred_at DESC)
    WHERE actor_id IS NOT NULL;

-- Metrics queries: card movements for burn charts
CREATE INDEX idx_audit_events_card_moves ON audit_events(board_id, occurred_at, action)
    WHERE entity_type = 'card' AND action IN ('card_moved', 'created', 'deleted', 'card_added_to_sprint', 'card_removed_from_sprint');

-- Time-based queries
CREATE INDEX idx_audit_events_time ON audit_events(occurred_at DESC);

-- Action type filtering
CREATE INDEX idx_audit_events_action ON audit_events(action, occurred_at DESC);
