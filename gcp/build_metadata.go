package gcp

import (
	"fmt"
	"io/ioutil"
	"os/user"

	"github.com/rjkroege/gocloud/config"
	"github.com/sanity-io/litter"
	compute "google.golang.org/api/compute/v1"
)

func makeMetadataObject(settings *config.Settings, configName string) (*compute.Metadata, error) {
	metas := make([]*compute.MetadataItems, 0)

	// username
	userinfo, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("can't get user: %v", err)
	}
	suname := string(userinfo.Username)
	metas = append(metas, &compute.MetadataItems{
		Key:   "username",
		Value: &suname,
	})

	// gitcredential (read from the keychain)
	cred, err := settings.GitCredential()
	if err != nil {
		return nil, fmt.Errorf("no git credential: %v", err)
	}
	metas = append(metas, &compute.MetadataItems{
		Key:   "gitcredential",
		Value: &cred,
	})

	// ssh key
	sshpath := settings.PublicKeyFile(userinfo.HomeDir)
	sshkey, err := ioutil.ReadFile(sshpath)
	if err != nil {
		return nil, fmt.Errorf("can't read ssh key %q: %v", sshpath, err)
	}
	ssshkey := string(sshkey)
	metas = append(metas, &compute.MetadataItems{
		Key:   "sshkey",
		Value: &ssshkey,
	})

	// user-data (from ween)
	// must be in the instance data
	userdatapath := settings.InstanceTypes[configName].UserDataFile
	if userdatapath == "" {
		return nil, fmt.Errorf("instancetype %q didn't specify userdatafile", configName)
	}
	userdata, err := ioutil.ReadFile(userdatapath)
	if err != nil {
		return nil, fmt.Errorf("can't read userdata file %q: %v", userdatapath, err)
	}
	suserdata := string(userdata)
	metas = append(metas, &compute.MetadataItems{
		Key:   "user-data",
		Value: &suserdata,
	})

	return &compute.Metadata{
		Items: metas,
	}, nil
}

func ShowMetadata(settings *config.Settings, configName string) error {
	metadata, err := makeMetadataObject(settings, configName)
	if err != nil {
		return err
	}

	litter.Dump(metadata)
	return nil
}
