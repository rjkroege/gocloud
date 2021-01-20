package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Settings struct {
	ProjectId string `json:"projectid"`
	Zone      string `json:"zone"`
}

func Read(path string) (*Settings, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("no config file %q: %v", path, err)
	}

	settings := &Settings{}
	decoder := json.NewDecoder(fd)
	if err := decoder.Decode(settings); err != nil {
		return nil, fmt.Errorf("error parsing config %q: %v", path, err)
	}
	return settings, nil
}
