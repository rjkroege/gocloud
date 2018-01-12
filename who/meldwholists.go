package who

import (
	"log"
	"time"
	"os"
)

func TimestampWhoList(wl WhoList) error {
	var savederr error
	for k, _ := range wl {
		fi, err := os.Stat(k)
		if err != nil {
			log.Println("couldn't stat pty path", k, "because", err)
			delete(wl, k)
			// I should aggregate them.
			savederr = err
		}
		wl[k] = fi.ModTime()
	}
	return savederr
}

func MergeWhoList(oldwl, newwl WhoList) {
	for k, _ := range newwl {
		if _, ok := oldwl[k]; !ok {
			oldwl[k] = newwl[k]
		}
	}
}

func UpdateWhoList(currentwl WhoList) error {
	whos, err := RunWho()
	if err != nil {
		return err
	}	

	newwl, err := ParseWho(whos)
	if err != nil {
		return err
	}

	MergeWhoList(currentwl, newwl)
	
	if err :=TimestampWhoList(currentwl); err != nil {
		return err
	}

	return nil
}

func AreIdle(currentwl WhoList, idletime time.Duration) bool {
	now := time.Now()
	for _, t := range currentwl {
		if now.Sub(t) < idletime {
			return false
		}
	}
	return true
}