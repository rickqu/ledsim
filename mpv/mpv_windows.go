package mpv

import (
	"net"
	"os"
	"os/exec"

	"github.com/Microsoft/go-winio"
)

func connect() (net.Conn, error) {
	return winio.DialPipe(`\\.\pipe\ledsimsocket`, nil)
}

func runMPV(pathToFile string, mpvArg string, debug bool) (*exec.Cmd, error) {
	cmd := exec.Command("mpv.exe", mpvArg, pathToFile, `--no-video`, `--input-ipc-server=\\.\pipe\ledsimsocket`, `--pause`)

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
