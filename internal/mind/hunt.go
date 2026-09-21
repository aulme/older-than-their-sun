package mind

import "math"

// Fighting what cannot be seen. A people that cannot perceive what is
// taking its ships and its worlds can perceive its losses, and
// information about an absence is not information about the thing. The
// ledger of gaps is the losses with no doer, by where they fell; the
// deduction is a region, not a people.

// Loss is one doerless loss on the ledger: where it fell and when, in
// thousand-year ticks before now.
type Loss struct {
	X, Y float64
	Ago  float64
}

// Gap is a hole in the ledger: the losses' weighted centre, a radius that
// takes them all in, and how many there were.
type Gap struct {
	X, Y   float64
	Radius float64
	Losses int
}

// Deduce reads the ledger for a hole: the loss with the most others of
// the window inside the hunt radius of it, if they are enough. The
// centre is the mean of those, the radius the farthest of them from it,
// never less than a quarter of the hunt radius. Ties go to the earliest
// loss on the ledger, so the reading is the same for the same ledger.
func Deduce(losses []Loss, t *Tuning) (Gap, bool) {
	k := &t.Kinds
	var best []int
	for _, l := range losses {
		if l.Ago > k.HuntWindow {
			continue
		}
		var near []int
		for j, m := range losses {
			if m.Ago <= k.HuntWindow && math.Hypot(l.X-m.X, l.Y-m.Y) <= k.HuntRadius {
				near = append(near, j)
			}
		}
		if len(near) > len(best) {
			best = near
		}
	}
	if len(best) < k.HuntLosses {
		return Gap{}, false
	}
	g := Gap{Losses: len(best)}
	for _, j := range best {
		g.X += losses[j].X
		g.Y += losses[j].Y
	}
	g.X /= float64(len(best))
	g.Y /= float64(len(best))
	for _, j := range best {
		g.Radius = max(g.Radius, math.Hypot(losses[j].X-g.X, losses[j].Y-g.Y))
	}
	g.Radius = max(g.Radius, k.HuntRadius/4)
	return g, true
}
