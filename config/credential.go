//go:build !darwin
// +build !darwin

package config

func getCredential() (string, error) {
	return "", nil
}
