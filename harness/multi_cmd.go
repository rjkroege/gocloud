package harness

// Provides basic infrastructure for supporting multiple different
// subcommands in a single executable.

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type Command interface {
	// Execute runs the command with the provided client and arguments
	// and returns any error that precluded the command from executing.
	Execute(*http.Client, []string) error

	// Scope returns the authentication scope needed for this command
	Scope() string

	// Name returns the name of this command
	Name() string

	// Usage returns a (possibly multi-line) string describing how to execute this subcommand.
	Usage() string
}

var subcommands = make(map[string]Command)

// AddSubCommand adds a new subcommand to the harness.
func AddSubCommand(c Command) {
	name := c.Name()
	if _, ok := subcommands[name]; ok {
		log.Fatalln(name, " already registered")
	}
	subcommands[name] = c
}

func Cmd(name string) (Command, bool) {
	c, ok := subcommands[name]
	return c, ok
}

func Usage(w io.Writer) error {
	for _, c := range subcommands {
		if _, err := fmt.Fprintf(os.Stderr, "%s\n\n", c.Usage()); err != nil {
			return err
		}
	}
	return nil
}
