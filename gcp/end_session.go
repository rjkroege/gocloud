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

	"cloud.google.com/go/compute/metadata"
	"github.com/rjkroege/gocloud/config"
	compute "google.golang.org/api/compute/v1"
)


func EndSession(settings *config.Settings, instance string) error {
	_, client, err := NewAuthenticatedClient([]string{
		compute.ComputeScope,
	})
	if err != nil {
		return fmt.Errorf("NewAuthenticatedClient failed: %v", err)
	}

	projectid := settings.ProjectId
	zone := settings.Zone

	if metadata.OnGCE() {
		projectid, err = metadata.ProjectID()
		if err != nil {
			return fmt.Errorf("couldn't fetch the projectid because %v", err)
		}

		zone, err = metadata.Zone()
		if err != nil {
			return fmt.Errorf("couldn't fetch the zone because %v", err)
		}

		instance, err = metadata.InstanceName()
		if err != nil {
			return fmt.Errorf("couldn't fetch the instance because %v", err)
		}
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
