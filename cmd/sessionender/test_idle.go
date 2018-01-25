// Experimental code to determine idleness

package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/rjkroege/gocloud/gcp"
	"github.com/rjkroege/gocloud/who"
	"golang.org/x/oauth2/google"
)


// Flags
var (
	delaytime = flag.Int("delay", 60 * 15, "Time in seconds before indicating idleness.")
	dryrun = flag.Bool("n", false, "log copiously and don't try to shut down for realz")
)

func main() {
	flag.Parse()
	idletime := time.Duration(*delaytime) * time.Second

	if *dryrun {
		log.Println("waiting for", idletime)
	}
	
	wholist := who.WhoList{}
	if err := who.UpdateWhoList(wholist); err != nil {
		log.Println("UpdateWhoList had a sad because", err)
	}
	
	if *dryrun {
		log.Println("starting wholist", wholist)
	}

	for {
		waiter := time.NewTimer(idletime)
		<-waiter.C

		if err := who.UpdateWhoList(wholist); err != nil {
			log.Println("UpdateWhoList had a sad because", err)
		}

		if who.AreIdle(wholist, idletime) {
			if *dryrun {
				log.Println("Would now do something responding to idleness")
				continue
			}

			cmd := gcp.MakeEndSession()
			ctx := context.Background()
			client, err := google.DefaultClient(ctx, cmd.Scope())
			if err != nil {
				log.Println("Can't setup an OAuth connection because", err)
			}
			if err := cmd.Execute(client, []string{}); err != nil {
				log.Println("failed to execute", cmd.Name(), "because", err)
			}
		}

		if *dryrun {
			log.Println("not idle.")
		}
	}
}
