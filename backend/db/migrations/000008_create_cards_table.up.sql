-- Create card priority enum
CREATE TYPE card_priority AS ENUM ('none', 'low', 'medium', 'high', 'urgent');

-- Create cards table
CREATE TABLE cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    column_id UUID NOT NULL REFERENCES board_columns(id) ON DELETE CASCADE,
    board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    position FLOAT NOT NULL DEFAULT 0,
    priority card_priority NOT NULL DEFAULT 'none',
    assignee_id UUID REFERENCES users(id) ON DELETE SET NULL,
    due_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL
);

-- Index for fast column lookups
CREATE INDEX idx_cards_column_id ON cards(column_id);

-- Index for fast board lookups
CREATE INDEX idx_cards_board_id ON cards(board_id);

-- Index for ordering cards by position within a column
CREATE INDEX idx_cards_position ON cards(column_id, position);

-- Index for finding cards assigned to a user
CREATE INDEX idx_cards_assignee_id ON cards(assignee_id);
