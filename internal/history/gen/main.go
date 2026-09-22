// Command gen writes kinds_gen.go, in the history package and in the
// record package, from data/events.json: a constant per
// kind, F-prefixed for a fact and K-prefixed for the rest, named from the
// key in CamelCase. Run by go generate in the history package; a test
// checks the file is current.
package main

import (
	"fmt"
	"os"

	"worldgen/internal/history"
)

func main() {
	for path, pkg := range map[string]string{"kinds_gen.go": "history", "../record/kinds_gen.go": "record"} {
		if err := os.WriteFile(path, []byte(history.KindsSource(pkg)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
