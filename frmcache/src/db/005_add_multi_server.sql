-- Write your migrate up statements here

ALTER TABLE cache ADD url varchar(200);
ALTER TABLE cache ADD save varchar(200);
ALTER TABLE cache_with_history ADD url varchar(200);
ALTER TABLE cache_with_history ADD save varchar(200);
CREATE INDEX cache_url_save ON cache(url, save);
CREATE INDEX cache_with_history_url_save ON cache_with_history(url, save);

---- create above / drop below ----

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.

DROP INDEX if exists cache_url_save;
DROP INDEX if exists cache_with_history_url_save;
ALTER TABLE cache DROP COLUMN url;
ALTER TABLE cache DROP COLUMN save;
ALTER TABLE cache_with_history DROP COLUMN url;
ALTER TABLE cache_with_history DROP COLUMN save;
