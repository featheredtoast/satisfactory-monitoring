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
(((data -> 'location' ->> 'x')::NUMERIC + 374950) * 0.0000001075) as longitude,
(((data -> 'location' ->> 'y')::NUMERIC + 375000) * -0.0000001172) as latitude
FROM cache
cross join jsonb_path_query(frm_data, '$[*]') as data
where metric = 'factory'
LIMIT 500
```
