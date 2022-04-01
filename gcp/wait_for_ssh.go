package gcp

import (
	"log"
	"time"
	"fmt"
	"os/user"
	"io/ioutil"

	"github.com/rjkroege/gocloud/config"
	"golang.org/x/crypto/ssh"
)

// TODO(rjk): setup something to store state for the connection
// Use the ssh.PublicKeys client config
// This really is starting to need some tests. Right?
// Dump the stuff returned about the instance to see if it contains some kind of public key
// I did this for spu and the machine's keys are not available.


// MakeSshClientConfig populates an ssh.ClientConfig for reuse by each
// connection attempt.
func MakeSshClientConfig(settings *config.Settings, ni *NodeInfo) (*ssh.ClientConfig, error) {
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
		User: userinfo.Username,

		Timeout: time.Second,

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
func WaitForSsh(settings *config.Settings, ni *NodeInfo) (*ssh.Client, error)  {
	log.Println("run WaitForSsh")
	sshconf, err := MakeSshClientConfig(settings, ni)
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
		case err != nil:  // and more stuffs.
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
