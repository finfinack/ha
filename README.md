# HA State

Small service which fetches the state of a Home Assistant server
and filters specific entities to later expose them in a new JSON format
via web API in order to use it for smart displays or so.

## Run (Dev Mode)

```
go run ha.go -config ./config.json
```

### Configuration

```
{
  "ha_auth_token": "<long lived access token from HA>",
  "ha_status_url": "<HA web endpoint>/api/states",
  "include_entities": [
    "(?i).*temperature.*"
  ],
  "exclude_entities": [
    "(?i).*mystrom.*",
    "(?i).*shelly.*",
  ]
}
```