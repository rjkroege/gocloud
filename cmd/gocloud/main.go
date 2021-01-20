// Derived from https://github.com/google/google-api-go-client/blob/master/examples/main.go
//
// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	//	"context"
	//	"flag"
	"fmt"
	"log"
	//	"net/http"
	"os"
	//	"path/filepath"

	"github.com/rjkroege/gocloud/gcp"
	//	"github.com/rjkroege/gocloud/harness"
	//	"golang.org/x/oauth2"
	//	"golang.org/x/oauth2/google"
	"github.com/alecthomas/kong"
	"github.com/rjkroege/gocloud/config"
)

var CLI struct {
	ConfigFile string `type:"path" help:"Set alternate configuration file" default:"~/.config/gocloud/gocloud.json"`

	Rm struct {
		Force     bool `help:"Force removal."`
		Recursive bool `help:"Recursively remove files."`

		Paths []string `arg name:"path" help:"Paths to remove." type:"path"`
	} `cmd help:"Remove files."`

	Make struct {
	} `cmd help:"Make node."`

	Del struct {
		Node string `arg name:"node" help:"Node to remove."`
	} `cmd help:"Delete node."`

	Ls struct {
	} `cmd help:"List running nodes."`
}

func main() {
	ctx := kong.Parse(&CLI)

	settings, err := config.Read(CLI.ConfigFile)
	if err != nil {
		fmt.Println("Fatai:", err)
		os.Exit(-1)
	}

	switch ctx.Command() {
	case "rm <path>":
		log.Println("run rm <path>")
	case "ls":
		log.Println("run ls")
		log.Println(CLI.ConfigFile, settings)
		if err := gcp.List(settings); err != nil {
			fmt.Println("can't list nodes:", err)
			os.Exit(-1)
		}
	case "make":
		log.Println("run make")
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

	/*
		ctx := context.Background()
		if *debug {
			ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
				Transport: gcp.NewTransport(http.DefaultTransport),
			})
		}
		client, err := google.DefaultClient(ctx, cmd.Scope())
		if err != nil {
			log.Fatalln("Can't setup an OAuth connection because", err)
		}

		if err := cmd.Execute(client, args); err != nil {
			log.Println("failed to execute", cmd.Name(), "because", err)
		}
	*/
}
