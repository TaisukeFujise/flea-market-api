UPDATE damage_detection_summaries SET condition = 'good' WHERE condition IS NULL;
UPDATE damage_detection_summaries SET condition_note = '' WHERE condition_note IS NULL;

ALTER TABLE damage_detection_summaries
  DROP COLUMN status,
  ALTER COLUMN condition SET NOT NULL,
  ALTER COLUMN condition_note SET NOT NULL;
