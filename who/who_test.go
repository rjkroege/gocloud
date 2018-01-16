package who

import (
	"testing"
)

const who_out_valid = `rjkroege pts/0        2018-01-10 10:58 (135.23.127.39)
`

const who_out_garbage_line = `
fuzzy
rjkroege pts/0        2018-01-10 10:58 (135.23.127.39)
`

func TestParseWho(t *testing.T) {
	whobytes := []byte{}
	got, err := ParseWho(whobytes)
	if err != nil {
		t.Fatal("expected success, got", err)
	}
	if len(got) != 0 {
		t.Fatal("expected 0 length array, got", got)
	}

	for _, who_out := range []string{who_out_valid, who_out_garbage_line} {
		whobytes := []byte(who_out)
		got, err := ParseWho(whobytes)
		if err != nil {
			t.Fatal("expected success, got", err)
		}
		if len(got) != 1 {
			t.Fatal("expected 1 val, got", got)
		}
		if _, ok := got["/dev/pts/0"]; !ok {
			t.Fatal("expected val to be present but", got)
		}
	}
}
