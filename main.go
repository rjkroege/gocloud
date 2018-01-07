// Derived from https://github.com/google/google-api-go-client/blob/master/examples/main.go
//
// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"context"

        "golang.org/x/oauth2/google"
	"golang.org/x/oauth2"

)

// Flags
var (
	debug      = flag.Bool("debug", false, "show HTTP traffic")
)

// TODO(rjk): Update the usage message.
func usage() {
	fmt.Fprintf(os.Stderr, "Usage: go-api-demo <api-demo-name> [api name args]\n\nPossible APIs:\n\n")
	for n := range demoFunc {
		fmt.Fprintf(os.Stderr, "  * %s\n", n)
	}
	os.Exit(2)
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		usage()
	}

	// TODO(rjk): clean up how the demo is populated
	// What does the above mean anyway? I should write more detailed TODO

	name := filepath.Base(os.Args[0])
	demo, ok := demoFunc[name]
	args := flag.Args()
	if !ok {
		// Or the name might be the first argument.
		name = flag.Arg(0)
		demo, ok = demoFunc[name]
		args = flag.Args()[1:]
		if !ok {
			usage()
		}
	}

	ctx := context.Background()
	if *debug {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
			Transport: &logTransport{http.DefaultTransport},
		})
	}
	client, err := google.DefaultClient(ctx, demoScope[name])
	if err != nil {
		log.Fatalln("Can't setup an OAuth connection because", err)
	}
	demo(client, args)
}

var (
	demoFunc  = make(map[string]func(*http.Client, []string))
	demoScope = make(map[string]string)
)

// registerDemo adds an additional demo function to the the list of available
// demos.
// part of the harness.
func registerDemo(name, scope string, main func(c *http.Client, argv []string)) {
	if demoFunc[name] != nil {
		panic(name + " already registered")
	}
	demoFunc[name] = main
	demoScope[name] = scope
}
