// Package galaxy holds the fixed substrate: the Milky Way, its laws, its
// named features, the real stars near the Sun, and the star field a
// history runs in.
//
// A field is a few hundred stars around a point in the galaxy (see
// field.go). Positions within it are in light years from the field's
// centre, x toward the galactic centre, y along rotation, z north. Stars
// carry multiplicity, a main-sequence lifetime and a scheduled death, so
// the history can face civilisations with supernovae and dying suns, and
// each has a system of worlds (see system.go).
package galaxy

import (
	"fmt"
	"math"
	"math/rand/v2"
	"sort"
)

// Age is how far back the substrate is simulated, in years before the dawn of the current age.
// Stars that died before this are remnants from the start.
const Age = 7_000_000_000

// Star is one star system.
type Star struct {
	ID       int
	Name     string // the human catalogue's label for a real star, a code for a synthetic one; never overwritten: what its peoples call it is the names pass's
	Class    byte   // spectral class O B A F G K M; W (white dwarf) or N (neutron star, black hole) once dead
	X, Y, Z  float64
	Hab      float64 // crude habitability weight in [0,1], used to decide where life arises
	Mult     int     // 1 single, 2 binary, 3 trinary
	Lifetime float64 // main-sequence lifetime in years
	DiesAt   int64   // year relative to the dawn of the current age when the star leaves the main sequence
	Failing  bool    // the star has begun to die; its worlds are degrading
	Real     bool    // a catalogued star, or a named feature
	Alt      string  // another designation
	Mag      float64 // apparent magnitude from Earth; 99 if not visible or not real
	Note     string  // giant, supergiant, white dwarf, brown dwarf, subdwarf
	Remnant  string  // for class N, a key: neutron_star, magnetar, black_hole, great_hole; RemnantWords say them
	cat      *CatStar
}

// Proper says whether the human name is a proper name (Sirius, Ran) as
// against a catalogue designation (HIP 56601): a strong human name, which
// the view leads with. A synthetic star has neither.
func (s *Star) Proper() bool {
	if !s.Real || s.Name == "" {
		return false
	}
	return s.cat == nil || s.cat.Proper()
}

// Hostility is how hard the star's worlds are to live on, for the habitable
// envelope: 0 gentle, 1 harsh, 2 lethal.
func (s *Star) Hostility() int {
	switch s.Class {
	case 'G', 'K':
		return 0
	case 'F', 'M':
		return 1
	default:
		return 2
	}
}

// Massive stars end as supernovae.
func (s *Star) Massive() bool { return s.Class == 'O' || s.Class == 'B' }

// Dead is true for stellar remnants.
func (s *Star) Dead() bool { return s.Class == 'W' || s.Class == 'N' }

// Kill turns the star into its remnant.
func (s *Star) Kill() {
	if s.Massive() {
		s.Class = 'N'
	} else {
		s.Class = 'W'
	}
	s.Hab = 0
	s.Failing = false
}

// RemnantWords say what a dead star's remnant key is.
var RemnantWords = map[string]string{"neutron_star": "neutron star", "magnetar": "magnetar", "black_hole": "black hole", "great_hole": "the great hole"}

// CompanionWords say what a system's companion key is.
var CompanionWords = map[string]string{"white_dwarf": "a white dwarf companion", "brown_dwarf": "a brown dwarf companion"}

// ClassName describes the star's class and multiplicity: "K-class binary".
func (s *Star) ClassName() string {
	m := map[int]string{1: "star", 2: "binary", 3: "trinary"}[s.Mult]
	switch s.Class {
	case 'W':
		return "white dwarf"
	case 'N':
		if s.Remnant != "" {
			return RemnantWords[s.Remnant]
		}
		return "stellar remnant"
	}
	if s.Note != "" && s.Note != "subdwarf" {
		return fmt.Sprintf("%c-class %s", s.Class, s.Note)
	}
	return fmt.Sprintf("%c-class %s", s.Class, m)
}

// Galaxy is the star field plus precomputed pairwise distances.
type Galaxy struct {
	Stars     []Star
	Sys       []*System
	Sol       int // index of Sol, or -1 if the field is elsewhere
	Radius    float64
	Thickness float64
	Region    Region
	Law       Law
	dist      []float64 // n*n
}

