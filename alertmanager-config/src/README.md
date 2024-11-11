# Alertmanager-config

Small app to generate valid alertmanager config

Configured entirely through env vars for simplicity.

## Env Vars

`FRM_HOST`: The host to the Ficsit Remote Monitoring server. EG: `172.17.0.1`.
`FRM_PORT`: The port of the Ficist Remote Monitoring server. EG: `8080`.
`FRM_HOSTS`: A comma separated list of Ficsit Remote Monitoring servers. If protocol is unspecified, it defaults to http. EG: `http://myserver1.frm.example:8080,myserver2.frm.example:8080,https://myserver3.frm.example:8081`
`DISCORD_WEBHOOK`: A single [discord webhook](https://discord.com/developers/docs/resources/webhook) to send for all configured FRM hosts.
`DISCORD_WEBHOOKS`: A comma-separated list of [discord webhooks](https://discord.com/developers/docs/resources/webhook) to pair up with the FRM host.
`OUTPUT_PATH`: The output path of the resulting alertmanager config.

Hosts will be paired with webhooks in order. EG for a env config like:
```
FRM_HOSTS=host1,host2,host3
DISCORD_WEBHOOKS=webhook1,webhook2,webhook3
```
host1 will fire on webhook1, host2 will fire on webhook2, and host3 will fire on webhook3.

Extra hosts will not have webhooks associated with them, and extra webhooks will not be associated with a host.
