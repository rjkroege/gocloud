package config

import (
	"testing"
)

func TestLocalNameSpace(t *testing.T) {
	for _, tv := range []struct {
		input string
		want  string
	}{
		{
			"root",
			"/tmp/ns.root.:0",
		},
		{
			"",
			"/tmp/ns.rjkroege.:0",
		},
	} {
		if got, want := LocalNameSpace(tv.input), tv.want; got != want {
			t.Errorf("input %q: got %q but want %q", tv.input, got, want)
		}
	}
}
