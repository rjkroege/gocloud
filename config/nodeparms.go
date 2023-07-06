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
	InstanceToken  string `json:"instancetoken"`
}

func GetNodeMetadata(client *http.Client) (*NodeMetadata, error) {
	nm, err := unifiedNodeMetadata(client)
	if err == nil {
		return nm, nil
	}

	nm, err = legacyNodeMetadata(client)
	if err != nil {
		return nil, err
	}
	return nm, nil
}

func unifiedNodeMetadata(client *http.Client) (*NodeMetadata, error) {
	return nil, fmt.Errorf("notimplemented")
}

// legacyNodeMetadata populates a NodeMetadata from the
// discrete metadata entries on a node.
func legacyNodeMetadata(client *http.Client) (*NodeMetadata, error) {
	username, err := readNodeMetadata("username", client)
	if err != nil {
		return nil, fmt.Errorf("can't get username %v", err)
	}
	log.Println("username", string(username))

	gitcred, err := readNodeMetadata("gitcredential", client)
	if err != nil {
		return nil, fmt.Errorf("can't get getcredential %v", err)
	}
	log.Println("gitcred", string(gitcred))

	sshkey, err := readNodeMetadata("sshkey", client)
	if err != nil {
		return nil, fmt.Errorf("can't get sshkey %v", err)
	}
	log.Println("sshkey", string(sshkey))

	rcloneconfig, err := readNodeMetadata("rcloneconfig", client)
	if err != nil {
		return nil, fmt.Errorf("can't get rcloneconfig %v", err)
	}
	log.Println("rcloneconfig", string(rcloneconfig))

	instancetoken, err := readNodeMetadata("instancetoken", client)
	if err != nil {
		return nil, fmt.Errorf("can't get rcloneconfig %v", err)
	}
	log.Println("instancetoken", string(instancetoken))

	return &NodeMetadata{
		Username:      string(username),
		GitCredential: string(gitcred),
		SshKey:        string(sshkey),
		RcloneConfig:  string(rcloneconfig),
		InstanceToken:  string(instancetoken),
	}, nil

}

func NewNodeDirectMetadataClient() *http.Client {
	// Timeout should reduce the time to discover that a Linux machine is not
	// a GCP instance.
	tr := &http.Transport{
		ResponseHeaderTimeout: 500 * time.Millisecond,
	}
	return  &http.Client{Transport: tr}
}

func NewNodeProxiedMetadataClient(sshtrans http.RoundTripper) *http.Client {
	return &http.Client{
		Transport: sshtrans,
	}
}

// TODO(rjk): Need to also write some kind of function to set the metadata.
func readNodeMetadata(entry string, client *http.Client) ([]byte, error) {
	// TODO(rjk): Is there some kind of better http library for this?
	path := metabase + entry

	req, err := http.NewRequest("GET", path, nil)
	req.Header.Add("Metadata-Flavor", "Google")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("can't fetch metadata %v: %v", path, err)
	}

	buffy, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read metadata body %v: %v", path, err)
	}
	return buffy, nil
}

const metabase = "http://metadata.google.internal/computeMetadata/v1/instance/attributes/"

func RunningInGcp(client *http.Client) bool {
	if _, err := readNodeMetadata("username", client); err == nil {
		return true
	}
	return false
}
