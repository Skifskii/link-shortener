-- Откат ограничения уникальности со столбца original
ALTER TABLE links
DROP CONSTRAINT IF EXISTS unique_original;
