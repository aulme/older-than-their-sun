// Package data holds the lookup tables the simulation and its readers
// share: what a key means, what it does, how a name is made. Every table
// is JSON and is embedded, so a binary carries its own; the simulation
// reads its mechanics from them (a node's cost, a filter's difficulty, an
// event kind's weight) and never their text, which the names pass and the
// view read. Each file opens with a "_" header that says what it holds
// and the rule for its text: every term is a technical term, a phrase a
// reader who never saw this project understands at once or can search
// for, never a capitalised coinage.
package data

import (
	"embed"
	"encoding/json"
	"fmt"
)

// FS is every table, by file name.
//
//go:embed *.json
var FS embed.FS

// Load reads a table into v, and panics if the file is missing or
// malformed: the tables are part of the binary, and a bad one is a build
// error, not a runtime condition.
func Load(name string, v any) {
	b, err := FS.ReadFile(name)
	if err != nil {
		panic(fmt.Sprintf("data: %v", err))
	}
	if err := json.Unmarshal(b, v); err != nil {
		panic(fmt.Sprintf("data: %s: %v", name, err))
	}
}
