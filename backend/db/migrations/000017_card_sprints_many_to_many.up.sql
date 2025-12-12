-- Create card_sprints join table for many-to-many relationship
-- This allows cards to be in multiple sprints (for carryover scenarios)

CREATE TABLE card_sprints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_id UUID NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    sprint_id UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(card_id, sprint_id)
);

-- Create indexes for efficient lookups
CREATE INDEX idx_card_sprints_card_id ON card_sprints(card_id);
CREATE INDEX idx_card_sprints_sprint_id ON card_sprints(sprint_id);

-- Migrate existing sprint assignments from cards table
INSERT INTO card_sprints (card_id, sprint_id, added_at)
SELECT id, sprint_id, updated_at
FROM cards
WHERE sprint_id IS NOT NULL;

-- Drop the old sprint_id column from cards
ALTER TABLE cards DROP COLUMN sprint_id;
