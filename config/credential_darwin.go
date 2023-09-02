//go:build darwin
// +build darwin

package config

import (
	"fmt"
	"os/exec"
)

// After some experimentation, I discovered that the shell command
//
// 	security find-generic-password -s gocloud.liqui.org -g -w
//
// would retrieve the contents of the password set in the keychain. Use
// this instead of linking against a native library to mtinain a cgo-free
// build.
func getCredential() (string, error) {
	cmd := exec.Command("/usr/bin/security", "find-generic-password",  "-s",  "gocloud.liqui.org",  "-g",  "-w")
	data, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("can't run keychain inquiry %v", err)
	}

	return string(data), nil
}
