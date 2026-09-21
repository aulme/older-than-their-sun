// Package data holds the lookup tables the simulation's output is read
// through: what a key means, how a name is made. The simulation never
// reads them; the names pass and the view do. Every file is JSON and is
// embedded, so a binary carries its own tables.
package data

import "embed"

// FS is every table, by file name.
//
//go:embed *.json
var FS embed.FS
