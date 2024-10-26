-- Write your migrate up statements here

ALTER TABLE cache RENAME COLUMN save session_name;
ALTER TABLE cache_with_history RENAME COLUMN save session_name;

---- create above / drop below ----

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.

ALTER TABLE cache RENAME COLUMN session_name save;
ALTER TABLE cache_with_history RENAME COLUMN session_name save;
