-- Rename labels table to tags
ALTER TABLE labels RENAME TO tags;

-- Rename card_labels junction table to card_tags
ALTER TABLE card_labels RENAME TO card_tags;

-- Rename the label_id column to tag_id in the junction table
ALTER TABLE card_tags RENAME COLUMN label_id TO tag_id;

-- Rename indexes
ALTER INDEX idx_labels_project_id RENAME TO idx_tags_project_id;
ALTER INDEX idx_card_labels_label_id RENAME TO idx_card_tags_tag_id;

-- Rename the unique constraint on tags table (project_id, name)
ALTER TABLE tags RENAME CONSTRAINT labels_project_id_name_key TO tags_project_id_name_key;

-- Rename foreign key constraints
ALTER TABLE card_tags RENAME CONSTRAINT card_labels_card_id_fkey TO card_tags_card_id_fkey;
ALTER TABLE card_tags RENAME CONSTRAINT card_labels_label_id_fkey TO card_tags_tag_id_fkey;

-- Rename primary key constraint
ALTER TABLE card_tags RENAME CONSTRAINT card_labels_pkey TO card_tags_pkey;
