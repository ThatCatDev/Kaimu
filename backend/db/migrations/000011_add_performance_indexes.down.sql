-- Rollback performance indexes

DROP INDEX IF EXISTS idx_cards_assignee_due_date;
DROP INDEX IF EXISTS idx_board_columns_visible;
DROP INDEX IF EXISTS idx_card_tags_card_id;
DROP INDEX IF EXISTS idx_cards_created_at;
DROP INDEX IF EXISTS idx_boards_created_at;
DROP INDEX IF EXISTS idx_tags_name;
