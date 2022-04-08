package gcp

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"

	"github.com/rjkroege/gocloud/config"
	"github.com/sanity-io/litter"
	compute "google.golang.org/api/compute/v1"
)

func convertMapToGcpFormat(metas map[string]string) *compute.Metadata {
	converted := make([]*compute.MetadataItems, 0)

	for k, v := range metas {
		// Taking the address of v is not well-defined.
		rv := v
		// log.Printf("key[%q] = %#v", k, v)
		converted = append(converted, &compute.MetadataItems{
			Key:   k,
			Value: &rv,
		})
	}

	return &compute.Metadata{
		Items: converted,
	}
}

func makeMetadataObject(settings *config.Settings, configName string) (map[string]string, error) {
	metas := make(map[string]string)

	// username
	userinfo, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("can't get user: %v", err)
	}
	metas["username"] = string(userinfo.Username)

	// unique token identifying this node
	rawtoken := make([]byte, 16)
	_, err = rand.Read(rawtoken)
	if err != nil {
		return nil, fmt.Errorf("can't make random token: %v", err)
	}
	// There can be nulls in token so encode.
	metas["instancetoken"] = base64.StdEncoding.EncodeToString(rawtoken)

	// TODO(rjk): Some of these should be optional.
	// gitcredential (read from the keychain)
	cred, err := settings.GitCredential()
	if err != nil {
		return nil, fmt.Errorf("no git credential: %v", err)
	}
	metas["gitcredential"] = cred

	// ssh key
	sshpath := settings.PublicKeyFile(userinfo.HomeDir)
	sshkey, err := ioutil.ReadFile(sshpath)
	if err != nil {
		return nil, fmt.Errorf("can't read ssh key %q: %v", sshpath, err)
	}
	metas["sshkey"] = string(sshkey)

	// Ship rclone configuration to the client.
	rclonepath := filepath.Join(userinfo.HomeDir, ".config", "rclone", "rclone.conf")
	rclonekey, err := ioutil.ReadFile(rclonepath)
	if err != nil {
		return nil, fmt.Errorf("can't read rclone config %q: %v", rclonepath, err)
	}
	metas["rcloneconfig"] = string(rclonekey)

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
	metas["user-data"] = string(userdata)

	// Insert the kopia connect restoration code.
	kopiaauth, err := readKopiaConfiguration()
	if err != nil {
		return nil, fmt.Errorf("kopia connection command not available: %v", err)
	}
	metas["kopiareconnection"] = kopiaauth

	return metas, nil
}

func ShowMetadata(settings *config.Settings, configName string) error {
	metadata, err := makeMetadataObject(settings, configName)
	if err != nil {
		return err
	}

	litter.Dump(metadata)
	litter.Dump(convertMapToGcpFormat(metadata))
	return nil
}

func readKopiaConfiguration() (string, error) {
	cmd := exec.Command("kopia", "repository", "status", "-ts")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("can't run the kopia commandL: %v", err)
	}

	const regexpsrc = "\n\\$(.*)\n"
	re := regexp.MustCompile(regexpsrc)

	res := re.FindSubmatch(output)

	if len(res) != 2 {
		return "", fmt.Errorf("can't find the kopia auth string in the spew")
	}
	return string(res[1]), nil
}
