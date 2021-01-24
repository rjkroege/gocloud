package gcp

import (
	"fmt"

	"github.com/rjkroege/gocloud/config"
	compute "google.golang.org/api/compute/v1"
)

func List(settings *config.Settings) error {
	_, client, err := NewAuthenticatedClient([]string{
		compute.ComputeScope,
	})
	if err != nil {
		return fmt.Errorf("NewAuthenticatedClient failed: %v", err)
	}

	service, err := compute.New(client)
	if err != nil {
		return fmt.Errorf("Unable to create Compute service: %v", err)
	}

	projectId := settings.ProjectId
	// TODO(rjk): Support multiple zones correctly.
	zone := settings.DefaultZone

	// List the current instances.
	res, err := service.Instances.List(projectId, zone).Do()
	if err != nil {
		return fmt.Errorf("getting instance list failed: %v", err)
	}

	for _, inst := range res.Items {
		fmt.Println(inst.Name, inst.MachineType)
	}
	return nil
}
