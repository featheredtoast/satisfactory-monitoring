-- Write your migrate up statements here

create table cache_with_history(
  id serial primary key,
  metric text NOT NULL,
  frm_data jsonb,
  time timestamp
);

---- create above / drop below ----

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.

drop table cache_with_history;
