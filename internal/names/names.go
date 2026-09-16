// Package names generates names for civilisations, stars, horrors and rulers.
// Purely syllabic for now; later this could be per-civilisation phonologies.
package names

import (
	"math/rand/v2"
	"strings"
)

var onsets = []string{"k", "t", "v", "sh", "r", "m", "n", "l", "th", "z", "kh", "d", "g", "s", "h", "y", "q", "dr", "tr", "vr", "ph", "ts"}
var nuclei = []string{"a", "e", "i", "o", "u", "a", "e", "o", "ai", "au", "ei", "ao", "ia", "uu"}
var codas = []string{"", "", "", "", "n", "r", "l", "sh", "th", "k", "m", "s", "x", "nd", "rr", "th"}

func pick(r *rand.Rand, xs []string) string { return xs[r.IntN(len(xs))] }

// Word builds a pronounceable word of the given syllable count.
func Word(r *rand.Rand, syllables int) string {
	var b strings.Builder
	for i := 0; i < syllables; i++ {
		if i > 0 || r.Float64() < 0.8 {
			b.WriteString(pick(r, onsets))
		}
		b.WriteString(pick(r, nuclei))
		if i == syllables-1 || r.Float64() < 0.3 {
			b.WriteString(pick(r, codas))
		}
	}
	w := b.String()
	return strings.ToUpper(w[:1]) + w[1:]
}

// Civ returns a civilisation name such as "Vashmuri".
func Civ(r *rand.Rand) string { return Word(r, 2+r.IntN(2)) }

// Star returns a proper name for a star, given by whoever lives there.
func Star(r *rand.Rand) string { return Word(r, 1+r.IntN(2)) }

var titleAdj = []string{"Eternal", "Last", "Undying", "Hollow", "Radiant", "Silent", "Grey", "Thousandth", "Sleepless", "Forgotten", "Ashen", "Unbroken", "Deathless", "Weeping"}
var titleNoun = []string{"Emperor", "Regent", "Matriarch", "Hierarch", "Custodian", "Speaker", "Archon", "Sovereign", "Steward", "Oracle", "Warden", "Autarch"}

// Title returns a ruler title for a contracted remnant civilisation.
func Title(r *rand.Rand) string { return "the " + pick(r, titleAdj) + " " + pick(r, titleNoun) }

var horrorAdj = []string{"Grey", "Hungry", "Patient", "Bright", "Quiet", "Cold", "Black", "Endless", "Pale", "Whispering"}
var horrorNoun = []string{"Tide", "Swarm", "Bloom", "Choir", "Sleeper", "Lattice", "Signal", "Host", "Garden", "Engine"}

// Horror returns an epithet such as "the Grey Tide".
func Horror(r *rand.Rand) string { return "the " + pick(r, horrorAdj) + " " + pick(r, horrorNoun) }
