// Package data holds the lookup tables the simulation's output is read
// through: what a key means, how a name is made. The simulation reads
// its mechanics from them (an event kind's sort and weight) and never
// their text; the names pass and the view read the rest. Every file is
// JSON and is embedded, so a binary carries its own tables.
package data

import "embed"

// FS is every table, by file name.
//
//go:embed *.json
var FS embed.FS
