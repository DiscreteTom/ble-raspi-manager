package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"

	"tinygo.org/x/bluetooth"
)

func main() {
	serviceUUID, _ := bluetooth.ParseUUID("f846752b-af47-43ed-bdf0-fba82da6fd58")
	setWifiCharUUID, _ := bluetooth.ParseUUID("e94f5099-db86-4b29-a4ce-08033fda1a7d")
	refreshCharUUID, _ := bluetooth.ParseUUID("565633b0-5a8e-42e6-9cd3-e058efb2b0c4")
	updateTimeCharUUID, _ := bluetooth.ParseUUID("59b123de-c658-4526-b280-0e58d11333b3")

	adapter := bluetooth.DefaultAdapter
	var lastUpdateTime = [8]byte{}
	binary.LittleEndian.PutUint64(lastUpdateTime[:], uint64(time.Now().Unix()))
	// var currentWifiSsid = [32]byte{}
	// var currentIP = [4]byte{}

	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())

	// Define the peripheral device info.
	adv := adapter.DefaultAdvertisement()
	must("config adv", adv.Configure(bluetooth.AdvertisementOptions{
		LocalName:    "Raspi Wifi Manager",
		ServiceUUIDs: []bluetooth.UUID{serviceUUID},
	}))

	// Start advertising
	must("start adv", adv.Start())

	var wifiChar bluetooth.Characteristic
	must("add service", adapter.AddService(&bluetooth.Service{
		UUID: serviceUUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				Handle: &wifiChar,
				UUID:   setWifiCharUUID,
				Flags:  bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					fmt.Print(offset)
					fmt.Print(value)
				},
			},
			{
				Handle: &wifiChar,
				UUID:   refreshCharUUID,
				Flags:  bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					fmt.Print(offset)
					fmt.Print(value)
				},
			},
			{
				Handle: &wifiChar,
				UUID:   updateTimeCharUUID,
				Flags:  bluetooth.CharacteristicReadPermission,
				Value:  lastUpdateTime[:],
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

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
