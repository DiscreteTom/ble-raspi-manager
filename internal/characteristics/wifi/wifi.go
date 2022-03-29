package wifi

import (
	"encoding/json"
	"fmt"
	"github/DiscreteTom/ble-raspi-manager/internal/shell"
	"github/DiscreteTom/ble-raspi-manager/internal/transport"
	"strings"

	"github.com/google/uuid"
	"tinygo.org/x/bluetooth"
)

type wifiInfo struct {
	SSID string // wifi name
	PSK  string // wifi password
}

type request struct {
	RefreshOnly bool
	WIFIs       []*wifiInfo
	StaticIP    string
	Router      string
}

type state struct {
	CurrentWIFI string // current wifi ssid
	CurrentIP   string // current ip address
	Router      string
	StaticIP    string // configured static ip

	WIFIs []*wifiInfo
}

func getState() state {
	return state{
		CurrentWIFI: getCurrentWIFI(),
		CurrentIP:   getCurrentIP(),
		Router:      getRouter(),
		StaticIP:    getStaticIP(),
		WIFIs:       getWifiInfo(),
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

func getWifiInfo() []*wifiInfo {
	ssids := strings.Split(getCommandOutput("cat /etc/wpa_supplicant/wpa_supplicant.conf | grep -Eo 'ssid=\"[^\"]*\"' | cut -d'\"' -f 2"), "\n")
	result := []*wifiInfo{}
	for _, ssid := range ssids {
		if len(ssid) != 0 {
			result = append(result, &wifiInfo{
				SSID: ssid,
				PSK:  "",
			})
		}
	}
	return result
}

func getCurrentWIFI() string {
	return getCommandOutput("iwgetid -r")
}

func getStaticIP() string {
	return getCommandOutput("cat /etc/dhcpcd.conf | grep -Eo '^static ip_address=.*' | cut -d'=' -f 2")
}

func getRouter() string {
	return getCommandOutput("netstat -nr | awk '$1 == \"0.0.0.0\"{print$2}'")
}

func setWifiInfo(wifis []*wifiInfo) {
	if len(getWifiInfo()) != 0 {
		// remove original settings
		shell.MustRunCommand("cp /etc/wpa_supplicant/wpa_supplicant.conf /etc/wpa_supplicant/wpa_supplicant.conf.backup")
		shell.MustRunCommand("cat /etc/wpa_supplicant/wpa_supplicant.conf.backup | grep -v 'network=' | grep -v 'ssid=' | grep -v 'psk=' | grep -v '}' > /etc/wpa_supplicant/wpa_supplicant.conf")
	}
	// write new settings
	for _, wifi := range wifis {
		shell.MustRunCommand(fmt.Sprintf("wpa_passphrase '%s' '%s' | grep -v '#psk' >> /etc/wpa_supplicant/wpa_supplicant.conf", wifi.SSID, wifi.PSK))
	}
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

	reader := &transport.ReadHandler{}
	writer := transport.NewWriteHandler(func(uuid, content []byte) {
		req := &request{}
		json.Unmarshal(content, req)
		current := getState()

		if req.RefreshOnly {
			bytes, err := json.Marshal(current)
			if err != nil {
				bytes = []byte(err.Error())
			}
			reader = transport.NewReadHandler(uuid, bytes)
			return
		}

		// check WIFIs
		if len(req.WIFIs) != len(current.WIFIs) {
			setWifiInfo(req.WIFIs)
		} else {
			for i := range req.WIFIs {
				if req.WIFIs[i].SSID != current.WIFIs[i].SSID || req.WIFIs[i].PSK != "" {
					setWifiInfo(req.WIFIs)
					break
				}
			}
		}

		// check cancel static ip
		if len(req.StaticIP) == 0 && len(current.StaticIP) != 0 {
			cancelStaticIp(true)
		}

		// check set new static ip
		if len(req.StaticIP) != 0 && (req.StaticIP != current.StaticIP || req.Router != current.Router) {
			setNewStaticIP(req.StaticIP, req.Router)
		}
	})

	return bluetooth.CharacteristicConfig{
		UUID:  wifiCharBleUUID,
		Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicReadPermission,
		ReadEvent: func(client bluetooth.Connection) ([]byte, error) {
			return reader.Read(), nil
		},
		WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
			writer.Write(value)
		},
	}
}
