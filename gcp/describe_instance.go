package gcp

import (
	"fmt"

	"github.com/rjkroege/gocloud/config"
	"github.com/sanity-io/litter"
	compute "google.golang.org/api/compute/v1"
)

// DescribeInstance describes instance |name| with a sump of its JSON
// description.
func DescribeInstance(settings *config.Settings, name string) error {
	instance, err := getInstance(settings, name)
	if err != nil {
		return err
	}

	litter.Dump(instance)
	//	litter.Dump(pullKeys(instance))

	return nil
}

func getInstance(settings *config.Settings, name string) (*compute.Instance, error) {
	_, client, err := NewAuthenticatedClient([]string{
		compute.ComputeScope,
	})
	if err != nil {
		return nil, fmt.Errorf("NewAuthenticatedClient failed: %v", err)
	}

	service, err := compute.New(client)
	if err != nil {
		return nil, fmt.Errorf("Unable to create Compute service: %v", err)
	}

	projectId := settings.ProjectId
	zone := settings.DefaultZone

	// It is impossible to filter to a single metadata element. I can however
	// shrink the total response size by just asking for the metadata key
	// strings. I find this limitation perplexing. This is what the Fields() method
	// does. Accept it.
	//	instance, err := service.Instances.Get(projectId, zone, name).Fields("metadata/items/key").Do()
	instance, err := service.Instances.Get(projectId, zone, name).Do()
	if err != nil {
		return nil, fmt.Errorf("getting instance failed: %v", err)
	}
	return instance, nil
}

func pullKeys(instance *compute.Instance) []string {
	keys := make([]string, 0, len(instance.Metadata.Items))
	for _, kn := range instance.Metadata.Items {
		keys = append(keys, kn.Key)
	}
	return keys
}

func GetMetadataKeys(settings *config.Settings, name string) ([]string, error) {
	instance, err := getInstance(settings, name)
	if err != nil {
		return []string{}, err
	}

	return pullKeys(instance), nil
}
