-- Restore sprint_id column on cards
ALTER TABLE cards ADD COLUMN sprint_id UUID REFERENCES sprints(id) ON DELETE SET NULL;

-- Migrate data back (take the most recent sprint assignment for each card)
UPDATE cards c
SET sprint_id = (
    SELECT cs.sprint_id
    FROM card_sprints cs
    WHERE cs.card_id = c.id
    ORDER BY cs.added_at DESC
    LIMIT 1
);

-- Drop the join table
DROP TABLE card_sprints;
