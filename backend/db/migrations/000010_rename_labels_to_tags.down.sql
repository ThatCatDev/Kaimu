-- Revert: Rename tags table back to labels
ALTER TABLE tags RENAME TO labels;

-- Revert: Rename card_tags junction table back to card_labels
ALTER TABLE card_tags RENAME TO card_labels;

-- Revert: Rename the tag_id column back to label_id
ALTER TABLE card_labels RENAME COLUMN tag_id TO label_id;

-- Revert: Rename indexes
ALTER INDEX idx_tags_project_id RENAME TO idx_labels_project_id;
ALTER INDEX idx_card_tags_tag_id RENAME TO idx_card_labels_label_id;

-- Revert: Rename the unique constraint
ALTER TABLE labels RENAME CONSTRAINT tags_project_id_name_key TO labels_project_id_name_key;

-- Revert: Rename foreign key constraints
ALTER TABLE card_labels RENAME CONSTRAINT card_tags_card_id_fkey TO card_labels_card_id_fkey;
ALTER TABLE card_labels RENAME CONSTRAINT card_tags_tag_id_fkey TO card_labels_label_id_fkey;

-- Revert: Rename primary key constraint
ALTER TABLE card_labels RENAME CONSTRAINT card_tags_pkey TO card_labels_pkey;
