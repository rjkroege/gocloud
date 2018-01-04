package oauth

import (
	"fmt"
	"os"
)

// GetConfig retrieves configuration from the specified configfile path. 
func GetConfigMap(configfile string)  (map[string]string, error) {

	f, err := os.Open(configfile)
	if err != nil {
		return map[string]string{}, fmt.Errorf("GetConfigMap failed to open %s because %v", configfile, err)
	}
	defer f.Close()

	// I can configure this for more as I discover needing them.
	configmap := map[string]string{
		"client_secret": "",
		"client_id":     "",
	}
	if err := FindProperties(f, configmap); err != nil {
		return map[string]string{}, fmt.Errorf("GetConfigMap failed because %v", err)
	}
	return configmap, nil
}


