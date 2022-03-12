package main

import (
	"bytes"
	"os/exec"
)

func getCurrentIP() string {
	return runCommand("bash", "-c", "ifconfig wlan0 | sed -En 's/127.0.0.1//;s/.*inet (addr:)?(([0-9]*\\.){3}[0-9]*).*/\\2/p'")
}

func getSSID() string {
	out := runCommand("bash", "-c", "cat /etc/wpa_supplicant/wpa_supplicant.conf | grep -Eo 'ssid=\"([^\"]*)\"'")
	if len(out) == 0 {
		return out
	}
	return out[6 : len(out)-1] // remove prefix `ssid="` and suffix `"`
}

func getPSK() string {
	out := runCommand("bash", "-c", "cat /etc/wpa_supplicant/wpa_supplicant.conf | grep -Eo 'psk=\"([^\"]*)\"'")
	if len(out) == 0 {
		return out
	}
	return out[5 : len(out)-1] // remove prefix `psk="` and suffix `"`
}

func getStaticIP() (bool, string) {
	out := runCommand("bash", "-c", "cat /etc/dhcpd.conf | grep -o 'static ip_address=.*'")
	if len(out) == 0 {
		return false, ""
	}
	return true, out[18:] // remove prefix `static ip_address=`
}

func runCommand(name string, arg ...string) (stdout string) {
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	cmd := exec.Command(name, arg...)
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if err := cmd.Run(); err != nil {
		panic(err)
	}

	stdout = stdoutBuf.String()
	return
}
