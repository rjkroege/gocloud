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

// convertMapToGcpFormat constructs a GCP compute.Metadata representation
// of a key-value map out of a key-value map of metadata attributes.
func convertMapToGcpFormat(metas map[string]string) *compute.Metadata {
	converted := make([]*compute.MetadataItems, 0)

	for k, v := range metas {
		// Taking the address of v is not well-defined.
		rv := v
		converted = append(converted, &compute.MetadataItems{
			Key:   k,
			Value: &rv,
		})
	}

	return &compute.Metadata{
		Items: converted,
	}
}

// makeMetadataObject makes a Go map of metadata key-value pairs.
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

	// githost, read from the configuration file.
	githost := settings.InstanceTypes[configName].GitHost	
	if githost != "" {
		metas["githost"] = githost
	}

	// gitcredential (read from the keychain)
	if cred, err := settings.GitCredential(); err != nil {
		fmt.Printf("can't add git credential to instance metadata because no git credential: %v", err)
	} else {
		metas["gitcredential"] = cred
	}

	// ssh key (always needed)
	sshpath := settings.PublicKeyFile(userinfo.HomeDir)
	sshkey, err := ioutil.ReadFile(sshpath)
	if err != nil {
		return nil, fmt.Errorf("can't read ssh key %q: %v", sshpath, err)
	}
	metas["sshkey"] = string(sshkey)

	// Ship rclone configuration to the client if it exists.
	rclonepath := filepath.Join(userinfo.HomeDir, ".config", "rclone", "rclone.conf")
	if rclonekey, err := ioutil.ReadFile(rclonepath); err != nil {
		fmt.Printf("not adding rclone config to instance metadata because can't read rclone config %q: %v", rclonepath, err)
	} else {
		metas["rcloneconfig"] = string(rclonekey)
	}

	// instance configuration data is required
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
	if kopiaauth, err := readKopiaConfiguration(); err != nil {
		fmt.Errorf("not adding kopia reconnection string to instance metadata because: %v", err)
	} else {
		metas["kopiareconnection"] = kopiaauth
	}

	return metas, nil
}

// ShowMetadata will display the metadata object.
func ShowMetadata(settings *config.Settings, configName string) error {
	metadata, err := makeMetadataObject(settings, configName)
	if err != nil {
		return err
	}

	litter.Dump(metadata)
	litter.Dump(convertMapToGcpFormat(metadata))
	return nil
}

// readKopiaConfiguration runs kopia to get a reconnection string.
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
