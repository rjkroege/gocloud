package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type InstanceConfig struct {
	Family        string `toml:"family"`
	Hardware      string `toml:"hardware"`
	DiskSize      int64  `toml:"disksize,omitempty"`
	Zone          string `toml:"zone,omitempty"`
	Description   string `toml:"description,omitempty"`
	PostSshConfig string `toml:"postsshconfig,omitempty"`
	GitHost       string `toml:"githost,omitempty"`
	UserData      string `toml:"userdata,omitempty"`
}

type Settings struct {
	DefaultZone       string                    `toml:"defaultzone"`
	ProjectId         string                    `toml:"projectid"`
	InstanceTypes     map[string]InstanceConfig `toml:"instance"`
	SshPublicKeyFile  string                    `toml:"sshpublickey,omitempty"`
	SshPrivateKeyFile string                    `toml:"sshprivatekey,omitempty"`
	Credential        string                    `toml:"credential,omitempty"`
	DefaultUserData   string                    `toml:"defaultuserdata,omitempty"`
}

func Read(path string) (*Settings, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("no config file %q: %v", path, err)
	}

	settings := &Settings{}
	decoder := toml.NewDecoder(fd)
	if _, err := decoder.Decode(settings); err != nil {
		return nil, fmt.Errorf("error parsing config %q: %v", path, err)
	}

	// TODO(rjk): Validate the configurable settings.
	return settings, nil
}

// Zone returns the zone for this instancetype.
func (s *Settings) Zone(instancetype string) string {
	if z, ok := s.InstanceTypes[instancetype]; ok && z.Zone != "" {
		return z.Zone
	}
	return s.DefaultZone
}

func (s *Settings) UserData(instancetype string) string {
	if z, ok := s.InstanceTypes[instancetype]; ok && z.UserData != "" {
		return z.UserData
	}
	return s.DefaultUserData
}

// UniqueFamilies returns the unique families used in settings.
func (s *Settings) UniqueFamilies() []string {
	fm := make(map[string]struct{})
	for _, v := range s.InstanceTypes {
		fm[v.Family] = struct{}{}
	}
	fa := make([]string, 0)
	for k := range fm {
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
	return filepath.Join(home, ".ssh", "id_rsa.pub")
}

func (s *Settings) PrivateKeyFile(home string) string {
	if s.SshPrivateKeyFile != "" {
		if filepath.IsAbs(s.SshPrivateKeyFile) {
			return s.SshPrivateKeyFile
		} else {
			return filepath.Join(home, ".ssh", s.SshPrivateKeyFile)
		}
	}
	return filepath.Join(home, ".ssh", "id_rsa")
}

func (s *Settings) GitCredential() (string, error) {
	cred, err := getCredential()
	if err != nil {
		return "", err
	}
	if cred != "" {
		return cred, nil
	}
	return s.Credential, nil
}
