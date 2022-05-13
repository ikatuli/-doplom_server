package clamav

import(
	"os/exec"
)

func Status () string {
	cmd:=exec.Command("systemctl","is-active","clamav-daemon.service")
	stdout, _ := cmd.Output()
	tmp:=string(stdout)
	return tmp[0:len(tmp)-1]
}

func Journal () string {
	cmd:=exec.Command("journalctl","-b","-u","clamav-daemon.service")
	stdout, _ := cmd.Output()
	return string(stdout)
}

func Update () (string, error) {
	cmd:=exec.Command("freshclam")
	stdout, err := cmd.Output()
	return string(stdout),err
}

func Start (action string) error {
	cmd:=exec.Command("systemctl",action,"clamav-daemon.service")
	_, err := cmd.Output()
	return err
}