// Anchor names the field's centre: Sol, or the feature or place it is at.
func (g *Galaxy) Anchor() string {
	if g.Sol >= 0 {
		return "Sol"
	}
	if g.Region.Anchor != nil {
		return g.Region.Anchor.Name
	}
	return "the centre of the field"
}

// FromCentre is a star's distance from the field's centre in light years.
func (g *Galaxy) FromCentre(id int) float64 {
	s := &g.Stars[id]
	return math.Sqrt(sq(s.X) + sq(s.Y) + sq(s.Z))
}

// Position of a star in the galaxy.
func (g *Galaxy) Position(id int) Vec {
	s := &g.Stars[id]
	return g.Region.Pos.Add(Vec{s.X / LyPerKpc, s.Y / LyPerKpc, s.Z / LyPerKpc})
}

func (g *Galaxy) index() {
	n := len(g.Stars)
	g.dist = make([]float64, n*n)
	for a := 0; a < n; a++ {
		for b := a + 1; b < n; b++ {
			d := math.Sqrt(sq(g.Stars[a].X-g.Stars[b].X) + sq(g.Stars[a].Y-g.Stars[b].Y) + sq(g.Stars[a].Z-g.Stars[b].Z))
			g.dist[a*n+b] = d
			g.dist[b*n+a] = d
		}
	}
}

// Approximate fraction of main-sequence stars by class, a habitability
// weight per class, and the main-sequence lifetime. M dwarfs get life
// sometimes but tidal locking and flares hurt; hot stars burn out too fast.
var classWeights = []struct {
	c    byte
	w    float64
	hab  float64
	life float64
}{
	{'M', 0.76, 0.3, 1e12}, {'K', 0.12, 0.85, 3e10}, {'G', 0.076, 1.0, 1e10},
	{'F', 0.03, 0.5, 3e9}, {'A', 0.006, 0.0, 1e9}, {'B', 0.0013, 0.0, 5e7}, {'O', 0.00003, 0.0, 5e6},
}

func pickClass(r *rand.Rand) (byte, float64, float64) {
	x := r.Float64()
	for _, cw := range classWeights {
		if x < cw.w {
			return cw.c, cw.hab, cw.life
		}
		x -= cw.w
	}
	return 'M', 0.3, 1e12
}

// diesAt schedules a star's death. This is a fudge, not astrophysics: the
// point is that some stars die during the simulated window so that
// supernovae happen and a few species are born under a failing sun.
func diesAt(r *rand.Rand, class byte, life float64) int64 {
	switch class {
	case 'O', 'B', 'A':
		return int64(-life + r.Float64()*2*life)
	case 'F', 'G', 'K':
		x := r.Float64()
		switch {
		case x < 0.10: // an old star, dead or dying
			return int64(-3e9 + r.Float64()*3.3e9)
		case x < 0.14: // dying during the current age
			return int64(-2e7 + r.Float64()*1.2e8)
		}
		return int64(life * (0.2 + 0.8*r.Float64()))
	}
	return 1e15
}

// Generate builds the field of the Sun's neighbourhood.
func Generate(r *rand.Rand, n int, radius, thickness float64) *Galaxy {
	rg, _ := RegionByName("sol")
	return GenerateAt(r, rg, n, radius, thickness)
}

func sq(x float64) float64 { return x * x }

// Dist returns the distance in light years between two stars.
func (g *Galaxy) Dist(a, b int) float64 { return g.dist[a*len(g.Stars)+b] }

// Near returns star IDs within maxDist of a, nearest first, excluding a.
func (g *Galaxy) Near(a int, maxDist float64) []int {
	var out []int
	for b := range g.Stars {
		if b != a && g.Dist(a, b) <= maxDist {
			out = append(out, b)
		}
	}
	sort.Slice(out, func(i, j int) bool { return g.Dist(a, out[i]) < g.Dist(a, out[j]) })
	return out
}

// Describe gives a short human-readable location for a star.
func (g *Galaxy) Describe(id int) string {
	return g.DescribeStar(&g.Stars[id], id)
}

// DescribeStar is Describe of a star as given, for a star as it was at
// some earlier time.
func (g *Galaxy) DescribeStar(s *Star, id int) string {
	if id == g.Sol {
		return "Sol"
	}
	return fmt.Sprintf("%s (%s, %.0f ly from %s)", s.Name, s.ClassName(), g.FromCentre(id), g.Anchor())
}
