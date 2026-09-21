package history

import (
	"fmt"
	"math"
	"sort"
	"strconv"

	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// knownOf lists what a civilisation knows in tree order. Ranging over the
// map directly would let Go's map order into the random stream and make a
// seed irreproducible; every loop that draws from the RNG uses this.
func knownOf(c *Civ) []string {
	var out []string
	for _, n := range tech.Nodes {
		if c.Known[n.Key] {
			out = append(out, n.Key)
		}
	}
	return out
}

// sortedInts lists the keys of an int set in order, for the same reason.
func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortedInts[V any](m map[int]V) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

func sprintf(format string, args ...any) string { return fmt.Sprintf(format, args...) }

func (w *World) trace(star int, kind string, civ int) {
	w.Traces = append(w.Traces, Trace{Star: star, Kind: kind, Civ: civ, Year: w.Now})
}

// Names. The simulation works on ids and never holds a name; a line that
// wants one carries a token, {kind:id}, that the view resolves through
// the names pass (internal/names). {^kind:id} is the same with its first
// letter raised, for a sentence start; {kind:id@by} is what one people
// calls the thing.

// star is a star's token.
func (w *World) star(id int) string { return "{star:" + itoa(id) + "}" }

// Tok is a people's token.
func (c *Civ) Tok() string { return "{civ:" + itoa(c.ID) + "}" }

// tokBy is a people's token as another people names it.
func (c *Civ) tokBy(by *Civ) string { return "{civ:" + itoa(c.ID) + "@" + itoa(by.ID) + "}" }

// Tok is a plague's token; the name carries its article.
func (p *Plague) Tok() string { return "{plague:" + itoa(p.ID) + "}" }

// Tok is a war's token: "the war of X", or nothing if the war has no name yet.
func (wr *War) Tok() string { return "{war:" + itoa(wr.ID) + "}" }

// Tok is an elder's token: what its finders call its makers.
func (e *Elder) Tok() string { return "{elder:" + itoa(e.ID) + "}" }

// speciesTok is a blood's token: the name of its first people.
func speciesTok(sp *species.Species) string { return "{species:" + itoa(sp.ID) + "}" }

// wordTok is a people's word for the state beneath, "" until it has one.
func (c *Civ) wordTok() string { return "{word:" + itoa(c.ID) + "}" }

// titleTok is a remnant's ruler title.
func (c *Civ) titleTok() string { return "{title:" + itoa(c.ID) + "}" }

// makersTok is a finder's name for the makers of a remain it cannot place.
func makersTok(l *Legacy) string { return "{makers:" + itoa(l.ID) + "}" }

// chance rolls a probability given per thousand years, scaled to the
// current tick so the middle and fine passes share one set of rates.
func (w *World) chance(p float64) bool {
	if p <= 0 {
		return false
	}
	if p >= 1 {
		return true
	}
	return w.R.Float64() < 1-math.Pow(1-p, w.dt)
}

// count draws how many times a per-kyr rate fires in the current tick.
func (w *World) count(rate float64) int {
	x := rate * w.dt
	n := int(x)
	if w.R.Float64() < x-float64(n) {
		n++
	}
	return n
}

func (w *World) pick(xs []int) int { return xs[w.R.IntN(len(xs))] }

func remove(xs []int, v int) []int {
	for i, x := range xs {
		if x == v {
			return append(xs[:i], xs[i+1:]...)
		}
	}
	return xs
}

func contains(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

func systems(n int) string {
	if n == 1 {
		return "a single world"
	}
	return sprintf("%d systems", n)
}

func clamp(x, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, x)) }

// LevelName turns a level into words.
func LevelName(x float64) string {
	switch {
	case x < 1.5:
		return "negligible"
	case x < 3:
		return "weak"
	case x < 5:
		return "modest"
	case x < 7:
		return "strong"
	case x < 8.5:
		return "formidable"
	default:
		return "overwhelming"
	}
}

func percent(x float64) string {
	if x < 0.01 {
		return "less than a hundredth"
	}
	return sprintf("%.0f%%", x*100)
}

// itoa is strconv.Itoa, for keys.
func itoa(n int) string { return strconv.Itoa(n) }
