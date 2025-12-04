-- Удаление флага удаления из таблицы links
ALTER TABLE links
DROP COLUMN is_deleted;