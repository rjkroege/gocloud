package config

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"text/template"
)

type fieldValues struct {
	Name   string
	IP     string
	Header string
	Footer string
}

// TODO(rjk): Consider letting the block innards be specified by the config file.
const machineblock = `
{{.Header}}
Host {{.Name}}
	HostName {{.IP}}
	ControlPath ~/.ssh/controlmasters/{{.Name}}-%r@%h:%p
	ControlMaster auto
	ControlPersist yes
	CheckHostIP=no
	StrictHostKeyChecking no
{{.Footer}}
`

const header = "#-- gocloud %s --"
const footer = "#---"

func makeFieldValues(name, ip string) *fieldValues {
	return &fieldValues{
		Name:   name,
		IP:     ip,
		Header: fmt.Sprintf(header, name),
		Footer: footer,
	}
}

// AddSshAlias adds a block to the user's ssh configuration file that
// provides an ssh alias to (typically of a created GCP node) ip
// (address).
func AddSshAlias(name, ip string) error {
	u, err := user.Current()
	if err != nil {
		return fmt.Errorf("no user, can't update ~/.ssh/config: %v", err)
	}
	h := u.HomeDir

	p := filepath.Join(u.HomeDir, ".ssh", "controlmasters")
	if err := os.MkdirAll(p, 0700); err != nil {
		return fmt.Errorf("can't make %q: %v", p, err)
	}
	p = filepath.Join(h, ".ssh", "config")

	return insertNameBlock(p, makeFieldValues(name, ip))
}

// insertNameBlock updates sshfile (which needs to be an ssh config file)
// with a machine configuration block specified by fields.
func insertNameBlock(sshfile string, fields *fieldValues) error {
	// TODO(rjk): need to error check templates before letting them be configurable.
	var t = template.Must(template.New("sshblock").Parse(machineblock))

	filebuffer, err := ioutil.ReadFile(sshfile)
	if err != nil {
		// No sshfile is not an error.
		filebuffer = []byte{}
	}

	tmpfilename := sshfile + ".tmp"
	tfd, err := os.OpenFile(tmpfilename, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("can't create tmp %q: %v", tmpfilename, err)
	}
	defer tfd.Close()
	defer os.Remove(tmpfilename)
	fd := bufio.NewWriter(tfd)

	pattern := "(?s)" + "\n?" + fields.Header + ".*?" + fields.Footer + "\n?"
	// log.Printf("complete regexp %q", pattern )
	re := regexp.MustCompile(pattern)
	locs := re.FindIndex(filebuffer)
	// log.Println("locs", locs)

	if locs == nil {
		if _, err := fd.Write(filebuffer); err != nil {
			return fmt.Errorf("can't write tmp %q: %v", tmpfilename, err)
		}
		if err := t.Execute(fd, fields); err != nil {
			return fmt.Errorf("can't expand machineblock: %v", err)
		}
	} else {
		if _, err := fd.Write(filebuffer[0:locs[0]]); err != nil {
			return fmt.Errorf("can't write tmp %q: %v", tmpfilename, err)
		}
		if err := t.Execute(fd, fields); err != nil {
			return fmt.Errorf("can't expand machineblock: %v", err)
		}
		if _, err := fd.Write(filebuffer[locs[1]:]); err != nil {
			return fmt.Errorf("can't write tmp %q: %v", tmpfilename, err)
		}
	}

	if err := fd.Flush(); err != nil {
		return fmt.Errorf("can't flush tmp %q: %v", tmpfilename, err)
	}

	return SafeReplaceFile(tmpfilename, sshfile)
}

// Copied from wikitools
func SafeReplaceFile(newpath, oldpath string) error {
	backup := oldpath + ".back"

	if _, err := os.Stat(oldpath); err == nil {
		if err := os.Link(oldpath, backup); err != nil {
			return fmt.Errorf("replaceFile backup: %v", err)
		}

		if err := os.Remove(oldpath); err != nil {
			return fmt.Errorf("replaceFile remove: %v", err)
		}
	}

	if err := os.Link(newpath, oldpath); err != nil {
		return fmt.Errorf("replaceFile emplace: %v", err)
	}

	if err := os.Remove(newpath); err != nil {
		return fmt.Errorf("replaceFile remove: %v", err)
	}
	if err := os.RemoveAll(backup); err != nil {
		return fmt.Errorf("replaceFile remove: %v", err)
	}
	return nil
}
