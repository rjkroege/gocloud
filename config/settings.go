package config

import (
	"encoding/json"
	"path/filepath"
	"fmt"
	"os"
)


type InstanceConfig struct {
	Family string `json:"family"`
	Hardware string `json:"hardware"`
	Zone      string `json:"zone,omitempty"`
	Description string `json:"description,omitempty"`
	UserDataFile string `json:"userdatafile,omitempty"`
}

type Settings struct {
	DefaultZone string `json:"defaultzone"`
	ProjectId string `json:"projectid"`
	InstanceTypes map[string]InstanceConfig  `json:"instancetypes"`
	SshPublicKeyFile string `json:"sshpublickey,omitempty"`
	Credential string `json:"credential,omitempty"`
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

	// TODO(rjk): Validate.
	return settings, nil
}

// Zone returns the zone for this instancetype.
func (s *Settings) Zone(instancetype string) string {
	if z, ok := s.InstanceTypes[instancetype]; ok && z.Zone != "" {
		return z.Zone
	}
	return s.DefaultZone
}

// UniqueFamilies returns the unique families used in settings.
func (s *Settings) UniqueFamilies() []string {
	fm := make(map[string]struct{})
	for _, v := range s.InstanceTypes {
		fm[v.Family] = struct{}{}
	}
	fa := make([]string, 0)
	for k, _ := range fm {
		fa = append(fa, k)
	}
	return fa
}

func (s *Settings) Description(instancetype, name string) string {
	ins := s.InstanceTypes[instancetype]
	if ins.Description != "" {
		return ins.Description
	}
	return fmt.Sprintf("%s: %s %s %s instance", name, instancetype, ins.Family, ins.Hardware)
}

func (s *Settings) PublicKeyFile(home string) string {
	if s.SshPublicKeyFile != "" {
		
		if filepath.IsAbs(s.SshPublicKeyFile) {
			return s.SshPublicKeyFile
		} else {
			return filepath.Join(home, ".ssh", s.SshPublicKeyFile)
		}
	}
	return filepath.Join(home, ".ssh",  "id_rsa.pub")
}

func (s *Settings) GitCredential() (string, error) {
	cred, err := getCredential()
	if err != nil {
		return "", err
	}
	if cred  != "" {
		return cred, nil
	}
	return s.Credential, nil
}
