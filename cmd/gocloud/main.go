package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alecthomas/kong"
	"github.com/rjkroege/gocloud/config"
	"github.com/rjkroege/gocloud/gcp"

	// For debugging
	"github.com/sanity-io/litter"
)

var CLI struct {
	ConfigFile string `type:"path" help:"Set alternate configuration file" default:"~/.config/gocloud/gocloud.json"`

	Make struct {
		Config string `arg name:"config" help:"Defined configuration for instance"`
		Name   string `arg name:"name" help:"Name of instance"`
	} `cmd help:"Make instance."`

	Del struct {
		Node string `arg name:"node" help:"Node to remove."`
	} `cmd help:"Delete node." aliases:"rm"`

	Describe struct {
		Name string `arg name:"name" help:"Name of instance"`
	} `cmd help:"Describe a specific node"`

	Ls struct {
	} `cmd help:"List running nodes."`

	LsImages struct {
	} `cmd help:"List available images."`

	ShowMeta struct {
		Config string `arg name:"config" help:"Defined configuration for instance"`
	} `cmd help:"Show metadata for configuration"`
}

func main() {
	ctx := kong.Parse(&CLI)

	settings, err := config.Read(CLI.ConfigFile)
	if err != nil {
		fmt.Println("Fatai:", err)
		os.Exit(-1)
	}

	// temp to see if the toml thing is working

	litter.Dump(settings)

	return

	switch ctx.Command() {
	case "ls":
		log.Println("run ls")
		log.Println(CLI.ConfigFile, settings)
		if err := gcp.List(settings); err != nil {
			fmt.Println("can't list nodes:", err)
			os.Exit(-1)
		}
	case "ls-images":
		log.Println("run lsimages")
		log.Println("dumping configuration", CLI.ConfigFile, settings)
		if err := gcp.ListImages(settings); err != nil {
			fmt.Println("can't list images:", err)
			os.Exit(-1)
		}
	case "make <config> <name>":
		log.Println("run make")
		// TODO(rjk): There's probably some fancy Kong way to do this that's better.
		if _, ok := settings.InstanceTypes[CLI.Make.Config]; !ok {
			fmt.Printf("undefined instance type %q\n", CLI.Make.Config)
			os.Exit(-1)
		}

		ni, err := gcp.MakeNode(settings, CLI.Make.Config, CLI.Make.Name)
		if err != nil {
			fmt.Println("can't make node:", err)
			os.Exit(-1)
		}

		// Wait for the Ssh server to be running.
		client, err := gcp.WaitForSsh(settings, ni)
		if err != nil {
			fmt.Printf("no ssh ever came up: %v", err)
			os.Exit(-1)
		}
		defer client.Close()

		if err := gcp.ConfigureViaSsh(settings, ni, client); err != nil {
			fmt.Printf("ConfigureViaSsh failed: %v", err)
			// Should I exit here?
		}

		if err := config.AddSshAlias(ni.Name, ni.Addr); err != nil {
			fmt.Printf("can't update ssh for node %v: %v", ni, err)
		}
	case "del <node>":
		log.Println("run del")
		log.Println(CLI.Del.Node)
		if err := gcp.EndSession(settings, CLI.Del.Node); err != nil {
			fmt.Printf("can't remove instance %s: %v", CLI.Del.Node, err)
			os.Exit(-1)
		}
	case "show-meta <config>":
		log.Println("run ShowMetadata")
		if _, ok := settings.InstanceTypes[CLI.ShowMeta.Config]; !ok {
			fmt.Printf("undefined instance type %q\n", CLI.ShowMeta.Config)
			os.Exit(-1)
		}
		if err := gcp.ShowMetadata(settings, CLI.ShowMeta.Config); err != nil {
			fmt.Printf("can't show metadata for config %s: %v\n", CLI.ShowMeta.Config, err)
			os.Exit(-1)
		}
	case "describe <name>":
		log.Println("run DescribeInstance")
		log.Println(CLI.ConfigFile, settings)
		if err := gcp.DescribeInstance(settings, CLI.Describe.Name); err != nil {
			fmt.Printf("can't describe %s: %v\n", CLI.Describe.Name, err)
			os.Exit(-1)
		}
	default:
		panic(ctx.Command())
	}
}
