package gcp

import (
	"fmt"
	"log"

	"github.com/rjkroege/gocloud/config"
	compute "google.golang.org/api/compute/v1"
	"github.com/sanity-io/litter"
	"google.golang.org/api/googleapi"
)

// based on https://github.com/googleapis/google-api-go-client/blob/master/examples/compute.go


func MakeNode(settings *config.Settings) error {
	_, client, err := NewAuthenticatedClient([]string{
		compute.ComputeScope,
	})
	if err != nil {
		return fmt.Errorf("NewAuthenticatedClient failed: %v", err)
	}

	// TODO(rjk): Pass in the context.
	// TODO(rjk): the family name needs to come from settings.
	latestimage, err := findNewestStableCosImage(client)
	if err != nil {
		fmt.Println("can't find desired stable image", err)
	}


	// TODO(rjk): reuse the service.
	service, err := compute.New(client)
	if err != nil {
		return fmt.Errorf("unable to create Compute service: %v", err)
	}

	projectID := settings.ProjectId
	zone := settings.Zone
	prefix := "https://www.googleapis.com/compute/v1/projects/" + projectID
	imageURL := "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/" + latestimage.Name

	// TODO(rjk): the machine name needs to come from an argument / settings
	instanceName := "ween"

	// TODO(rjk): the machine type needs to come from the settings file.
	machinetype := "e2-small"
	// TODO(rjk): the disk configuration needs to come from the settings.

	instance := &compute.Instance{
		Name:        instanceName,
		Description: "ween instance",
		MachineType: prefix + "/zones/" + zone + "/machineTypes/" + machinetype,

		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				Type:       "PERSISTENT",
				InitializeParams: &compute.AttachedDiskInitializeParams{
					// TODO(rjk): compute something better
					DiskName:    "ween-root",
					SourceImage: imageURL,
				},
			},
		},
//		Metadata: metadata,
		NetworkInterfaces: []*compute.NetworkInterface{
			{
				AccessConfigs: []*compute.AccessConfig{
					{
						Type: "ONE_TO_ONE_NAT",
						Name: "External NAT",
					},
				},
				Network: prefix + "/global/networks/default",
			},
		},
		ServiceAccounts: []*compute.ServiceAccount{
			{
				// TODO(rjk): read this from the config file.
				Email: "default",
				Scopes: []string{
					// TODO(rjk): I have no idea if this will do what I want.
					compute.DevstorageFullControlScope,
					compute.ComputeScope,
				},
			},
		},
	}

	// TODO(rjk): check somehow that I've not attempted to remake a node.
	// Aside: wouldn't trying to make a node more than once fail?

	// TODO(rjk): shouldn't I check the status here.
	op, err := service.Instances.Insert(projectID, zone, instance).Do()
	if err != nil {
		return fmt.Errorf("instance insertion failed: %v", err)
	}

	// TODO(rjk): verbosity controlled by settings
	litter.Dump(op)

	// TODO(rjk): Why did the original example do this? Wouldn't the "right" thing be to
	// retry the Insert a few times?
	etag := op.Header.Get("Etag")
	log.Printf("Etag=%v", etag)

	inst, err := service.Instances.Get(projectID, zone, instanceName).IfNoneMatch(etag).Do()
	if err != nil {
		return fmt.Errorf("instance Get failed: %v", err)
	}
	litter.Dump(inst)
	if googleapi.IsNotModified(err) {
		log.Printf("Instance not modified since insert.")
	} else {
		log.Printf("Instance modified since insert.")
	}

// TODO(rjk): Need to update the .ssh/config to let me ssh to the node.
// TODO(rjk): Need a flag to turn that off probably.
	fmt.Printf("hostname is either %s.c.%s.internal or %s.%s.c.%s.internal\n", instanceName, projectID, instanceName, zone, projectID)

	return nil
}

// TODO(rjk): Generate metadata in a rational way.
func makeMetadataObject() (*compute.Metadata, error) {
	// gitcredential (read from the keychain)
	// username
	// user-data (from ween)
	// How to handle use-data? Compile in? Read from an external file? The config
	// builder constructs it right? Maybe read from a file.
	// ssh-key 

	return nil, nil
}
