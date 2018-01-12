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
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	compute "google.golang.org/api/compute/v1"
)

func init() {
	scopes := strings.Join([]string{
		compute.ComputeScope,
	}, " ")

	// Associates computeMain with the compute word.
	//  Could just as well be sessionender for example. Or something
	//  else.
	registerDemo("list", scopes, listInstancesMain)
}

func listInstancesMain(client *http.Client, argv []string) {
	if len(argv) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: list project_id zone")
		return
	}

	// TODO(rjk): default the zone. Even better, keep it in the cache file?

	service, err := compute.New(client)
	if err != nil {
		log.Fatalf("Unable to create Compute service: %v", err)
	}

	projectId := argv[0]
	zone := argv[1]

	// List the current instances.
	res, err := service.Instances.List(projectId, zone).Do()
	if err != nil {
		log.Println("Getting instance list failed:", err)
		return
	}

	log.Println("Got compute.Images.List")

	for _, inst := range res.Items {
		log.Println(inst.Name, inst.MachineType)
	}

}

