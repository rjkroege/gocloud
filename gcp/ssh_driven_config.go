package gcp

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"github.com/rjkroege/gocloud/config"
	"golang.org/x/crypto/ssh"
)

/*
const command = `curl -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/instancetoken`
*/

const metabase = "http://metadata.google.internal/computeMetadata/v1/instance/attributes/"

// TODO(rjk): This needs tests that can run locally. For that, I'd need a
// mock ssh server and a mock metadata service?

// Monstrous featurism is possible.
// TODO(rjk): support reconnection, remote forwarding, etc.?
func ConfigureViaSsh(settings *config.Settings, ni *NodeInfo, client *ssh.Client) error {
	log.Println("running ConfigureViaSsh")

	// Verify that the node is who I want it to be.
	gottoken, err := readStingFromMetadata("instancetoken", client)
	if err != nil {
		return fmt.Errorf("can't read the instancetoken: %v", err)
	}

	if gottoken != ni.Token {
		return fmt.Errorf("Got token %q, want %q. Maybe this is an IP hijack?", gottoken, ni.Token)
	}

	return nil
}

func readStingFromMetadata(entry string, sshclient *ssh.Client) (string, error) {
	path := metabase + entry
	client := &http.Client{
		Transport: NewSshProxiedTransport(sshclient),
	}
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

func NewSshProxiedTransport(client *ssh.Client) http.RoundTripper {
	dolly := http.DefaultTransport.(*http.Transport).Clone()

	dolly.DialContext = func(_ context.Context, network, addr string) (net.Conn, error) {
		return client.Dial(network, addr)
	}
	return dolly
}
