-- Write your migrate up statements here
alter table cache
  rename column "frm_data" to "data";
alter table cache_with_history
  rename column "frm_data" to "data";

---- create above / drop below ----

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.

alter table cache
  rename column "data" to "frm"_data;
alter table cache_with_history
  rename column "data" to "frm_data";
