package gcp

import (
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/compute/metadata"
	"github.com/rjkroege/gocloud/harness"
)

type printmetaCmd struct {
	listInstanceCmd
}

func init() {
	harness.AddSubCommand(&printmetaCmd{listInstanceCmd{
		"printmeta",
		"",
		"printmeta displays a sampling of metadata on an instance",
	}})
}

func (c *printmetaCmd) Execute(client *http.Client, argv []string) error {

	if !metadata.OnGCE() {
		log.Fatalf("printmeta only works on GCE instances")
	}

	pid, err := metadata.ProjectID()
	if err != nil {
		return fmt.Errorf("couldn't fetch the projectid because %v", err)
	} else {
		log.Println(pid)
	}
	zone, err := metadata.Zone()
	if err != nil {
		return fmt.Errorf("couldn't fetch the zone because %v", err)
	} else {
		log.Println(zone)
	}

	name, err := metadata.InstanceName()
	if err != nil {
		return fmt.Errorf("couldn't fetch the name because %v", err)
	} else {
		log.Println(name)
	}
	return nil
}
