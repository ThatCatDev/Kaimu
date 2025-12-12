-- Remove performance indexes added in migration 000021

DROP INDEX IF EXISTS idx_sprints_board_status;
DROP INDEX IF EXISTS idx_sprints_board_status_enddate;
DROP INDEX IF EXISTS idx_cards_board_position;
DROP INDEX IF EXISTS idx_card_sprints_card_added;
DROP INDEX IF EXISTS idx_audit_events_metadata_gin;
DROP INDEX IF EXISTS idx_users_email_verified;
DROP INDEX IF EXISTS idx_invitations_org_expires;
