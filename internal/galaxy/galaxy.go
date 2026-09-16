// Package galaxy holds the fixed substrate: stars and their positions.
//
// v1 is still a random star field around Sol with a realistic spectral class
// mix, but stars now carry multiplicity, a main-sequence lifetime and a
// scheduled death, so the history can face civilisations with supernovae
// and dying suns. Positions are in light years, Sol at the origin, Z through
// the galactic disc.
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
	Name     string // catalogue designation, replaced by a proper name if someone lives there
	Class    byte   // spectral class O B A F G K M; W (white dwarf) or N (neutron star, black hole) once dead
	X, Y, Z  float64
	Hab      float64 // crude habitability weight in [0,1], used to decide where life arises
	Mult     int     // 1 single, 2 binary, 3 trinary
	Lifetime float64 // main-sequence lifetime in years
	DiesAt   int64   // year relative to the dawn of the current age when the star leaves the main sequence
	Failing  bool    // the star has begun to die; its worlds are degrading
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

// ClassName describes the star's class and multiplicity: "K-class binary".
func (s *Star) ClassName() string {
	m := map[int]string{1: "star", 2: "binary", 3: "trinary"}[s.Mult]
	switch s.Class {
	case 'W':
		return "white dwarf"
	case 'N':
		return "stellar remnant"
	}
	return fmt.Sprintf("%c-class %s", s.Class, m)
}

// Galaxy is the star field plus precomputed pairwise distances.
type Galaxy struct {
	Stars  []Star
	Sol    int
	Radius float64
	dist   []float64 // n*n
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

// Generate builds a random disc-shaped star field. Real stellar density near
// Sol is about 0.004 stars per cubic light year, which would be far too many
// to simulate individually; the field is sparse on purpose.
func Generate(r *rand.Rand, n int, radius, thickness float64) *Galaxy {
	g := &Galaxy{Radius: radius}
	g.Stars = make([]Star, 0, n)
	g.Stars = append(g.Stars, Star{ID: 0, Name: "Sol", Class: 'G', Hab: 1.0, Mult: 1, Lifetime: 1e10, DiesAt: 5e9})
	g.Sol = 0
	for i := 1; i < n; i++ {
		rr := radius * math.Sqrt(r.Float64())
		th := r.Float64() * 2 * math.Pi
		z := (r.NormFloat64()) * thickness / 2
		c, hab, life := pickClass(r)
		mult := 1
		switch x := r.Float64(); {
		case x < 0.07:
			mult = 3
		case x < 0.45:
			mult = 2
		}
		s := Star{
			ID: i, Name: fmt.Sprintf("HIP-%d", 1000+r.IntN(90000)),
			Class: c, X: rr * math.Cos(th), Y: rr * math.Sin(th), Z: z, Hab: hab,
			Mult: mult, Lifetime: life, DiesAt: diesAt(r, c, life),
		}
		if s.DiesAt < -Age {
			s.Kill()
		}
		g.Stars = append(g.Stars, s)
	}
	g.dist = make([]float64, n*n)
	for a := 0; a < n; a++ {
		for b := a + 1; b < n; b++ {
			d := math.Sqrt(sq(g.Stars[a].X-g.Stars[b].X) + sq(g.Stars[a].Y-g.Stars[b].Y) + sq(g.Stars[a].Z-g.Stars[b].Z))
			g.dist[a*n+b] = d
			g.dist[b*n+a] = d
		}
	}
	return g
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
	s := &g.Stars[id]
	if id == g.Sol {
		return "Sol"
	}
	return fmt.Sprintf("%s (%s, %.0f ly from Sol)", s.Name, s.ClassName(), g.Dist(id, g.Sol))
}
