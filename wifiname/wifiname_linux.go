// +build linux

package wifiname

import (
	"io/ioutil"
	"os/exec"
	"strings"
	"regexp"
	"runtime"
)

func WifiName() string {
	platform := runtime.GOOS
	if platform == "darwin" {
		return forOSX()
	} else {
		return forLinux()
	}
}

func forOSX() string {
	const osxCmd = "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport"
	const osxArgs = "-I"
	cmd := exec.Command(osxCmd, osxArgs)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ""
	}

	// start the command after having set up the pipe
	if err := cmd.Start(); err != nil {
		return ""
	}

	var str string

	if b, err := ioutil.ReadAll(stdout); err == nil {
		str += string(b) + "\n"
	}

	r := regexp.MustCompile(`s*SSID: (.+)s*`)

	name := r.FindAllStringSubmatch(str, -1)

	if len(name) <= 1 {
		return ""
	} else {
		return name[1][1]
	}
}

func forLinux() string {
	const linuxCmd = "iwgetid"
	const linuxArgs = "--raw"
	cmd := exec.Command(linuxCmd, linuxArgs)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ""
	}

	// start the command after having set up the pipe
	if err := cmd.Start(); err != nil {
		return ""
	}

	var str string

	if b, err := ioutil.ReadAll(stdout); err == nil {
		str += string(b) + "\n"
	}

	name := strings.Replace(str, "\n", "", -1)

	return name
}
