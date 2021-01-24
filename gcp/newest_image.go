package gcp

import (
	"context"
	"net/http"
	"fmt"
	"strings"
	"strconv"

	"github.com/rjkroege/gocloud/config"
	compute "google.golang.org/api/compute/v1"
)

// TODO(rjk): Rationalize project name arguments.
func ListImages(settings *config.Settings) error {
	// is there a way to determine the family 
	ctx, client, err := NewAuthenticatedClient([]string{
		compute.ComputeScope,
	})
	if err != nil {
		return fmt.Errorf("NewAuthenticatedClient failed: %v", err)
	}

	notdeprecated, err := listProjectImages(ctx, client,"cos-cloud")
	if err != nil {
		return fmt.Errorf("listProjectImages: %v", err)
	}
	// Output the not-deprecated.
	for _, im := range notdeprecated {
		// TODO(rjk): have a verbose setting to dump more info?
		fmt.Println(im.Name, im.Family)
	}

	neweststable, err := findNewestStableCosImage(ctx, client)
	if err != nil {
		fmt.Println("can't find stable image", err)
	}
	fmt.Println(">> picked image", neweststable.Name)

	return nil
}

type VersionTuple [4]int

func parseCosName(name string) (string, VersionTuple, error) {
	vt := VersionTuple{0,0,0,0}

	ps := strings.Split(name, "-")

	if ps[0] != "cos" {
		return "", vt, fmt.Errorf("can't parse %q", name)
	}
	channel := "lts"
	switch ps[1] {
	case "stable", "beta", "dev":
		channel  = ps[1]
		ps = ps[2:]
	default:
		ps = ps[1:]
	}

	for i, s := range ps {
		v, err := strconv.Atoi(s)
		if err != nil {
			return "",  [4]int{0,0,0,0}, fmt.Errorf("[%d] can't parse int from %q", i, name)
		}
		vt[i] = v
	}
	return channel, vt, nil
}

// TODO(rjk): Make this into a function that can be used as a sort comparison
// function.
// I think that I didn't need this code. Whatever.
func newest(v1, v2 VersionTuple) VersionTuple {
	for i , _ := range v1 {
		if v1[i] > v2[i] {
			return v1
		} else {
			return v2
		}
	}
	// I think that this means that they're equal
	return v1
}

func findNewestStableCosImage(ctx context.Context, client *http.Client) (*compute.Image, error) {
	notdeprecated, err := listProjectImages(ctx, client, "cos-cloud")
	if err != nil {
		return nil, fmt.Errorf("listProjectImages: %v", err)
	}

	for _, im := range notdeprecated {
		c, _, err := parseCosName(im.Name)
		if err != nil {
			return nil, fmt.Errorf("can't parse: %v", err)
		}		
		if c == "stable" {
			return im, nil
		}
		// TODO(rjk): Use newest() to find an lts image as necessary.
	}
	return nil, fmt.Errorf("no stable cos-cloud image")	
}

func listProjectImages(ctx context.Context, client *http.Client, project string) ([]*compute.Image, error) {
	service, err := compute.New(client)
	if err != nil {
		return nil, fmt.Errorf("Unable to create Compute service: %v", err)
	}

	// Show the current images that are available.
	listcommand := service.Images.List(project).Context(ctx)

	// TODO(rjk): Use a filter operation?
	notdeprecated := make([]*compute.Image, 0)
	for {
		// TODO(rjk): should be able to use filters...
		res, err := listcommand.Do()
		if err != nil {
			return nil, fmt.Errorf("can't list images: %v", err)
		}

		for _, im := range res.Items {
			if im.Deprecated == nil {
				notdeprecated = append(notdeprecated, im)
			}
		}

		if res.NextPageToken == "" {
			break
		}
		listcommand = listcommand.PageToken(res.NextPageToken)
	}
	return notdeprecated, nil
}
