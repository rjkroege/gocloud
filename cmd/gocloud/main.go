package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rjkroege/gocloud/gcp"
	"github.com/alecthomas/kong"
	"github.com/rjkroege/gocloud/config"
)

var CLI struct {
	ConfigFile string `type:"path" help:"Set alternate configuration file" default:"~/.config/gocloud/gocloud.json"`

	Make struct {
		Config string `arg name:"config" help:"Defined configuration for instance"`
		Name string `arg name:"name" help:"Name of instance"`
	} `cmd help:"Make instance."`

	Del struct {
		Node string `arg name:"node" help:"Node to remove."`
	} `cmd help:"Delete node."`

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

		if err := gcp.MakeNode(settings, CLI.Make.Config, CLI.Make.Name); err != nil {
			fmt.Println("can't make node:", err)
			os.Exit(-1)
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
	default:
		panic(ctx.Command())
	}
}
