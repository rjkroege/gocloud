package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

// user holds a single cloudconfig user entry.
type user struct {
	Name string
	Uid  int
}

// fileentry holds a single file to write.
type fileentry struct {
	Path string
	// TODO(rjk): does it parse this a number? It ends up as one for the
	// system call.
	Permissions int
	Owner       string
	Content     string
	skip        bool
}

type cloudconfig struct {
	Users       []user
	Write_files []fileentry
	Runcmd      []string
}

// readuserdata reads per-user configuration from each service directory.
func readuserdata(dirs []string) []user {
	userslist := make([]user, 0, len(dirs))
	for _, d := range dirs {
		fi, err := os.Stat(d)
		if err != nil {
			log.Fatalf("Giving up. No file/dir %s because %v", d, err)
		}

		if !fi.IsDir() {
			continue
		}

		fd, err := os.Open(filepath.Join(d, "user.yaml"))
		if err != nil {
			log.Fatalf("Giving up. Can't open %s/user.yaml because %v", d, err)
		}

		userraw, err := ioutil.ReadAll(fd)
		if err != nil {
			log.Fatalf("Giving up. Can't read %s/user.yaml because %v ", d, err)
		}
		fd.Close()

		var u user
		if err := yaml.Unmarshal(userraw, &u); err != nil {
			log.Fatalf("Giving up. Can't decode %s/user.yaml because %v", d, err)
			continue
		}

		if u.Name == "" {
			u.Name = filepath.Base(d)
		}
		userslist = append(userslist, u)
	}
	return userslist
}

// processRegularFile handles supplementary file writes. The innards of
// the file are broken out in this scheme so that I don't have to escape
// things.
func processRegularFile(rfn string) (fileentry, error) {
	content, err := ioutil.ReadFile(rfn)
	if err != nil {
		return fileentry{}, fmt.Errorf("processRegularFile can't read %s: %v", rfn, err)
	}

	metadataname := rfn + ".meta"
	mdf, err := ioutil.ReadFile(metadataname)
	if err != nil {
		return fileentry{}, fmt.Errorf("processRegularFile can't read %s: %v", metadataname, err)
	}
	var fi fileentry
	if err := yaml.Unmarshal(mdf, &fi); err != nil {
		return fileentry{}, fmt.Errorf("processRegularFile can't decode %s: %v", metadataname, err)
	}

	fi.Content = string(content)
	fi.skip = true
	return fi, nil
}

func readservicedefn(dirs []string) []fileentry {
	servicefiles := make([]fileentry, 0, len(dirs))
	for _, d := range dirs {
		fi, err := os.Stat(d)
		if err != nil {
			log.Fatalf("Giving up. No file/dir %s because %v", d, err)
		}

		if !fi.IsDir() {
			fe, err := processRegularFile(d)
			if err != nil {
				log.Fatalf("Giving up. Regular file %s failed because %v", d, err)
			} else {
				servicefiles = append(servicefiles, fe)
			}
			continue
		}

		sfiles, err := filepath.Glob(filepath.Join(d, "*.service"))
		if err != nil {
			log.Fatalf("Giving up. No service files in %d because %v", d, err)
		}

		for _, fn := range sfiles {
			fd, err := os.Open(fn)
			if err != nil {
				log.Fatalf("Giving up. Can't open %s %v", fn, err)
			}

			svcfile, err := ioutil.ReadAll(fd)
			if err != nil {
				log.Fatalf("Giving up. Can't read %s because %v", fn, err)
			}
			fd.Close()

			servicefiles = append(servicefiles, fileentry{
				Path:        filepath.Join("/etc/systemd/system", filepath.Base(fn)),
				Permissions: 0644,
				Owner:       "root",
				Content:     string(svcfile),
			})

		}
	}

	return servicefiles
}

// mksvccmds assembles the systemctl start commands for the
// configured services.
func mksvccmds(svcs []fileentry) []string {
	cmds := make([]string, 0, len(svcs))
	cmds = append(cmds, "systemctl daemon-reload")

	for _, svc := range svcs {
		if !svc.skip {
			cmds = append(cmds, "systemctl start "+filepath.Base(svc.Path))
		}
	}

	return cmds
}

// loadpreamblecmds loads the yaml file fn. This file should contain
// a list of commands in yaml format.
func loadpreamblecmds(fn string) []string {
	if fn == "" {
		return []string{}
	}

	fd, err := os.Open(fn)
	if err != nil {
		log.Fatalf("Invalid preabmle %s: %v", fn, err)
	}

	pcmdsraw, err := ioutil.ReadAll(fd)
	if err != nil {
		log.Fatalf("Invalid preabmle %s: %v", fn, err)
	}
	fd.Close()

	var cmds []string
	if err := yaml.Unmarshal(pcmdsraw, &cmds); err != nil {
		log.Fatalf("Invalid preabmle %s: %v", fn, err)
	}
	return cmds
}

var pcmdfn = flag.String("preamblecmds", "", "Cmds to be added as a preamble to the service starts")

const helpmessage = `makdcloudconfig assembles directories containing services together
into a single cloudconfig user-data payload for the gcloud compute
instances create's --metadata-from-file's option. Use the result of
this command to set the value of the user-data key in the metadata.

Each argument directory should contain 1 or more service files with a
.service suffix. Also each directory may contain a user.yaml file is
present to specify users to create in the cloudconfig. `

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n%s\n", os.Args[0], helpmessage)
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	dirs := flag.Args()
	if len(dirs) < 1 {
		log.Fatalf("Can't proceed without at least one service directory")
	}

	// Each directory is expected to contain a service definition (a file ending
	// with the service suffix) and a user.yaml file that specifies the uid and user
	// name for this service.
	//
	// If any errors occur, the cloudconfig file is probably not valid.

	fes := readservicedefn(dirs)
	config := cloudconfig{
		Users:       readuserdata(dirs),
		Write_files: fes,
		Runcmd:      append(loadpreamblecmds(*pcmdfn), mksvccmds(fes)...),
	}

	d, err := yaml.Marshal(config)
	if err != nil {
		log.Fatalf("can't make yaml from data error: %v", err)
	}

	// The presence of this string at the top is required for the cloudconfig
	// tooling to correctly parse the file.
	if _, err := os.Stdout.Write([]byte("#cloud-config\n")); err != nil {
		log.Fatalf("can't emit final result error: %v", err)
	}
	if _, err := os.Stdout.Write(d); err != nil {
		log.Fatalf("can't emit final result error: %v", err)
	}
}
