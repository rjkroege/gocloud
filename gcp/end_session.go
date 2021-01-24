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
	// TODO(rjk): Doesn't handle being outside of the default zone.
	// TODO(rjk): zone handling needs to be addressed better.
	zone := settings.DefaultZone

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
