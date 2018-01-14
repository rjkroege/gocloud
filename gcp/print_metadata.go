package gcp

import (
	"log"
	"net/http"

	"cloud.google.com/go/compute/metadata"
	"github.com/rjkroege/sessionender/harness"
)

type printmetaCmd struct {
	listInstanceCmd
}


func init() {
	harness. AddSubCommand(&printmetaCmd{listInstanceCmd{
		"printmeta", 
		"",
		"printmeta displays a sampling of metadata on an instance",
	}})
}

func (c *printmetaCmd) Execute(client *http.Client, argv []string) {

	log.Println("printmeta trying to do something")

	if !metadata.OnGCE() {
		log.Fatalf("printmeta only works on GCE instances") 
	}
	
	pid, err :=  metadata.ProjectID()
	if err != nil {
		log.Println("couldn't fetch the projectid because", err)
	} else {
		log.Println(pid)
	}
	zone, err :=  metadata.Zone()
	if err != nil {
		log.Println("couldn't fetch the zone because", err)
	} else {
		log.Println(zone)
	}

	name, err :=  metadata.InstanceName()
	if err != nil {
		log.Println("couldn't fetch the name because", err)
	} else {
		log.Println(name)
	}
	
}