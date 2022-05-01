package config

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const createcase = `
#-- gocloud instancename --
Host instancename
	HostName 10.0.0.1
	ControlPath ~/.ssh/controlmasters/instancename-%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
	StrictHostKeyChecking no
#---
`

const appendcase = `
#-- gocloud instancename --
Host instancename
	HostName 10.0.0.1
	ControlPath ~/.ssh/controlmasters/instancename-%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
	StrictHostKeyChecking no
#---

#-- gocloud secondinstance --
Host secondinstance
	HostName 10.0.0.2
	ControlPath ~/.ssh/controlmasters/secondinstance-%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
	StrictHostKeyChecking no
#---
`

const replacecase = `
#-- gocloud instancename --
Host instancename
	HostName 10.0.0.1
	ControlPath ~/.ssh/controlmasters/instancename-%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
	StrictHostKeyChecking no
#---

#-- gocloud secondinstance --
Host secondinstance
	HostName 10.0.1.3
	ControlPath ~/.ssh/controlmasters/secondinstance-%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
	StrictHostKeyChecking no
#---

#-- gocloud suffixinstance --
Host suffixinstance
	HostName 10.0.0.3
	ControlPath ~/.ssh/controlmasters/suffixinstance-%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
	StrictHostKeyChecking no
#---
`

const secondappend = `
#-- gocloud instancename --
Host instancename
	HostName 10.0.0.1
	ControlPath ~/.ssh/controlmasters/instancename-%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
	StrictHostKeyChecking no
#---

#-- gocloud secondinstance --
Host secondinstance
	HostName 10.0.0.2
	ControlPath ~/.ssh/controlmasters/secondinstance-%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
	StrictHostKeyChecking no
#---

#-- gocloud suffixinstance --
Host suffixinstance
	HostName 10.0.0.3
	ControlPath ~/.ssh/controlmasters/suffixinstance-%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
	StrictHostKeyChecking no
#---
`

const replacefirst = `
#-- gocloud instancename --
Host instancename
	HostName 10.0.2.1
	ControlPath ~/.ssh/controlmasters/instancename-%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
	StrictHostKeyChecking no
#---

#-- gocloud secondinstance --
Host secondinstance
	HostName 10.0.1.3
	ControlPath ~/.ssh/controlmasters/secondinstance-%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
	StrictHostKeyChecking no
#---

#-- gocloud suffixinstance --
Host suffixinstance
	HostName 10.0.0.3
	ControlPath ~/.ssh/controlmasters/suffixinstance-%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
	StrictHostKeyChecking no
#---
`

const replaceend = `
#-- gocloud instancename --
Host instancename
	HostName 10.0.2.1
	ControlPath ~/.ssh/controlmasters/instancename-%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
	StrictHostKeyChecking no
#---

#-- gocloud secondinstance --
Host secondinstance
	HostName 10.0.1.3
	ControlPath ~/.ssh/controlmasters/secondinstance-%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
	StrictHostKeyChecking no
#---

#-- gocloud suffixinstance --
Host suffixinstance
	HostName 10.0.3.3
	ControlPath ~/.ssh/controlmasters/suffixinstance-%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
	StrictHostKeyChecking no
#---
`

func TestInsertNameBlock(t *testing.T) {
	dir := t.TempDir()

	newfile := filepath.Join(dir, "newconfig")

	// Create a file.
	if err := insertNameBlock(newfile, makeFieldValues("instancename", "10.0.0.1")); err != nil {
		t.Fatal("can't create new file?", err)
	}

	// Validate that it's correct.
	contents, err := ioutil.ReadFile(newfile)
	if err != nil {
		t.Fatal("didn't make a file, create case", err)
	}

	if diff := cmp.Diff(createcase, string(contents)); diff != "" {
		t.Errorf("createcase:  mismatch (-want +got):\n%s", diff)
	}

	// Append a block
	if err := insertNameBlock(newfile, makeFieldValues("secondinstance", "10.0.0.2")); err != nil {
		t.Fatal("can't create new file?", err)
	}

	// Validate that it's correct.
	contents, err = ioutil.ReadFile(newfile)
	if err != nil {
		t.Fatal("didn't make a file, create case", err)
	}

	if diff := cmp.Diff(appendcase, string(contents)); diff != "" {
		t.Errorf("appendcase:  mismatch (-want +got):\n%s", diff)
	}

	// Append an extra block
	if err := insertNameBlock(newfile, makeFieldValues("suffixinstance", "10.0.0.3")); err != nil {
		t.Fatal("can't create new file?", err)
	}

	// Validate that it had the right contents.
	contents, err = ioutil.ReadFile(newfile)
	if err != nil {
		t.Fatal("didn't make a file, create case", err)
	}
	if diff := cmp.Diff(secondappend, string(contents)); diff != "" {
		t.Errorf("secondappend:  mismatch (-want +got):\n%s", diff)
	}

	// Replace a block
	if err := insertNameBlock(newfile, makeFieldValues("secondinstance", "10.0.1.3")); err != nil {
		t.Fatal("can't create new file?", err)
	}

	// Validate
	contents, err = ioutil.ReadFile(newfile)
	if err != nil {
		t.Fatal("didn't make a file, create case", err)
	}
	if diff := cmp.Diff(replacecase, string(contents)); diff != "" {
		t.Errorf("replacecase:  mismatch (-want +got):\n%s", diff)
	}

	// Replace a block at the beginning
	if err := insertNameBlock(newfile, makeFieldValues("instancename", "10.0.2.1")); err != nil {
		t.Fatal("can't create new file?", err)
	}

	// Validate
	contents, err = ioutil.ReadFile(newfile)
	if err != nil {
		t.Fatal("didn't make a file, create case", err)
	}
	if diff := cmp.Diff(replacefirst, string(contents)); diff != "" {
		t.Errorf("replacefirst:  mismatch (-want +got):\n%s", diff)
	}

	// Replace a block at the end
	if err := insertNameBlock(newfile, makeFieldValues("suffixinstance", "10.0.3.3")); err != nil {
		t.Fatal("can't create new file?", err)
	}

	// Validate
	contents, err = ioutil.ReadFile(newfile)
	if err != nil {
		t.Fatal("didn't make a file, create case", err)
	}
	if diff := cmp.Diff(replaceend, string(contents)); diff != "" {
		t.Errorf("replaceend:  mismatch (-want +got):\n%s", diff)
	}
}
