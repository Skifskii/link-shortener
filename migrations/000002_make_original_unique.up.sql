-- Добавление столбцу original ограничения уникальности
ALTER TABLE links
ADD CONSTRAINT unique_original UNIQUE(original);
