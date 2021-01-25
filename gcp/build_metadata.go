package gcp

import (	
	"io/ioutil"
	"os/user"
	"fmt"

	compute "google.golang.org/api/compute/v1"
	"github.com/sanity-io/litter"
	"github.com/rjkroege/gocloud/config"
)

func makeMetadataObject(settings *config.Settings, configName string ) (*compute.Metadata, error) {
	metas := make([]*compute.MetadataItems, 0)
	
	// username
	userinfo, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("can't get user: %v", err)
	}
	suname := string(userinfo.Username)
	metas = append(metas, &compute.MetadataItems{
		Key: "username",
		Value: &suname,
	})

	// gitcredential (read from the keychain)
// TODO(rjk): mine the git credentials...
// I need a setter tool probably?
	

	// ssh key
	sshpath := settings.PublicKeyFile(userinfo.HomeDir)
	sshkey, err := ioutil.ReadFile(sshpath)
	if err != nil {
		return nil, fmt.Errorf("can't read ssh key %q: %v", sshpath, err)
	}
	ssshkey := string(sshkey)
	metas = append(metas, &compute.MetadataItems{
		Key: "sshkey",
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
		Key: "user-data",
		Value: &suserdata,
	})

	return &compute.Metadata{
		Items: metas,
	}, nil
}

func ShowMetadata(settings *config.Settings,configName string) error {
	metadata, err := makeMetadataObject(settings, configName)
	if err != nil {
		return err
	}

	litter.Dump(metadata)
	return nil
}

