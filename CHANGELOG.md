# CHANGELOG

## v0.2.0

- Optimize communication protocol for long content.
  - So this version is not compatible with previous ones.
- Hide wifi password using `wpa_passphrase [ ssid ] [ passphrase ]`.
  - And wifi password can no longer be retrieved.
- Support multiple wifi configuration.

## v0.1.2

Add local name suffix based on device's MAC address.

## v0.1.1

Add health check to restart the service, since bluetooth will be non-discoverable and non-pairable for unknown reasons on startup.

## v0.1.0

Initial release.
