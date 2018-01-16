// Experimental code to determine idleness

package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/rjkroege/sessionender/gcp"
	"github.com/rjkroege/sessionender/who"
	"golang.org/x/oauth2/google"
)

const idletime = time.Minute * 15

func main() {
	logfile, err := os.Create("otherlog")
	if err != nil {
		log.Fatalln("can't open file to log because", err)
	}
	// So that I don't perturb standard I/O with log messages.
	log.SetOutput(logfile)

	wholist := who.WhoList{}
	for {
		waiter := time.NewTimer(time.Minute)
		<-waiter.C

		if err := who.UpdateWhoList(wholist); err != nil {
			log.Println("UpdateWhoList had a sad because", err)
		}

		log.Println("WhoList: ", wholist)

		if who.AreIdle(wholist, idletime) {
			log.Println("Would now do something responding to idleness")

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
	}
}
