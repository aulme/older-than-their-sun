// Command gen writes kinds_gen.go from data/events.json: a constant per
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
	if err := os.WriteFile("kinds_gen.go", []byte(history.KindsSource()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
