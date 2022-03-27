//go:build darwin
// +build darwin

package config

import (
	"fmt"
	"os/user"

	"github.com/keybase/go-keychain"
)

// readKeyChain reads a value from the keychain identified by service and
// accessgroup for username and returns the read value, true if there was
// a read value and an error if one occurred.
func readKeyChain(service, username, accessgroup string) ([]byte, bool, error) {
	query := keychain.NewItem()

	// Generic password type. I want this kind
	query.SetSecClass(keychain.SecClassGenericPassword)

	// The service name. I'm using gcs.liqui.org. Which is sort of made-up
	query.SetService(service)

	// The name of the current user.
	query.SetAccount(username)

	// This is suppose to be the team id (from signing / notarization) with
	// .org.liqui.mkconfig appended. I have made it up. It doesn't seem to matter.
	query.SetAccessGroup(accessgroup)

	// We only want one result
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnData(true)

	results, err := keychain.QueryItem(query)
	if err != nil {
		return nil, false,
			fmt.Errorf("tried to read keychain: %s,%s,%s didn't works: %v", service, username, accessgroup, err)
	} else if len(results) != 1 {
		return nil, false, nil
	}
	return results[0].Data, true, nil
}

func getCredential() (string, error) {
	userinfo, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("can't get the user name: %v", err)
	}

	data, exists, err := readKeyChain("gocloud.liqui.org", userinfo.Username, "groovy.org.liqui.gocloud")
	if err != nil {
		return "", fmt.Errorf("can't read credential from keychain: %v", err)
	} else if !exists {
		return "", fmt.Errorf("no keychain. Try adding a keychain login (i.e. \"New Password Item...\") application password for your account (i.e. your username) and name gocloud.liqui.org with git credential as password")
	}

	return string(data), nil
}
