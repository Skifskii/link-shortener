-- Создание промежуточной таблицы пользователей и ссылок
CREATE TABLE users_links (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    link_id INTEGER NOT NULL REFERENCES links(id)
);
