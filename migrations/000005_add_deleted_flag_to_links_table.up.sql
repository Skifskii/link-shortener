-- Добавление флага удаления в таблицу ссылок
ALTER TABLE links
ADD COLUMN is_deleted BOOLEAN DEFAULT FALSE;