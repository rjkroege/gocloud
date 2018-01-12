// Copyright 2017 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"log"
	"net/http"
	"strings"

	compute "google.golang.org/api/compute/v1"
	"cloud.google.com/go/compute/metadata"
)

func init() {
	scopes := strings.Join([]string{
		compute.ComputeScope,
	}, " ")

	// Associates endsessionMain with the main.
	registerDemo("endsession", scopes, endsessionMain)
}

func endsessionMain(client *http.Client, argv []string) {
	// This command is intended to be used on a single instance and
	// will cause the instance to shut itself down. Otherwise, it
	// will permit running a command to end a running instance.


	var projectid, zone, instance string

	if metadata.OnGCE() {
		argi := 0
		var err error

		projectid, err =  metadata.ProjectID()
		if err != nil {
			log.Println("couldn't fetch the projectid because", err)

			if len(argv) > argi {
				projectid = argv[argi]
				argi++
			} else {
				log.Fatalln("no projectid from argument or metadata")
			}
		}

		zone, err =  metadata.Zone()
		if err != nil {
			log.Println("couldn't fetch the zone because", err)

			if len(argv) > argi {
				zone = argv[argi]
				argi++
			} else {
				log.Fatalln("no zone from argument or metadata")
			}
		}

		
		instance, err =  metadata.InstanceName()
		if err != nil {
			log.Println("couldn't fetch the instance because", err)

			if len(argv) > argi {
				instance = argv[argi]
				argi++
			} else {
				log.Fatalln("no instance from argument or metadata")
			}
		}
	} else {
		if len(argv) != 3 {
			log.Fatalln("Usage: endsession project_id zone instance")
		}

		projectid = argv[0]
		zone = argv[1]
		instance = argv[2]
	}

	service, err := compute.New(client)
	if err != nil {
		log.Fatalf("Unable to create Compute service: %v", err)
	}


	log.Println("shutting down instance", projectid, zone, instance)

	_, err = service.Instances.Delete(projectid, zone, instance).Do()
	if err != nil { 
		log.Println("Failed to delete instance", instance)
		return
	} 
}

