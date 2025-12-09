-- Create card_labels junction table
CREATE TABLE card_labels (
    card_id UUID NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    label_id UUID NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (card_id, label_id)
);

-- Index for finding all cards with a specific label
CREATE INDEX idx_card_labels_label_id ON card_labels(label_id);
