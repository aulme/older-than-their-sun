package mind

import "math/rand/v2"

// The research pick: a lottery over what is open, weighted by species
// tilt, focus, momentum in the domain and aptitude; a miracle by the leap.

// Pursuit is one open node as the chooser weighs it.
type Pursuit struct {
	Weight   float64 // the node's own
	Domain   float64 // the species' tilt to the domain
	Focus    float64 // current focus on the domain, 1 when none
	Depth    int     // nodes already known in the domain
	Aptitude float64 // the cost multiplier of the aptitude; divides
	Miracle  bool
	Leap     float64 // the leap weight, for a miracle
}

// Choose weighs the open nodes and draws one; -1 when nothing is open,
// and then nothing is drawn. The weights are returned for the trace.
func Choose(open []Pursuit, r *rand.Rand, t *Tuning) (int, []float64) {
	weights := make([]float64, len(open))
	total := 0.0
	for i, n := range open {
		wt := n.Weight * n.Domain * n.Focus * (1 + t.Research.DepthBonus*float64(n.Depth)) / n.Aptitude
		if n.Miracle {
			wt = n.Weight * n.Leap
		}
		weights[i] = wt
		total += wt
	}
	if len(open) == 0 || total <= 0 {
		return -1, weights
	}
	x := r.Float64() * total
	for i := range open {
		x -= weights[i]
		if x < 0 {
			return i, weights
		}
	}
	return len(open) - 1, weights
}
