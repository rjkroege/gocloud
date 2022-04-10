package gcp

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/rjkroege/gocloud/config"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

func parseDiskSize(szs string) (int64, error) {
	if szs == "" {
		return 0, nil
	}
	return strconv.ParseInt(szs, 10, 64)
}

// NodeInfo holds all the state necessary for subsequent utilities to be
// able to connect to the node.
type NodeInfo struct {
	Name  string
	ConfigName string
	Addr  string
	Token string
}

// Ssh returns the address for an SSH connection to the node.
func (ni *NodeInfo) Ssh() string {
	return ni.Addr + ":22"
}

// based on https://github.com/googleapis/google-api-go-client/blob/master/examples/compute.go

func MakeNode(settings *config.Settings, configName, instanceName string) (*NodeInfo, error) {
	ctx, client, err := NewAuthenticatedClient([]string{
		compute.ComputeScope,
	})
	if err != nil {
		return nil, fmt.Errorf("NewAuthenticatedClient failed: %v", err)
	}

	familyName := settings.InstanceTypes[configName].Family
	latestimage, err := findNewestStableImage(ctx, client, familyName)
	if err != nil {
		fmt.Println("can't find desired stable image", err)
	}

	// TODO(rjk): reuse the service.
	service, err := compute.New(client)
	if err != nil {
		return nil, fmt.Errorf("unable to create Compute service: %v", err)
	}

	projectID := settings.ProjectId
	zone := settings.Zone(configName)
	prefix := "https://www.googleapis.com/compute/v1/projects/" + projectID
	imageURL := "https://www.googleapis.com/compute/v1/projects/" + familyName + "/global/images/" + latestimage.Name

	machinetype := settings.InstanceTypes[configName].Hardware

	disksize, err := parseDiskSize(settings.InstanceTypes[configName].DiskSize)
	if err != nil {
		return nil, fmt.Errorf("unable to create Compute service, bad disk size: %q %v", settings.InstanceTypes[configName].DiskSize, err)
	}
	// TODO(rjk): add disk type (e.g. flash, persistent, etc.)

	metadata, err := makeMetadataObject(settings, configName)
	if err != nil {
		return nil, fmt.Errorf("can't make metadata: %v", err)
	}

	instance := &compute.Instance{
		Name:        instanceName,
		Description: settings.Description(configName, instanceName),
		MachineType: prefix + "/zones/" + zone + "/machineTypes/" + machinetype,

		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				Type:       "PERSISTENT",
				InitializeParams: &compute.AttachedDiskInitializeParams{
					// TODO(rjk): compute something better
					DiskName:    "ween-root",
					DiskSizeGb:  disksize,
					SourceImage: imageURL,
				},
			},
		},
		Metadata: convertMapToGcpFormat(metadata),
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
					compute.CloudPlatformScope,
				},
			},
		},
	}

	// TODO(rjk): check somehow that I've not attempted to remake a node.
	// Aside: wouldn't trying to make a node more than once fail?

	op, err := service.Instances.Insert(projectID, zone, instance).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("instance insertion failed: %v", err)
	}

	etag := op.Header.Get("Etag")
	log.Printf("Etag=%#v", etag)

	for i := 0; i < 12; i++ {
		// Wait a bit for the GCP to have done something.
		delayms := time.Duration(64*(1<<i)) * time.Millisecond
		log.Printf("wating %v...", delayms)
		delay := time.NewTimer(delayms)
		<-delay.C

		log.Println("polling for the instance running as desired")
		// TODO(rjk): Could use Fields() here?
		inst, err := service.Instances.Get(projectID, zone, instanceName).Context(ctx).IfNoneMatch(etag).Do()
		if err != nil && !googleapi.IsNotModified(err) {
			// Something went wrong and we should stop trying
			return nil, fmt.Errorf("getting inserted instance %s failed: %v", instanceName, err)
		}

		if err == nil {
			log.Printf("got %q, status %q", inst.Name, inst.Status)
			// Instance has changed but are we in the right state?
			ip, err := getExternalIP(inst)
			if err == nil && inst.Status == "RUNNING" {
				// Yes, it's running and has an IP.
				return &NodeInfo{
					Name:  inst.Name,
					ConfigName: configName,
					Addr:  ip,
					Token: metadata["instancetoken"],
				}, nil
			}
			// Not in the right state yet. Try again.
			etag = inst.Header.Get("Etag")
		}
	}
	return nil, fmt.Errorf("too many tries failing to get running state for %s", instanceName)
}

// getExternalIP digs through inst looking for its external (i.e. via NAT) IP
func getExternalIP(inst *compute.Instance) (string, error) {
	// I don't know how much variety that there would be in the structure of the info
	// I want one external IP. Not necessarily all of them.
	for _, ni := range inst.NetworkInterfaces {
		for _, ac := range ni.AccessConfigs {
			if ip := ac.NatIP; ip != "" {
				return ip, nil
			}
		}
	}
	return "", fmt.Errorf("%s doesn't have external ip", inst.Name)
}
