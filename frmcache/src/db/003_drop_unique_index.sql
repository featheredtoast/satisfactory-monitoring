-- Write your migrate up statements here

DROP INDEX if exists metric_idx;
CREATE INDEX cache_metric ON cache(metric);

---- create above / drop below ----

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.

CREATE UNIQUE INDEX metric_idx ON cache(metric);
DROP INDEX if exists cache_metric;
