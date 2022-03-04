//go:build darwin || linux
// +build darwin linux

package mpv

import (
	"net"
	"os"
	"os/exec"
	"time"
)

func connect() (net.Conn, error) {
	return net.DialTimeout("unix", "/tmp/ledsimsocket", time.Second)
}

func runMPV(pathToFile string, mpvArg string, debug bool) (*exec.Cmd, error) {
	cmd := exec.Command("mpv", pathToFile, mpvArg, `--no-video`, `--input-ipc-server=/tmp/ledsimsocket`, `--pause`)

	if debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	return cmd, nil
}
