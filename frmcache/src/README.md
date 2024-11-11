# FRM Cache

A small go app that grabs raw json metrics and stats from Ficsit Remote Monitor and stuffs them to a postgres DB.

This allows for raw stats to be exposed directly to grafana's postgres datasource.

This is not a timeseries so there is no metrics over time available, only raw data. Timeseries and alerts are still better handled by prometheus.

Queries are only made once/minute, as metrics we're using here do not get stale easily.

Metrics here are best used for:
* location based tracking of buildings
* summation - prometheus fields are more difficult to aggregate together.

This also allows queries that convert from in-game coordinate system to somewhere on the world map, so that we can display it on grafana without a custom map plugin.

EG: All factory locations, note conversions to longitude and latitude:

```
SELECT
data ->> 'building' as building,
(data -> 'location' ->> 'x')::NUMERIC/100 as x,
(data -> 'location' ->> 'y')::NUMERIC/100 as y,
(data -> 'location' ->> 'z')::NUMERIC/100 as z,
(((data -> 'location' ->> 'x')::NUMERIC + 375000) * 0.0000001015) as longitude,
(((data -> 'location' ->> 'y')::NUMERIC + 375000) * -0.0000001172) as latitude
FROM cache
cross join jsonb_path_query(frm_data, '$[*]') as data
where metric = 'factory'
LIMIT 500
```

## Env vars

Configuration is done strictly through ENV vars.

`FRM_HOST`: The host to the Ficsit Remote Monitoring server. EG: `172.17.0.1`.
`FRM_PORT`: The port of the Ficist Remote Monitoring server. EG: `8080`.
`FRM_HOSTS`: A comma separated list of Ficsit Remote Monitoring servers. If protocol is unspecified, it defaults to http. EG: `http://myserver1.frm.example:8080,myserver2.frm.example:8080,https://myserver3.frm.example:8081`
`PG_HOST`: Postgres host. Defaults to `postgres`
`PG_PORT`: Postgres port. Defaults to `5432`
`PG_PASSWORD`: Postgres password. Defaults to `secretpassword`
`PG_USER`: Postgres user to connect as. Defaults to `postgres`
`PG_DB`: Postgres database to connect to. Defaults to `postgres`
`MIGRATION_DIR`: List of [tern](https://github.com/jackc/tern) migrations to execute. Defaults to `/var/lib/cache`
