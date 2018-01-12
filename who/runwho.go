package who

import (
	"os/exec"
)

func RunWho() ([]byte, error) {
	cmd := exec.Command("/usr/bin/who")
	return cmd.Output()
}

