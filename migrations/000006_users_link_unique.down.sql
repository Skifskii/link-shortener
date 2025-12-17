-- Применяет откат уникального ограничения на сочетание user_id и link_id в таблице users_links
ALTER TABLE users_links DROP CONSTRAINT IF EXISTS users_links_user_id_link_id_key;
