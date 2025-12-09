-- Drop board_columns table and indexes
DROP INDEX IF EXISTS idx_board_columns_position;
DROP INDEX IF EXISTS idx_board_columns_board_id;
DROP TABLE IF EXISTS board_columns;
