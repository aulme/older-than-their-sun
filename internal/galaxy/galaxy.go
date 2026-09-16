// Package galaxy holds the fixed substrate: stars and their positions.
//
// v0 is a random star field around Sol with a realistic spectral class mix.
// The intent is to replace this with real catalogue data (Gaia, HYG) for
// nearby stars and statistical fill further out. Positions are in light years,
// Sol at the origin, Z through the galactic disc.
package galaxy

import (
	"fmt"
	"math"
	"math/rand/v2"
	"sort"
)

// Star is one star system. Hab is a crude habitability weight in [0,1] used
// by the history simulation to decide where life can arise.
type Star struct {
	ID      int
	Name    string // catalogue designation, replaced by a proper name if someone lives there
	Class   byte   // spectral class O B A F G K M
	X, Y, Z float64
	Hab     float64
}

// Galaxy is the star field plus precomputed pairwise distances.
type Galaxy struct {
	Stars  []Star
	Sol    int
	Radius float64
	dist   []float64 // n*n
}

// Approximate fraction of main-sequence stars by class, and a habitability
// weight per class. M dwarfs get life sometimes but tidal locking and flares
// hurt; hot stars burn out too fast.
var classWeights = []struct {
	c   byte
	w   float64
	hab float64
}{
	{'M', 0.76, 0.3}, {'K', 0.12, 0.85}, {'G', 0.076, 1.0},
	{'F', 0.03, 0.5}, {'A', 0.006, 0.0}, {'B', 0.0013, 0.0}, {'O', 0.00003, 0.0},
}

func pickClass(r *rand.Rand) (byte, float64) {
	x := r.Float64()
	for _, cw := range classWeights {
		if x < cw.w {
			return cw.c, cw.hab
		}
		x -= cw.w
	}
	return 'M', 0.3
}

// Generate builds a random disc-shaped star field. Real stellar density near
// Sol is about 0.004 stars per cubic light year, which would be far too many
// to simulate individually; v0 is sparse on purpose.
func Generate(r *rand.Rand, n int, radius, thickness float64) *Galaxy {
	g := &Galaxy{Radius: radius}
	g.Stars = make([]Star, 0, n)
	g.Stars = append(g.Stars, Star{ID: 0, Name: "Sol", Class: 'G', Hab: 1.0})
	g.Sol = 0
	for i := 1; i < n; i++ {
		// uniform in a disc, gaussian-ish in Z
		rr := radius * math.Sqrt(r.Float64())
		th := r.Float64() * 2 * math.Pi
		z := (r.NormFloat64()) * thickness / 2
		c, hab := pickClass(r)
		g.Stars = append(g.Stars, Star{
			ID: i, Name: fmt.Sprintf("HIP-%d", 1000+r.IntN(90000)),
			Class: c, X: rr * math.Cos(th), Y: rr * math.Sin(th), Z: z, Hab: hab,
		})
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
	s := g.Stars[id]
	if id == g.Sol {
		return "Sol"
	}
	return fmt.Sprintf("%s (%c-class, %.0f ly from Sol)", s.Name, s.Class, g.Dist(id, g.Sol))
}
