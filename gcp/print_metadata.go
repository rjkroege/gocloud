package gcp

import (
	"fmt"
	"log"

	"cloud.google.com/go/compute/metadata"
)


func PrintMetadata() error {
	if !metadata.OnGCE() {
		log.Fatalf("printmeta only works on GCE instances")
	}

	pid, err := metadata.ProjectID()
	if err != nil {
		return fmt.Errorf("couldn't fetch the projectid because %v", err)
	}
		log.Println(pid)
	zone, err := metadata.Zone()
	if err != nil {
		return fmt.Errorf("couldn't fetch the zone because %v", err)
	} 
		log.Println(zone)

	name, err := metadata.InstanceName()
	if err != nil {
		return fmt.Errorf("couldn't fetch the name because %v", err)
	} 
		log.Println(name)
	return nil
}
