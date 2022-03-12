package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"tinygo.org/x/bluetooth"
)

type info struct {
	SSID      string // wifi name
	PSK       string // wiki password
	CurrentIP string // current ip address
	Static    bool   // whether ip is static
	StaticIP  string // configured static ip
}

func main() {
	config := getConfig()

	namespaceUUID := uuid.NewSHA1(uuid.NameSpaceDNS, []byte("discretetom.github.io"))
	serviceUUID := uuid.NewSHA1(namespaceUUID, []byte(config.Secret))
	charUUID := uuid.NewSHA1(serviceUUID, []byte("info"))

	serviceBleUUID, _ := bluetooth.ParseUUID(serviceUUID.String())
	charBleUUID, _ := bluetooth.ParseUUID(charUUID.String())

	adapter := bluetooth.DefaultAdapter

	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())

	// Define the peripheral device info.
	adv := adapter.DefaultAdvertisement()
	must("config adv", adv.Configure(bluetooth.AdvertisementOptions{
		LocalName:    "Raspi Wifi Manager",
		ServiceUUIDs: []bluetooth.UUID{serviceBleUUID},
	}))

	// Start advertising
	must("start adv", adv.Start())

	// Add Services
	var wifiChar bluetooth.Characteristic
	must("add service", adapter.AddService(&bluetooth.Service{
		UUID: serviceBleUUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				Handle: &wifiChar,
				UUID:   charBleUUID,
				Flags:  bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicReadPermission,
				ReadEvent: func(client bluetooth.Connection) ([]byte, error) {
					static, staticIP := getStaticIP()
					result := info{
						SSID:      getSSID(),
						PSK:       getPSK(),
						CurrentIP: getCurrentIP(),
						Static:    static,
						StaticIP:  staticIP,
					}
					return json.Marshal(result)
				},
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					fmt.Print(value)
				},
			},
		},
	}))

	for { // run forever
		time.Sleep(100 * time.Second)
	}
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
