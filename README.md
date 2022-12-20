little getting started with grafana and prometheus.

Uses sample node exporter for prom.

localhost:3000 for grafana
localhost:9090 for prometheus.

from grafana, add new datasource prometheus:9090
uses docker networking here for DNS :)

# Ficsit Remote Monitoring Companion Bundle

Docker compose setup for Ficsit Remote Monitoring and alerting. Requires [FicsitRemoteMonitoring](https://ficsit.app/mod/FicsitRemoteMonitoring) plugin.
The http server booted up `/frmweb start` in game.

## Env Vars

Set environment variables for

FRM_HOST - Host where the Ficsit Remote Monitoring webserver is running. Generally this is your computer's IP address, EG: `192.168.1.20`. (Default: `fakeserver`)
FRM_PORT - Port where the Ficsit Remote Monitoring webserver is running. `8080` is FRM's default at the time of writing. (Default: `8080`)
DISCORD_WEBHOOK - Webhook for discord fuse and low battery notifications. Something like `https://discord.com/api/webhooks/12345/abcd12345`.

## Services and ports

- [frmcompanion](http://localhost:9000/metrics): A webapp that converts JSON data from FRM into Prometheus metrics at `localhost:9000/metrics`. There is also a realtime map app. `localhost:8000?frmhost=localhost&frmport=8080`.
- [prometheus](http://localhost:9090): Ingest metrics from the remote monitoring companion. Generates alert metrics for interesting anomalies.
- [alertmanager](http://localhost:9093): Forwards critical alerts to notification components.
- alertmanager-discord: Sends critical alerts to Discord.
- [grafana](http://localhost:3000): Time series graphing dashboard.
- fakeserver: Test server for fake metrics used for testing. Maps to host port 8081 to avoid port conflicts if FRM is running on localhost.
  - [getFactory](http://localhost:8081/getFactory)
  - [getPower](http://localhost:8081/getPower)
  - [getProdStats](http://localhost:8081/getProdStats)
  - [getTrains](http://localhost:8081/getTrains)
