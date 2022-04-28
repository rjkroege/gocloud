package gcp

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/user"
	"time"

	"github.com/rjkroege/gocloud/config"
	"golang.org/x/crypto/ssh"
)

// TODO(rjk): Figure out how to write nice tests for all of this.

// MakeSshClientConfig populates an ssh.ClientConfig for reuse by each
// connection attempt.
func MakeSshClientConfig(settings *config.Settings) (*ssh.ClientConfig, error) {
	// username
	userinfo, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("can't get user: %v", err)
	}

	// Read the private ssh key
	sshpath := settings.PrivateKeyFile(userinfo.HomeDir)
	sshkey, err := ioutil.ReadFile(sshpath)
	if err != nil {
		return nil, fmt.Errorf("can't read ssh key %q: %v", sshpath, err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(sshkey)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key %q: %v", sshpath, err)
	}

	config := &ssh.ClientConfig{
		// needs to come out of the right place
		User:            userinfo.Username,
		Timeout:         time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
	}
	return config, nil
}

// WaitForSsh waits for a ssh server to be up on the newly created node.
// Run this after making the node.
func WaitForSsh(settings *config.Settings, ni *NodeInfo) (*ssh.Client, error) {
	log.Println("run WaitForSsh")
	sshconf, err := MakeSshClientConfig(settings)
	if err != nil {
		return nil, fmt.Errorf("can't MakeSshClientConfig: %v", err)
	}

	// wait for the ssh to come up
	for i := 0; i < 12; i++ {
		// Wait a bit for the GCP to have done something.
		delayms := time.Duration(64*(1<<i)) * time.Millisecond
		log.Printf("wating for ssh %v...", delayms)
		delay := time.NewTimer(delayms)
		<-delay.C

		log.Println("polling for the instance ssh up")

		switch client, err := connectToSsh(sshconf, ni.Ssh()); {
		case err == nil:
			log.Println("ssh is running")
			return client, nil
		case err != nil: // and more stuffs.
			log.Printf("no ssh yet %v", err)
		}
	}
	return nil, fmt.Errorf("too many tries failing to get ssh for %s", ni.Name)
}

func connectToSsh(sshconf *ssh.ClientConfig, addr string) (*ssh.Client, error) {
	client, err := ssh.Dial("tcp", addr, sshconf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
