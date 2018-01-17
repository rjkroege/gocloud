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
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/rjkroege/gocloud/harness"
	compute "google.golang.org/api/compute/v1"
)

type listInstanceCmd struct {
	name   string
	scopes string
	usage  string
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

	harness.AddSubCommand(&listInstanceCmd{"list", scopes, "list project_id zone"})
}

func (c *listInstanceCmd) Execute(client *http.Client, argv []string) error {
	if len(argv) != 2 {
		return fmt.Errorf("Wrong number of args. Usage: %s", c.Usage())
	}

	service, err := compute.New(client)
	if err != nil {
		return fmt.Errorf("Unable to create Compute service: %v", err)
	}

	projectId := argv[0]
	zone := argv[1]

	// List the current instances.
	res, err := service.Instances.List(projectId, zone).Do()
	if err != nil {
		return fmt.Errorf("Getting instance list failed: %v", err)
	}

	for _, inst := range res.Items {
		log.Println(inst.Name, inst.MachineType)
	}
	return nil
}
