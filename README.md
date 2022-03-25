# ble-raspi-manager

Manage Raspberry Pi using Bluetooth Low Energy.

This repo contains the BLE server, which should be run on raspi. The BLE client is a browser app, you can access it [here](https://discretetom.github.io/Omnitrix/ble-raspi-manager/) from your PC or your phone.

## Build

```bash
go build .
```

## Installation

**Run as root**:

```bash
# create a folder
mkdir /root/brm
cd /root/brm

# download a release
wget -q https://github.com/DiscreteTom/raspi-ble-wifi-manager/releases/latest/download/brm-arm.zip
# or, for arm 64:
# wget https://github.com/DiscreteTom/raspi-ble-wifi-manager/releases/latest/download/brm-arm64.zip

# unzip and remove zip file
unzip brm* && rm *.zip

# install the service
cp brm.service /etc/systemd/system/

# reload systemd
systemctl daemon-reload

# optional: modify config
# vim ~/brm/config.json

# start the service
systemctl start brm

# start the service on system startup
systemctl enable brm
```
