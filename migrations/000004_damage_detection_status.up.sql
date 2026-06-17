ALTER TABLE damage_detection_summaries
  ADD COLUMN status TEXT NOT NULL DEFAULT 'processing',
  ALTER COLUMN condition DROP NOT NULL,
  ALTER COLUMN condition_note DROP NOT NULL;
