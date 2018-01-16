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
	"net/http"
	"strings"

	compute "google.golang.org/api/compute/v1"
	"cloud.google.com/go/compute/metadata"
	"github.com/rjkroege/sessionender/harness"
)

type endSessionCmd struct {
	listInstanceCmd
}


func init() {
	// Associates endsessionMain with the main.
	harness. AddSubCommand(MakeEndSession())
}

func MakeEndSession() harness.Command {
	scopes := strings.Join([]string{
		compute.ComputeScope,
	}, " ")

	return &endSessionCmd{listInstanceCmd{"endsession", scopes, "endsession [projectId zone instance]"}}
}

func (c *endSessionCmd) Execute(client *http.Client, argv []string) error {
	var projectid, zone, instance string

	if metadata.OnGCE() {
		argi := 0
		var err error

		projectid, err =  metadata.ProjectID()
		if err != nil {
			return fmt.Errorf("couldn't fetch the projectid because %v", err)

			if len(argv) > argi {
				projectid = argv[argi]
				argi++
			} else {
				return fmt.Errorf("no projectid from argument or metadata")
			}
		}

		zone, err =  metadata.Zone()
		if err != nil {
			return fmt.Errorf("couldn't fetch the zone because", err)

			if len(argv) > argi {
				zone = argv[argi]
				argi++
			} else {
				return fmt.Errorf("no zone from argument or metadata")
			}
		}
		
		instance, err =  metadata.InstanceName()
		if err != nil {
			return fmt.Errorf("couldn't fetch the instance because", err)

			if len(argv) > argi {
				instance = argv[argi]
				argi++
			} else {
				return fmt.Errorf("no instance from argument or metadata")
			}
		}
	} else {
		if len(argv) != 3 {
			return fmt.Errorf("Invalid argument list. Usage %s", c.Usage())
		}

		projectid = argv[0]
		zone = argv[1]
		instance = argv[2]
	}

	service, err := compute.New(client)
	if err != nil {
		return fmt.Errorf("Unable to create Compute service: %v", err)
	}

	_, err = service.Instances.Delete(projectid, zone, instance).Do()
	if err != nil { 
		return fmt.Errorf("Failed to delete instance %s because %v", instance, err)
	} 
	return nil
}


