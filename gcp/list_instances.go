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

package gcp

import (
	"log"
	"net/http"
	"strings"

	compute "google.golang.org/api/compute/v1"
	"github.com/rjkroege/sessionender/harness"
)

type listInstanceCmd struct {
	name string	
	scopes string	
	usage string
}


func (c *listInstanceCmd) Scope() string {
	return c.scopes
}
func (c *listInstanceCmd) Name() string {
	return c.name
}
func (c *listInstanceCmd) Usage() string {
	return c.usage
}

func init() {
	scopes := strings.Join([]string{
		compute.ComputeScope,
	}, " ")

	harness. AddSubCommand(&listInstanceCmd{"list", scopes, "list project_id zone"})
}

func (c *listInstanceCmd) Execute(client *http.Client, argv []string) {
	if len(argv) != 2 {
		log.Println("bad number of args", c.Usage())
		return
	}

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

