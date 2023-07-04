package who

import (
	"os"
	"time"
)

func TimestampWhoList(wl WhoList) {
	for k := range wl {
		fi, err := os.Stat(k)
		if err != nil {
			delete(wl, k)
			continue
		}
		wl[k] = fi.ModTime()
	}
}

func MergeWhoList(oldwl, newwl WhoList) {
	for k := range newwl {
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

	TimestampWhoList(currentwl)
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
