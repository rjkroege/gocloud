package main

import (
	"log"
	"net/http"
	"strings"
	

	"cloud.google.com/go/compute/metadata"
)

func init() {
	scopes := strings.Join([]string{
//		compute.ComputeScope,
	}, " ")

	// Associates computeMain with the compute word.
	//  Could just as well be sessionender for example. Or something
	//  else.
	registerDemo("printmeta", scopes, printmetadataMain)
}

func printmetadataMain(client *http.Client, argv []string) {

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