package main

import "fmt"

type wifi struct {
	SSID      string // wifi name
	PSK       string // wifi password
	CurrentIP string // current ip address
	Router    string
	Static    bool   // whether ip is static
	StaticIP  string // configured static ip
}

func getWifiInfo() wifi {
	static, staticIP := getStaticIP()
	return wifi{
		SSID:      getSSID(),
		PSK:       getPSK(),
		CurrentIP: getCurrentIP(),
		Static:    static,
		StaticIP:  staticIP,
		Router:    getRouter(),
	}
}

func getCurrentIP() string {
	out, err := runCommand("ifconfig wlan0 | sed -En 's/127.0.0.1//;s/.*inet (addr:)?(([0-9]*\\.){3}[0-9]*).*/\\2/p'")
	if err != nil {
		return ""
	}
	return out
}

func getSSID() string {
	out, err := runCommand("cat /etc/wpa_supplicant/wpa_supplicant.conf | grep -Eo 'ssid=\"([^\"]*)\"'")
	if len(out) == 0 || err != nil {
		return ""
	}
	return out[6 : len(out)-2] // remove prefix `ssid="` and suffix `"\n`
}

func getPSK() string {
	out, err := runCommand("cat /etc/wpa_supplicant/wpa_supplicant.conf | grep -Eo 'psk=\"([^\"]*)\"'")
	if len(out) == 0 || err != nil {
		return out
	}
	return out[5 : len(out)-2] // remove prefix `psk="` and suffix `"\n`
}

func getStaticIP() (bool, string) {
	out, err := runCommand("cat /etc/dhcpcd.conf | grep -Eo '^static ip_address=.*'")
	if len(out) == 0 || err != nil {
		return false, ""
	}
	return true, out[18 : len(out)-1] // remove prefix `static ip_address=` and suffix `\n`
}

func getRouter() string {
	out, _ := runCommand("netstat -nr | awk '$1 == \"0.0.0.0\"{print$2}'")
	return out
}

func setNewWifi(ssid, psk string) {
	if len(getSSID()) != 0 {
		// remove original settings
		runCommand("cp /etc/wpa_supplicant/wpa_supplicant.conf /etc/wpa_supplicant/wpa_supplicant.conf.backup")
		runCommand("sudo cat /etc/wpa_supplicant/wpa_supplicant.conf.backup | grep -v network | grep -v ssid | grep -v psk | grep -v '}' > /etc/wpa_supplicant/wpa_supplicant.conf")
	}
	// write new settings
	runCommand(fmt.Sprintf(`
sudo cat << EOF >> /etc/wpa_supplicant/wpa_supplicant.conf
network={
	ssid="%s"
	psk="%s"
}
EOF`, ssid, psk))
	runCommand("wpa_cli -i wlan0 reconfigure") // restart interface to apply
}

func setNewStaticIP(ip, routers string) {
	cancelStaticIp(false)
	runCommand("echo 'interface wlan0' >> /etc/dhcpcd.conf")
	runCommand(fmt.Sprintf("echo 'static ip_address=%s' >> /etc/dhcpcd.conf", ip))
	runCommand(fmt.Sprintf("echo 'static routers=%s' >> /etc/dhcpcd.conf", routers))

	fmt.Printf("Reboot by set new static ip/routers: %s %s.", ip, routers)
	runCommand("reboot now")
}

func cancelStaticIp(reboot bool) {
	runCommand("cp /etc/dhcpcd.conf /etc/dhcpcd.conf.backup")
	runCommand("cat /etc/dhcpcd.conf.backup | grep -Ev '^interface wlan0' | grep -Ev '^static ip_address=' | grep -Ev '^static routers=' > /etc/dhcpcd.conf")
	if reboot {
		fmt.Println("Reboot by cancel static ip.")
		runCommand("reboot now")
	}
}
