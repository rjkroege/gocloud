// Derived from https://github.com/google/google-api-go-client/blob/master/examples/main.go
//
// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rjkroege/gocloud/gcp"
	"github.com/rjkroege/gocloud/harness"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Flags
var (
	debug = flag.Bool("debug", false, "show HTTP traffic")
)

// TODO(rjk): Update the usage message.
func usage() {
	fmt.Fprintf(os.Stderr, "Usage: gocloud <subcommand> [api name args]\n\nPossible APIs:\n\n")
	harness.Usage(os.Stderr)
	os.Exit(2)
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		usage()
	}

	name := filepath.Base(os.Args[0])
	cmd, ok := harness.Cmd(name)
	args := flag.Args()
	if !ok {
		// Or the name might be the first argument.
		name = flag.Arg(0)
		cmd, ok = harness.Cmd(name)
		args = flag.Args()[1:]
		if !ok {
			usage()
		}
	}

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
}
