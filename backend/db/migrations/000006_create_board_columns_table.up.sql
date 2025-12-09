-- Create board_columns table
CREATE TABLE board_columns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    is_backlog BOOLEAN NOT NULL DEFAULT FALSE,
    is_hidden BOOLEAN NOT NULL DEFAULT FALSE,
    color VARCHAR(7) DEFAULT '#6B7280',
    wip_limit INTEGER,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for fast board lookups
CREATE INDEX idx_board_columns_board_id ON board_columns(board_id);

-- Index for ordering columns by position
CREATE INDEX idx_board_columns_position ON board_columns(board_id, position);
