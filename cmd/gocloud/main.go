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
	} `cmd help:"Make node."`

	Del struct {
		Node string `arg name:"node" help:"Node to remove."`
	} `cmd help:"Delete node."`

	Ls struct {
	} `cmd help:"List running nodes."`

	LsImages struct {
	} `cmd help:"List available images."`
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
		log.Println(CLI.ConfigFile, settings)
		if err := gcp.ListImages(settings); err != nil {
			fmt.Println("can't list images:", err)
			os.Exit(-1)
		}
	case "make":
		log.Println("run make")
		if err := gcp.MakeNode(settings); err != nil {
			fmt.Println("can't make node:", err)
			os.Exit(-1)
		}
	case "del <node>":
		log.Println("run del")
		log.Println(CLI.Del.Node)
		if err := gcp.EndSession(settings, CLI.Del.Node); err != nil {
			fmt.Println("can't list nodes:", err)
			os.Exit(-1)
		}
	default:
		panic(ctx.Command())
	}
}
