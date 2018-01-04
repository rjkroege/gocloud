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

	"github.com/rjkroege/sessionender/oauth"
)

// Flags
var (
	cacheToken = flag.Bool("cachetoken", true, "cache the OAuth 2.0 token")
	debug      = flag.Bool("debug", false, "show HTTP traffic")
	configfile = flag.String("configfile", "config.json", "Name of configuration JSON file containing a ClientID and a ClientSecret JSON")
)

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

	// Get OAuth configuration.
	configmap, err := oauth.GetConfigMap(*configfile)
	if err != nil {
		log.Fatalln("Couldn't open oauth configuration", err)
	}

	transport := http.DefaultTransport
	if *debug {
		transport = &logTransport{http.DefaultTransport}
	}

	// TODO(rjk): push logging into a separate place
	c, _ := oauth.FriendlyNewOauthClient(
		configmap["client_id"],
		configmap["client_secret"],
		demoScope[name],
		*cacheToken,
		transport)

	//	c := newOAuthClient(ctx, config)
	demo(c, args)
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
