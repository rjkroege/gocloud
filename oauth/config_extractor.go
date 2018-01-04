// This is replicated from specialremote.
// TODO(refactor that there)

package oauth

import (
	"encoding/json"
	"io"
	"log"
)

func hunter(forest map[string]interface{}, wanted map[string]string) {
	for k, v := range forest {
		switch tv := v.(type) {
		case map[string]interface{}:
			hunter(tv, wanted)
		case string:
			if _, ok := wanted[k]; ok {
				wanted[k] = tv
			}
		case []string:

		default:
			log.Println("don't know how to deal: ", k, v)
		}
	}
}

// FindProperties hunts through a JSON-encoded input source for each key
// in wanted and updates the key with the found value if it exists. The
// routine attempts to be flexible about the contents of the input
// stream. Returns decoding errors.
func FindProperties(r io.Reader, wanted map[string]string) error {
	dec := json.NewDecoder(r)
	m := make(map[string]interface{})
	if err := dec.Decode(&m); err != nil {
		return err
	}
	hunter(m, wanted)
	return nil	
}

