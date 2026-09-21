// Package names generates names for civilisations, stars, plagues and rulers.
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

var plagueAdj = []string{"Red", "Grey", "Weeping", "Glass", "Silent", "Sweating", "Hollow", "Slow", "Quick", "Black", "White", "Wandering", "Crystal"}
var plagueBody = []string{"Fever", "Rot", "Blight", "Wasting", "Sleep", "Cough", "Bloom", "Pox", "Fade", "Sweat", "Ague"}
var plagueMind = []string{"Song", "Question", "Doctrine", "Laugh", "Silence", "Certainty", "Word", "Number", "Joke", "Prayer", "Dream", "Argument"}

// Plague names a sickness: "the Grey Rot" for one of the body, "the Silent
// Question" for one of the mind. One in three is named for its first host
// instead, "the Qaosh Sweat" or "the Fever of Wolf 359", which is what the
// neighbours call it; host is the people's name or the star's, and named
// says whether that was done.
func Plague(r *rand.Rand, memetic bool, host string) (name string, named bool) {
	noun := pick(r, plagueBody)
	if memetic {
		noun = pick(r, plagueMind)
	}
	if host != "" && r.IntN(3) == 0 {
		if r.IntN(2) == 0 {
			return "the " + host + " " + noun, true
		}
		return "the " + noun + " of " + host, true
	}
	return "the " + pick(r, plagueAdj) + " " + noun, false
}
