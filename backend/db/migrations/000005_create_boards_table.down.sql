-- Drop boards table and indexes
DROP INDEX IF EXISTS idx_unique_default_board_per_project;
DROP INDEX IF EXISTS idx_boards_project_id;
DROP TABLE IF EXISTS boards;
