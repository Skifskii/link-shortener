-- Создание таблицы ссылок
CREATE TABLE links (
    id SERIAL PRIMARY KEY,
    short TEXT NOT NULL UNIQUE,
    original TEXT NOT NULL
);

-- Базовый индекс для поиска по названию
CREATE INDEX idx_links_short ON links(short);
