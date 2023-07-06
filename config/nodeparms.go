package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// NodeMetadata is the metadata that we have communicated to the node.
type NodeMetadata map[string]string

func GetNodeMetadata(client *http.Client) (NodeMetadata, error) {
	nm, err := addNodeMetadatav1(client)
	if err != nil {
		return nil, err
	}

	// This might error but we don't care. We just use the previous version.
	addNodeMetadatav2(client, nm)
	return nm, nil
}


func addNodeMetadatav2(client *http.Client, nm NodeMetadata) {
	// TODO(rjk): Populate this with the new path.
	if err := addNodeMetadataImpl(client, nm, []string{
		"githost",
	}); err != nil {
		nm["githost"] = "https://git.liqui.org/rjkroege/scripts.git"
	}
}

func addNodeMetadataImpl(client *http.Client, nm NodeMetadata, keys []string) error {
	for _, k := range keys {
		v, err := readNodeMetadata(k, client)
		if err != nil {
			return  fmt.Errorf("can't get %s %v", k, err)
		}
		log.Println(k, "->", v)
		nm[k] = string(v)
	}
	return nil
}

// addNodeMetadatav1 populates a NodeMetadata from the
// discrete metadata entries on a node.
func addNodeMetadatav1(client *http.Client) (NodeMetadata, error) {
	nm := make(NodeMetadata)

	if err := addNodeMetadataImpl(client, nm, []string{
		"username",
		"gitcredential",
		"sshkey",
		"rcloneconfig",
		"instancetoken",
	}); err != nil {
		return nil, err	
	}
	return nm, nil
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
