package config


import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// NodeMetadata is serializable parameters for transmission to the node.
type NodeMetadata struct {
	Username      string `json:"username"`
	GitCredential string `json:"gitcredential"`
	SshKey        string `json:"sshkey"`
	RcloneConfig  string `json:"rcloneconfig"`
}

func GetNodeMetadata() (*NodeMetadata, error) {
	nm, err := unifiedNodeMetadata()
	if err == nil {
		return nm, nil
	}

	nm, err = legacyNodeMetadata()
	if err != nil {
		return nil, err
	}
	return nm, nil
}

func unifiedNodeMetadata() (*NodeMetadata, error) {
	return nil, fmt.Errorf("notimplemented")
}

// legacyNodeMetadata populates a NodeMetadata from the
// discrete metadata entries on a node.
func legacyNodeMetadata() (*NodeMetadata, error) {
	username, err := readStringFromMetadata("username")
	if err != nil {
		return nil, fmt.Errorf("can't get username %v", err)
	}
	log.Println("username", username)

	gitcred, err := readStringFromMetadata("gitcredential")
	if err != nil {
		return nil, fmt.Errorf("can't get getcredential %v", err)
	}
	log.Println("gitcred", gitcred)

	sshkey, err := readStringFromMetadata("sshkey")
	if err != nil {
		return nil, fmt.Errorf("can't get sshkey %v", err)
	}
	log.Println("sshkey", sshkey)

	rcloneconfig, err := readStringFromMetadata("rcloneconfig")
	if err != nil {
		return nil, fmt.Errorf("can't get rcloneconfig %v", err)
	}
	log.Println("rcloneconfig", rcloneconfig)

	return &NodeMetadata{
		Username:      username,
		GitCredential: gitcred,
		SshKey:        sshkey,
		RcloneConfig:  rcloneconfig,
	}, nil

}

// TODO(rjk): Need to also write some kind of function to set the metadata.
func readStringFromMetadata(entry string) (string, error) {
	path := metabase + entry

	// Adjust this for the timeout.
	// This should be faster now.
	tr := &http.Transport{
		ResponseHeaderTimeout: 500 * time.Millisecond,
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", path, nil)
	req.Header.Add("Metadata-Flavor", "Google")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("can't fetch metadata %v: %v", path, err)
	}

	buffy, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("can't read metadata body %v: %v", path, err)
	}
	return string(buffy), nil
}

const metabase = "http://metadata.google.internal/computeMetadata/v1/instance/attributes/"

func RunningInGcp() bool {
	if _, err := readStringFromMetadata("username"); err == nil {
		return true
	}
	return false
}
