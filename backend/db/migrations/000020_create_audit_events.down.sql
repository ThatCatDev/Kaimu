-- Drop indexes
DROP INDEX IF EXISTS idx_audit_events_org_time;
DROP INDEX IF EXISTS idx_audit_events_project_time;
DROP INDEX IF EXISTS idx_audit_events_board_time;
DROP INDEX IF EXISTS idx_audit_events_entity;
DROP INDEX IF EXISTS idx_audit_events_actor;
DROP INDEX IF EXISTS idx_audit_events_card_moves;
DROP INDEX IF EXISTS idx_audit_events_time;
DROP INDEX IF EXISTS idx_audit_events_action;

-- Drop table
DROP TABLE IF EXISTS audit_events;

-- Drop types
DROP TYPE IF EXISTS audit_entity_type;
DROP TYPE IF EXISTS audit_action;
