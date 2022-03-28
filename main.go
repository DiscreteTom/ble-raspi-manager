package main

import (
	"fmt"
	"github/DiscreteTom/ble-raspi-manager/internal/characteristics/command"
	"github/DiscreteTom/ble-raspi-manager/internal/characteristics/wifi"
	"github/DiscreteTom/ble-raspi-manager/internal/config"
	"github/DiscreteTom/ble-raspi-manager/internal/shell"
	"net"
	"time"

	"github.com/google/uuid"
	"tinygo.org/x/bluetooth"
)

func main() {
	conf := config.GetConfig()

	namespaceUUID := uuid.NewSHA1(uuid.NameSpaceDNS, []byte("discretetom.github.io"))
	serviceUUID := uuid.NewSHA1(namespaceUUID, []byte(conf.Secret))
	serviceBleUUID, _ := bluetooth.ParseUUID(serviceUUID.String())
	localNameSuffix := uuid.NewMD5(uuid.Nil, []byte(mustGetMacAddr())).String()[:8]

	adapter := bluetooth.DefaultAdapter

	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())

	// Define the peripheral device info.
	adv := adapter.DefaultAdvertisement()
	must("config adv", adv.Configure(bluetooth.AdvertisementOptions{
		LocalName:    "BLE Raspi Manager - " + localNameSuffix,
		ServiceUUIDs: []bluetooth.UUID{serviceBleUUID},
	}))

	// Start advertising
	must("start adv", adv.Start())

	// Add Services
	must("add service", adapter.AddService(&bluetooth.Service{
		UUID: serviceBleUUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			wifi.NewCharacteristicConfig(serviceUUID),
			command.NewCharacteristicConfig(serviceUUID),
		},
	}))

	fmt.Println("ble-raspi-manager is running...")

	for { // run forever
		time.Sleep(time.Duration(conf.HealthCheckIntervalMs) * time.Millisecond)

		// health check
		if (shell.MustRunCommand("bluetoothctl show | grep 'Discoverable:' | cut -d' ' -f 2") == "no\n") || (shell.MustRunCommand("bluetoothctl show | grep ActiveInstances | cut -d'(' -f 2 | cut -d')' -f 1") == "0\n") {
			panic("health check failed.")
		}
	}
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}

func mustGetMacAddr() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	for _, interf := range interfaces {
		a := interf.HardwareAddr.String()
		if a != "" {
			return a
		}
	}
	panic("no MAC address found.")
}
