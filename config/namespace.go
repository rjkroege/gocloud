package config

import (
	"log"
	"os/user"
)

// LocalNameSpace returns the local namespace string. Not having a user
// or home directory are treated as fatal errors.
func LocalNameSpace(username string) string {
	if username == "" {
		uinfo, err := user.Current()
		if err != nil {
			log.Fatalf("Can't determine username: %v", err)
			return ""
		}
		username = uinfo.Username
	}

	return "/tmp/ns." + username + ".:0"
}
