-- Drop cards table and indexes
DROP INDEX IF EXISTS idx_cards_assignee_id;
DROP INDEX IF EXISTS idx_cards_position;
DROP INDEX IF EXISTS idx_cards_board_id;
DROP INDEX IF EXISTS idx_cards_column_id;
DROP TABLE IF EXISTS cards;
DROP TYPE IF EXISTS card_priority;
