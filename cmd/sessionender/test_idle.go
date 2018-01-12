// Experimental code to determine idleness

package main

import (
	"log"
	"os"
	"time"

	"github.com/rjkroege/sessionender/who"

)


const  idletime = time.Minute * 2

func main() {
	logfile, err := os.Create("otherlog")
	if err != nil {
		log.Fatalln("can't open file to log because", err)
	}
	// So that I don't perturb standard I/O with log messages.
	log.SetOutput(logfile)

	
	wholist := who.WhoList{}
	for i := 0; i < 1000 ; i++  {
		waiter := time.NewTimer(time.Minute)
		<- waiter.C

		if err := who.UpdateWhoList(wholist); err != nil {
			log.Println("UpdateWhoList had a sad because", err)
		}
		
		log.Println("WhoList: ", wholist)

		if who.AreIdle(wholist, idletime) {
			log.Println("Would now do something responding to idleness")
			// TODO(rjk): Here is where I should shutdown the node.
		}


	}	

}
