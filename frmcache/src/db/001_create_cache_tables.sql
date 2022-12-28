-- Write your migrate up statements here

create table cache(
  id serial primary key,
  metric text NOT NULL,
  frm_data jsonb
  );

CREATE UNIQUE INDEX metric_idx ON cache(metric);

---- create above / drop below ----

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.

  drop table cache;
