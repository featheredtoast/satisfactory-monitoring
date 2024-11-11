-- Write your migrate up statements here

CREATE INDEX cache_with_history_metric ON cache_with_history(metric);
CREATE INDEX cache_with_history_time ON cache_with_history(time);

---- create above / drop below ----

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.

DROP INDEX if exists cache_with_history_metric;
DROP INDEX if exists cache_with_history_time;
