# Ficsit Remote Monitoring Companion Bundle

A docker-compose for getting Satisfactory dashboards and alerting. Configured for:

### General production and power stats
![dashboard](resources/satisfactory-dash.png)

Power grid overview and Production vs Consumption levels.
![alert-dashboard](resources/satisfactory-alerts-dash.png)

Power and production alerts are also displayed.

### Discord alerts
![discord alert](resources/satisfactory-alert.png)

Discord alerting for tripped grids.

### Vehicle dashboard overview
![vehicle-dash](resources/vehicle-dash1.png)

Get an overview of your vehicle, train, and drone infrastructure.

View vehicles real-time round trip times and fuel.

View Trains round trip times and travel times between stations.

![vehicle-dash2](resources/vehicle-dash2.png)

View drone battery usage and round trip times.

Docker compose setup for Ficsit Remote Monitoring and alerting. Requires [FicsitRemoteMonitoring](https://ficsit.app/mod/FicsitRemoteMonitoring) plugin.
With FicsitRemoteMonitoring plugin, make sure you boot the http server `/frmweb start` in game.

## Env Vars

- `FRM_HOST` - Host where the Ficsit Remote Monitoring webserver is running. Generally this is your computer's IP address, EG: `192.168.1.20`. (Default: `fakeserver`)
- `FRM_PORT` - Port where the Ficsit Remote Monitoring webserver is running. `8080` is FRM's default at the time of writing. (Default: `8080`)
- `DISCORD_WEBHOOK` - Webhook for discord fuse and low battery notifications. Something like `https://discord.com/api/webhooks/12345/abcd12345`.

## Services and ports

- [frmcompanion](http://localhost:9000/metrics): A webapp that converts JSON data from FRM into Prometheus metrics at `localhost:9000/metrics`. There is also a realtime map app. `localhost:8000?frmhost=localhost&frmport=8080`.
- [prometheus](http://localhost:9090): Ingest metrics from the remote monitoring companion. Generates alert metrics for interesting anomalies.
- [alertmanager](http://localhost:9093): Forwards critical alerts to notification components.
- alertmanager-discord: Sends critical alerts to Discord.
- [grafana](http://localhost:3000): Time series graphing dashboard. Default username/password is `admin/admin`
- fakeserver: Test server for fake metrics used for testing. Maps to host port 8081 to avoid port conflicts if FRM is running on localhost.
  - [getFactory](http://localhost:8081/getFactory)
  - [getPower](http://localhost:8081/getPower)
  - [getProdStats](http://localhost:8081/getProdStats)
  - [getTrains](http://localhost:8081/getTrains)
