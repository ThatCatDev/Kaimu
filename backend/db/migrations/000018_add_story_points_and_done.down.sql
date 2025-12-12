-- Remove is_done index and column
DROP INDEX IF EXISTS idx_board_columns_is_done;
ALTER TABLE board_columns DROP COLUMN IF EXISTS is_done;

-- Remove story_points from cards
ALTER TABLE cards DROP COLUMN IF EXISTS story_points;
