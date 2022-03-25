package main

import (
	"fmt"
	"github/DiscreteTom/ble-raspi-manager/characteristics/command"
	"github/DiscreteTom/ble-raspi-manager/characteristics/wifi"
	"github/DiscreteTom/ble-raspi-manager/internal/config"
	"time"

	"github.com/google/uuid"
	"tinygo.org/x/bluetooth"
)

func main() {
	conf := config.GetConfig()

	namespaceUUID := uuid.NewSHA1(uuid.NameSpaceDNS, []byte("discretetom.github.io"))
	serviceUUID := uuid.NewSHA1(namespaceUUID, []byte(conf.Secret))
	serviceBleUUID, _ := bluetooth.ParseUUID(serviceUUID.String())

	adapter := bluetooth.DefaultAdapter

	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())

	// Define the peripheral device info.
	adv := adapter.DefaultAdvertisement()
	must("config adv", adv.Configure(bluetooth.AdvertisementOptions{
		LocalName:    "BLE Raspi Manager",
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
		time.Sleep(100 * time.Second)
	}
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
