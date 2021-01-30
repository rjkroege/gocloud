package config

import (
	"testing"
	"path/filepath"
	"io/ioutil"

	"github.com/sergi/go-diff/diffmatchpatch"
)

const createcase = `
#-- gocloud instancename --
Host instancename
	HostName 10.0.0.1
	ControlPath ~/.ssh/controlmasters/%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
#---
`

const appendcase = `
#-- gocloud instancename --
Host instancename
	HostName 10.0.0.1
	ControlPath ~/.ssh/controlmasters/%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
#---

#-- gocloud secondinstance --
Host secondinstance
	HostName 10.0.0.2
	ControlPath ~/.ssh/controlmasters/%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
#---
`

const replacecase = `
#-- gocloud instancename --
Host instancename
	HostName 10.0.0.1
	ControlPath ~/.ssh/controlmasters/%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
#---

#-- gocloud secondinstance --
Host secondinstance
	HostName 10.0.1.3
	ControlPath ~/.ssh/controlmasters/%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
#---

#-- gocloud suffixinstance --
Host suffixinstance
	HostName 10.0.0.3
	ControlPath ~/.ssh/controlmasters/%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
#---
`

const secondappend = `
#-- gocloud instancename --
Host instancename
	HostName 10.0.0.1
	ControlPath ~/.ssh/controlmasters/%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
#---

#-- gocloud secondinstance --
Host secondinstance
	HostName 10.0.0.2
	ControlPath ~/.ssh/controlmasters/%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
#---

#-- gocloud suffixinstance --
Host suffixinstance
	HostName 10.0.0.3
	ControlPath ~/.ssh/controlmasters/%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
#---
`

const replacefirst = `
#-- gocloud instancename --
Host instancename
	HostName 10.0.2.1
	ControlPath ~/.ssh/controlmasters/%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
#---

#-- gocloud secondinstance --
Host secondinstance
	HostName 10.0.1.3
	ControlPath ~/.ssh/controlmasters/%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
#---

#-- gocloud suffixinstance --
Host suffixinstance
	HostName 10.0.0.3
	ControlPath ~/.ssh/controlmasters/%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
#---
`

const replaceend = `
#-- gocloud instancename --
Host instancename
	HostName 10.0.2.1
	ControlPath ~/.ssh/controlmasters/%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
#---

#-- gocloud secondinstance --
Host secondinstance
	HostName 10.0.1.3
	ControlPath ~/.ssh/controlmasters/%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
#---

#-- gocloud suffixinstance --
Host suffixinstance
	HostName 10.0.3.3
	ControlPath ~/.ssh/controlmasters/%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
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
	contents, err :=  ioutil.ReadFile(newfile)
	if err != nil {
		t.Fatal("didn't make a file, create case", err)
	}
	
	if got, want := string(contents), createcase; got != want {
		t.Errorf("create case: got %q want %q", got, want)
	}

	// Append a block
	if err := insertNameBlock(newfile, makeFieldValues("secondinstance", "10.0.0.2")); err != nil {
		t.Fatal("can't create new file?", err)
	}

	// Validate that it's correct.
	contents, err =  ioutil.ReadFile(newfile)
	if err != nil {
		t.Fatal("didn't make a file, create case", err)
	}
	
	if got, want := string(contents), appendcase; got != want {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(got, want, false)
		t.Errorf("appendcase: mismatch\n%s\n%v", got, diffs)
	}

	// Append an extra block
	if err := insertNameBlock(newfile, makeFieldValues("suffixinstance", "10.0.0.3")); err != nil {
		t.Fatal("can't create new file?", err)
	}

	// Validate that it had the right contents.
	contents, err =  ioutil.ReadFile(newfile)
	if err != nil {
		t.Fatal("didn't make a file, create case", err)
	}
	if got, want := string(contents), secondappend; got != want {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(got, want, false)
		t.Errorf("appendcase: mismatch\n%s\n%v", got, diffs)
	}

	// Replace a block
	if err := insertNameBlock(newfile, makeFieldValues("secondinstance", "10.0.1.3")); err != nil {
		t.Fatal("can't create new file?", err)
	}

	// Validate
	contents, err =  ioutil.ReadFile(newfile)
	if err != nil {
		t.Fatal("didn't make a file, create case", err)
	}
	if got, want := string(contents), replacecase; got != want {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(got, want, false)
		t.Errorf("appendcase: mismatch\n%s\n%v", got, diffs)
	}

	
	// Replace a block at the beginning
	if err := insertNameBlock(newfile, makeFieldValues("instancename", "10.0.2.1")); err != nil {
		t.Fatal("can't create new file?", err)
	}

	// Validate
	contents, err =  ioutil.ReadFile(newfile)
	if err != nil {
		t.Fatal("didn't make a file, create case", err)
	}
	if got, want := string(contents), replacefirst; got != want {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(got, want, false)
		t.Errorf("appendcase: mismatch\n%s\n%v", got, diffs)
	}

	// Replace a block at the end
	if err := insertNameBlock(newfile, makeFieldValues("suffixinstance", "10.0.3.3")); err != nil {
		t.Fatal("can't create new file?", err)
	}

	// Validate
	contents, err =  ioutil.ReadFile(newfile)
	if err != nil {
		t.Fatal("didn't make a file, create case", err)
	}
	if got, want := string(contents), replaceend; got != want {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(got, want, false)
		t.Errorf("appendcase: mismatch\n%s\n%v", got, diffs)
	}
}
