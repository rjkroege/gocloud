package gcp

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rjkroege/gocloud/config"
	"golang.org/x/crypto/ssh"

	"log"
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

	gottoken := pnm["instancetoken"]
	if gottoken != ni.Token {
		return fmt.Errorf("Got token %q, want %q. Maybe this is an IP hijack?", gottoken, ni.Token)
	}

	return InstallViaSsh(settings, ni, client)
}

func NewSshProxiedTransport(client *ssh.Client) http.RoundTripper {
	dolly := http.DefaultTransport.(*http.Transport).Clone()

	dolly.DialContext = func(_ context.Context, network, addr string) (net.Conn, error) {
		return client.Dial(network, addr)
	}
	return dolly
}

func InstallViaSsh(settings *config.Settings, ni *NodeInfo, client *ssh.Client) error {
	// Run tar on the remote to extract the copied binaries.
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("can't make an ssh execution session: %v", err)
	}
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	inpipe, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("can't get an ssh in pipe: %v", err)
	}

	// tar up the scripts and binaries locally.
	go func() {
		defer inpipe.Close()
		if err := TarGZTools(inpipe); err != nil {
			// TODO(rjk): I think that I can do something better about exiting.
			log.Println("can't tar: %v", err)
		}
	}()

	// TODO(rjk): Should I take these parameters from the toml file?
	cmd := "cd / ; tar xzf -"
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("can't extract %q: %v", cmd, err)
	}

	// Start the supplementary services.
	// TODO(rjk): Consider controlling this from the toml file?
	for _, cmd := range []string{
		"sudo /usr/local/bin/sessionender",
		"sudo /usr/local/bin/cpud -pk /usr/local/keys/pk",
	} {
		session, err := client.NewSession()
		if err != nil {
			return fmt.Errorf("can't make an ssh execution session for %q: %v", cmd, err)
		}
		// Do the commands keep running?
		if err := session.Start(cmd); err != nil {
			return fmt.Errorf("can't Start %q: %v", cmd, err)
		}
	}
	return nil
}

type Paths struct {
	From    string
	To      string
	Pattern string
}

func TarGZTools(w io.Writer) error {
	zfd := gzip.NewWriter(w)
	defer zfd.Close()
	tw := tar.NewWriter(zfd)
	defer tw.Close()

	for _, ptho := range []Paths{
		{
			From:    "/usr/local/script",
			To:      "/usr/local/script",
			Pattern: "*",
		},
		{
			From:    "/Users/rjkroege/wrks/archive/bins/linux/amd64",
			To:      "/usr/local/bin",
			Pattern: "*",
		},
	} {
		dfs := os.DirFS(ptho.From)
		files, err := fs.Glob(dfs, ptho.Pattern)
		if err != nil {
			return fmt.Errorf("can't glob %q: %v", filepath.Join(ptho.From, ptho.Pattern), err)
		}

		for _, f := range files {
			file := filepath.Join(ptho.From, f)
			fi, err := os.Stat(file)
			if err != nil {
				return fmt.Errorf("can't stat %q: %v", file, err)
			}

			hdr := &tar.Header{
				Name: filepath.Join(ptho.To, f),
				Mode: int64(fi.Mode()),
				Size: fi.Size(),
			}
			if err := tw.WriteHeader(hdr); err != nil {
				return fmt.Errorf("fatal %v", err)
			}

			fd, err := os.Open(file)
			if err != nil {
				return fmt.Errorf("can't Open %q: %v", file, err)
			}
			defer fd.Close()

			if _, err := io.Copy(tw, fd); err != nil {
				return fmt.Errorf("can't copy %q: %v", file, err)
			}
		}
	}
	return nil
}
