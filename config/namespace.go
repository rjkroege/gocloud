package config

import (
	"log"
	"os/user"
)

// LocalNameSpace returns the local namespace string. Not having a user
// or home directory are treated as fatal errors.
func LocalNameSpace() string {
	uinfo, err := user.Current()
	if err != nil {
		log.Fatalf("Can't determine username: %v", err)
		return ""
	}

	return "/tmp/ns." + uinfo.Username + ".:0"
}
