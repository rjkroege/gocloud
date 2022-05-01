// Experimental code to determine idleness

package main

import (
	"flag"
	"log"
	"time"

	"github.com/rjkroege/gocloud/config"
	"github.com/rjkroege/gocloud/gcp"
	"github.com/rjkroege/gocloud/who"
)

// Flags
var (
	delaytime = flag.Int("delay", 60*15, "Time in seconds before indicating idleness.")
	dryrun    = flag.Bool("n", false, "log copiously and don't try to shut down for realz")
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

	// Make a keep-alive socket.
	c, err := setupkeepalive(config.LocalNameSpace(""))
	if err != nil {
		log.Println("setupkeepalive had a sad:", err)
	}

	for {
		waiter := time.NewTimer(idletime)

		select {
		case <-c:
			if *dryrun {
				log.Println("socket watcher saw activity, resetting timer")
			}

			if !waiter.Stop() {
				<-waiter.C
			}
			waiter.Reset(idletime)
		case <-waiter.C:

			if err := who.UpdateWhoList(wholist); err != nil {
				log.Println("UpdateWhoList had a sad because", err)
			}

			if who.AreIdle(wholist, idletime) {
				if *dryrun {
					log.Println("Would now do something responding to idleness")
					continue
				}

				if err := gcp.EndSession(&config.Settings{}, ""); err != nil {
					log.Println("failed to EndSession:", err)
				}
			}

			if *dryrun {
				log.Println("not idle.")
			}

		}
	}
}
