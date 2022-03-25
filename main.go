package main

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"tinygo.org/x/bluetooth"
)

type command struct {
	UUID   string
	Cmd    string
	Output string
}

func main() {
	config := getConfig()

	namespaceUUID := uuid.NewSHA1(uuid.NameSpaceDNS, []byte("discretetom.github.io"))
	serviceUUID := uuid.NewSHA1(namespaceUUID, []byte(config.Secret))
	wifiCharUUID := uuid.NewSHA1(serviceUUID, []byte("wifi"))
	cmdCharUUID := uuid.NewSHA1(serviceUUID, []byte("cmd"))

	serviceBleUUID, _ := bluetooth.ParseUUID(serviceUUID.String())
	wifiCharBleUUID, _ := bluetooth.ParseUUID(wifiCharUUID.String())
	cmdCharBleUUID, _ := bluetooth.ParseUUID(cmdCharUUID.String())

	adapter := bluetooth.DefaultAdapter
	currentCmd := &command{}

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
			{
				UUID:  wifiCharBleUUID,
				Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicReadPermission,
				ReadEvent: func(client bluetooth.Connection) ([]byte, error) {
					return json.Marshal(getWifiInfo())
				},
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					currentInfo := getWifiInfo()
					newInfo := wifi{}
					json.Unmarshal(value, &newInfo)

					if newInfo.SSID != currentInfo.SSID || newInfo.PSK != currentInfo.PSK {
						setNewWifi(newInfo.SSID, newInfo.PSK)
					}
					if !newInfo.Static && currentInfo.Static {
						cancelStaticIp(true)
					}
					if newInfo.Static && (newInfo.StaticIP != currentInfo.StaticIP || newInfo.Router != currentInfo.Router) {
						setNewStaticIP(newInfo.StaticIP, newInfo.Router)
					}
				},
			},
			{
				UUID:  cmdCharBleUUID,
				Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicReadPermission,
				ReadEvent: func(client bluetooth.Connection) ([]byte, error) {
					return json.Marshal(currentCmd)
				},
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					newCmd := &command{}
					json.Unmarshal(value, &newCmd)
					go func() {
						output, err := runCommand(newCmd.Cmd)
						if err != nil {
							output = "Error: " + err.Error()
						}
						currentCmd.Output = output
						currentCmd.UUID = newCmd.UUID
						currentCmd.Cmd = newCmd.Cmd
					}()
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
