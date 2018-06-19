// +build windows

package wifiname

import (
	"io/ioutil"
	"os/exec"
	"regexp"
	"syscall"
)

func WifiName() string {
	return forWindows()
}

func forWindows() string {
	windowsCmd := "cmd.exe"
	windowsArgs := []string{"/c", "netsh", "wlan", "show", "interface"}
	cmd := exec.Command(windowsCmd, windowsArgs...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
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
		str += (string(b) + "\n")
	}

	re := regexp.MustCompile("\\sSSID\\s*:\\s*(\\S+)")
	result := re.FindStringSubmatch(str)

	if len(result) < 1 {
		return ""
	}

	return result[1]
}

