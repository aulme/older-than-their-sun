// Package names is the names pass: a pure function from the record of a
// run to the names things carry. The simulation works on ids and never
// draws for a name; this package reads what it left (the peoples, the
// facts, the catalogue) and coins every name from a hash of the seed, the
// namer, the thing and the relationship, so a name never changes because
// another name was added, a relationship was removed or the history's
// rolls moved. Every name the pass emits is a human rendering: a
// transcription of a sound, a translation of a meaning, or a catalogue
// designation.
package names

import (
	"hash/fnv"
	"math/rand/v2"
	"strconv"
	"strings"
)

// stream is the small generator every draw for one name comes from,
// seeded by the canonical string of what is being named and by whom.
// seed|by|kind|id|tone is the form; the parts are joined with '|'.
func stream(seed uint64, parts ...string) *rand.Rand {
	h := fnv.New64a()
	h.Write([]byte(strconv.FormatUint(seed, 10)))
	for _, p := range parts {
		h.Write([]byte{'|'})
		h.Write([]byte(p))
	}
	s := h.Sum64()
	return rand.New(rand.NewPCG(s, s^0x9e3779b97f4a7c15))
}

func itoa(n int) string { return strconv.Itoa(n) }

func pick(r *rand.Rand, xs []string) string {
	if len(xs) == 0 {
		return ""
	}
	return xs[r.IntN(len(xs))]
}

// capital is the name with its first letter upper-cased.
func capital(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
