-- Применяет уникальное ограничение на сочетание user_id и link_id в таблице users_links
ALTER TABLE users_links ADD CONSTRAINT users_links_user_id_link_id_key UNIQUE (user_id, link_id);
