package mpv

func connect(p string) (net.Conn, error) {
	return net.DialTimeout("unix", path, time.Second)
}

func runMPV(pathToFile string, debug bool) (*exec.Cmd, error) {
	cmd := exec.Command("mpv", pathToFile, `--no-video`, `--input-ipc-server=\\.\pipe\ledsimsocket`, `--pause`)
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	return cmd, nil
}
