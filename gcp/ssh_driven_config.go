package gcp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/rjkroege/gocloud/config"
	"golang.org/x/crypto/ssh"
)

const metabase = "http://metadata.google.internal/computeMetadata/v1/instance/attributes/"

// TODO(rjk): This needs tests that can run locally. For that, I'd need a
// mock ssh server and a mock metadata service? (Yes?)

// ConfigureViaSsh invokes the specified command string via ssh to
// perform additional configuration of the target node. Significant
// additional featurism is possible.
func ConfigureViaSsh(settings *config.Settings, ni *NodeInfo, client *ssh.Client) error {
	// I have no way of knowing the hostKey because I didn't set it. The
	// system is newly launched and it makes the key for itself. But: I could
	// make a bespoke key. Then, the "public" key would also be private. Or I
	// could set some other kind of key and read it back.
	//
	// I want to preserve the key and use it when reconnecting. I need to
	// verify that the node is who I think it is. I can set a secret _on_ the
	// node at creation (shortly before) and then discover if if it has the
	// secret?
	//
	// Given that the IP address comes over a secure connection, the only way
	// that an adversary could man-in-the-middle me is if a router between me and
	// Google has been misconfigured and can forward traffic to an arbitrary
	// third party. I must validate some kind of shared secret.

	pnm, err := config.GetNodeMetadata(
		config.NewNodeProxiedMetadataClient(NewSshProxiedTransport(client)))
	if err != nil {
		return fmt.Errorf("can't read proxied node metadata: %v", err)
	}

	gottoken := pnm.InstanceToken
	if gottoken != ni.Token {
		return fmt.Errorf("Got token %q, want %q. Maybe this is an IP hijack?", gottoken, ni.Token)
	}

	// Run mk to download and setup the node
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("can't make an ssh execution session: %v", err)
	}
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	cmd := settings.InstanceTypes[ni.ConfigName].PostSshConfig
	if cmd != "" {
		if err := session.Run(cmd); err != nil {
			return fmt.Errorf("can't run %q: %v", cmd, err)
		}
	}

	return nil
}

// TODO(rjk): Must refactor this to address the fact that I may want to do this over a
// different transport/client combo.
// So want to plumb the transport into the config top level interface.
// func readStingFromMetadata(entry string, sshclient *ssh.Client) (string, error) {
// 	path := metabase + entry
// 	client := &http.Client{
// 		Transport: NewSshProxiedTransport(sshclient),
// 	}
// 	req, err := http.NewRequest("GET", path, nil)
// 	req.Header.Add("Metadata-Flavor", "Google")
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return "", fmt.Errorf("can't fetch metadata %v: %v", path, err)
// 	}
// 
// 	buffy, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return "", fmt.Errorf("can't read metadata body %v: %v", path, err)
// 	}
// 	return string(buffy), nil
// }
// 
func NewSshProxiedTransport(client *ssh.Client) http.RoundTripper {
	dolly := http.DefaultTransport.(*http.Transport).Clone()

	dolly.DialContext = func(_ context.Context, network, addr string) (net.Conn, error) {
		return client.Dial(network, addr)
	}
	return dolly
}
