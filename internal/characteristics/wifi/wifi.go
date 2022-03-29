package wifi

import (
	"encoding/json"
	"fmt"
	"github/DiscreteTom/ble-raspi-manager/internal/shell"

	"github.com/google/uuid"
	"tinygo.org/x/bluetooth"
)

type wifiInfo struct {
	SSID      string // wifi name
	PSK       string // wifi password
	CurrentIP string // current ip address
	Router    string
	Static    bool   // whether ip is static
	StaticIP  string // configured static ip
}

func getWifiInfo() wifiInfo {
	staticIP := getStaticIP()
	return wifiInfo{
		SSID:      getSSID(),
		PSK:       "password can NOT be retrieved",
		CurrentIP: getCurrentIP(),
		Static:    len(staticIP) != 0,
		StaticIP:  staticIP,
		Router:    getRouter(),
	}
}

func getCommandOutput(command string) string {
	out, err := shell.RunCommand(command)
	if len(out) == 0 || err != nil {
		return ""
	}
	return out[:len(out)-1] // remove suffix `\n`
}

func getCurrentIP() string {
	return getCommandOutput("ifconfig wlan0 | sed -En 's/127.0.0.1//;s/.*inet (addr:)?(([0-9]*\\.){3}[0-9]*).*/\\2/p'")
}

func getSSID() string {
	return getCommandOutput("cat /etc/wpa_supplicant/wpa_supplicant.conf | grep -Eo 'ssid=\"[^\"]*\"' | cut -d'\"' -f 2")
}

func getStaticIP() string {
	return getCommandOutput("cat /etc/dhcpcd.conf | grep -Eo '^static ip_address=.*' | cut -d'=' -f 2")
}

func getRouter() string {
	return getCommandOutput("netstat -nr | awk '$1 == \"0.0.0.0\"{print$2}'")
}

func setNewWifi(ssid, psk string) {
	if len(getSSID()) != 0 {
		// remove original settings
		shell.MustRunCommand("cp /etc/wpa_supplicant/wpa_supplicant.conf /etc/wpa_supplicant/wpa_supplicant.conf.backup")
		shell.MustRunCommand("cat /etc/wpa_supplicant/wpa_supplicant.conf.backup | grep -v 'network=' | grep -v 'ssid=' | grep -v 'psk=' | grep -v '}' > /etc/wpa_supplicant/wpa_supplicant.conf")
	}
	// write new settings
	shell.MustRunCommand(fmt.Sprintf("wpa_passphrase '%s' '%s' | grep -v '#psk' >> /etc/wpa_supplicant/wpa_supplicant.conf", ssid, psk))
	shell.MustRunCommand("wpa_cli -i wlan0 reconfigure") // restart interface to apply
}

func setNewStaticIP(ip, routers string) {
	cancelStaticIp(false)
	shell.MustRunCommand("echo 'interface wlan0' >> /etc/dhcpcd.conf")
	shell.MustRunCommand(fmt.Sprintf("echo 'static ip_address=%s' >> /etc/dhcpcd.conf", ip))
	shell.MustRunCommand(fmt.Sprintf("echo 'static routers=%s' >> /etc/dhcpcd.conf", routers))

	fmt.Printf("Reboot by set new static ip/routers: %s %s.", ip, routers)
	shell.MustRunCommand("reboot now")
}

func cancelStaticIp(reboot bool) {
	shell.MustRunCommand("cp /etc/dhcpcd.conf /etc/dhcpcd.conf.backup")
	shell.MustRunCommand("cat /etc/dhcpcd.conf.backup | grep -Ev '^interface wlan0' | grep -Ev '^static ip_address=' | grep -Ev '^static routers=' > /etc/dhcpcd.conf")
	if reboot {
		fmt.Println("Reboot by cancel static ip.")
		shell.MustRunCommand("reboot now")
	}
}

func NewCharacteristicConfig(serviceUUID uuid.UUID) bluetooth.CharacteristicConfig {
	wifiCharUUID := uuid.NewSHA1(serviceUUID, []byte("wifi"))
	wifiCharBleUUID, _ := bluetooth.ParseUUID(wifiCharUUID.String())

	return bluetooth.CharacteristicConfig{
		UUID:  wifiCharBleUUID,
		Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicReadPermission,
		ReadEvent: func(client bluetooth.Connection) ([]byte, error) {
			return json.Marshal(getWifiInfo())
		},
		WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
			currentInfo := getWifiInfo()
			newInfo := wifiInfo{}
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
	}
}
