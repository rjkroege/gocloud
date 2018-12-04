package main

import (
	"log"
	"flag"
	"path/filepath"
	"fmt"
	"os"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"

)

// user holds a single cloudconfig user entry.
type user struct {
	Name string
	Uid int
}

// fileentry holds a single file to write.
type fileentry struct {
	Path string
	// TODO(rjk): does it parse this a number? It ends up as one for the
	// system call.
	Permissions int
	Owner string
	Content string
}

type cloudconfig struct {
	Users []user
	Write_files []fileentry
	Runcmd []string
}



// readuserdata reads per-user configuration from each service directory.
func readuserdata(dirs []string)  []user {
	userslist := make([]user, 0, len(dirs))
	for _, d := range dirs {
		fd, err := os.Open(filepath.Join(d, "user.yaml"))
		if err != nil {
			log.Printf("can't open %s/user.yaml because %v. Skipping: ", d, err)
			continue
		}

		userraw, err := ioutil.ReadAll(fd)
		if err != nil {
			log.Printf("can't read %s/user.yaml because %v. Skipping: ", d, err)	
			fd.Close()
			continue
			
		}
		fd.Close()

		var u user
		if err := yaml.Unmarshal(userraw, &u); err != nil {
			log.Printf("can't decode %s/user.yaml because %v. Skipping: ", d, err)	
			continue
		}
	
		if u.Name == "" {
			u.Name = 	filepath.Base(d)
		}
		userslist = append(userslist, u)
	}
	return userslist
}

func readservicedefn(dirs []string) []fileentry {
	servicefiles := make([]fileentry, 0, len(dirs))
	for _, d := range dirs {
		// TODO(rjk): I only support service files. Should I
		// consider supporting more kinds of files? Multiple service files are possible
		sfiles, err := filepath.Glob(filepath.Join(d, "*.service" ))
		if err != nil {
			log.Printf("can't find service files in %d because %v. Skipping", d, err)
			continue
		}

		log.Println("possible services", sfiles)
		
		for _, fn := range sfiles {
			fd, err := os.Open(fn)
			if err != nil {
				log.Printf("can't open %s %v. Skipping: ", fn, err)
				continue
			}
		
			svcfile, err := ioutil.ReadAll(fd)
			if err != nil {
				log.Printf("can't read %s because %v. Skipping: ", fn, err)	
				fd.Close()
				continue
			}
			fd.Close()
			
			servicefiles = append(servicefiles, fileentry{
				Path: filepath.Join("/etc/systemd/system", filepath.Base(fn)),
				Permissions: 0644,
				Owner: "root",
				Content: string(svcfile),
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
		cmds = append(cmds, "systemctl start " +  filepath.Base(svc.Path))
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
				log.Printf("can't open %s %v. Skipping: ", fn, err)
				return []string{}
			}
		
			pcmdsraw, err := ioutil.ReadAll(fd)
			if err != nil {
				log.Printf("can't read %s because %v. Skipping: ", fn, err)	
				fd.Close()
				return []string{}
			}
			fd.Close()

		var cmds []string
		if err := yaml.Unmarshal(pcmdsraw, &cmds); err != nil {
			log.Printf("can't decode %s because %v. Skipping: ", fn, err)	
			return []string{}
		}
return 	cmds		
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
		log.Println("Can't proceed without at least one service directory")
		return
	}

	// Each directory is expected to contain a service definition (a file ending
	// with the service suffix) and a user.yaml file that specifies the uid and user
	// name for this service.

	log.Println("dirs:", dirs)
	
	// If any errors occur, the cloudconfig file is probably not valid.

	fes := readservicedefn(dirs)
	config := cloudconfig{
		Users: readuserdata(dirs),
		Write_files: fes,
		Runcmd:    append(loadpreamblecmds(*pcmdfn), mksvccmds(fes)...),
		
	}
	
	
        d, err := yaml.Marshal(config)
        if err != nil {
                log.Fatalf("error: %v", err)
        }
        fmt.Printf("--- m dump:\n%s\n\n", string(d))	
	
}
