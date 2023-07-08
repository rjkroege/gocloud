package who

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTimestampWhoList(t *testing.T) {

	// Setup a directory.
	// that's not going to be convenient...

	tdir, err := ioutil.TempDir("", "meldwholist")
	if err != nil {
		t.Fatal("couldn't make the temporary directory because", err)
	}
	defer os.RemoveAll(tdir)

	// Make some standard paths
	p0 := filepath.Join(tdir, "0")
	p1 := filepath.Join(tdir, "1")
	p2 := filepath.Join(tdir, "2")

	// Make a single file in the temp directory.
	if _, err = os.Create(p1); err != nil {
		t.Fatal("couldn't make a file", p1, "because", err)
	}
	// Make a single file in the temp directory.
	if _, err = os.Create(p2); err != nil {
		t.Fatal("couldn't make a file", p2, "because", err)
	}

	now := time.Now()

	// Make an inputwholist
	inputwholist := WhoList{
		p0: now.Add(-time.Minute),
		p1: now.Add(-time.Minute),
	}

	TimestampWhoList(inputwholist)

	if _, ok := inputwholist[p0]; ok {
		t.Error("inputwholist should have p0 removed", inputwholist)
	}
	if ti, ok := inputwholist[p1]; !ok || now.Sub(ti) > time.Duration(10*time.Second) {
		t.Error("inputwholist should have p1 updated", inputwholist)
	}
	if _, ok := inputwholist[p2]; ok {
		t.Error("inputwholist should have not have added p2", inputwholist)
	}
}
