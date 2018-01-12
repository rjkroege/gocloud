package who

import (
	"time"
	"bufio"
	"bytes"
	"log"
	"path/filepath"
)

type WhoList map[string]time.Time

// ParseWho parse the output of the who command and returns an array of pty entries.
func ParseWho(whocmdouput []byte) (WhoList, error) {
	list := make(WhoList)

	// Over lines
	scanner := bufio.NewScanner(bytes.NewBuffer(whocmdouput))
	for scanner.Scan() {
		line := scanner.Bytes()
		ls := bufio.NewScanner(bytes.NewBuffer(line))
		ls.Split(bufio.ScanWords)

		// We want the second word.
		if ls.Scan() && ls.Scan() {
			dpth := filepath.Join("/dev", ls.Text())
			list[dpth] = time.Time{}
		} else {
			log.Println("who line", string(line), "is not in the expected format, skipping")
		}
		// No need to check that we have read from the line buffer. The line must exist (though
		// might be empty)
	}
	if err := scanner.Err(); err != nil {
		log.Println("couldn't successfully read a line from the who file output because", err)
		return WhoList{}, nil
	}
	return list, nil
}
