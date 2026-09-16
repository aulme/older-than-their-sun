package history

import "math"

// The appraisal is one estimate of a fight against a target, made from
// beliefs, that every choice routes through: whether to declare, to send a
// fleet, to scout, to accept a pact, to come when called, to sue.

// Appraisal is what a people thinks of a fight.
type Appraisal struct {
	Margin float64 // believed strength difference, attacker minus defender
	Spread float64 // uncertainty of the belief, in levels
	Odds   float64 // on the mean
	Low    float64 // on the pessimistic tail
	High   float64 // on the hopeful tail
	Acted  float64 // what this people acts on, by its risk dial
	Front  []int   // the enemy's worlds within strike reach, nearest first
	Target int     // the world appraised, or -1
	Lag    float64 // years for a strike or a fleet to arrive
}

// phi is the normal cumulative distribution.
func phi(x float64) float64 { return 0.5 * math.Erfc(-x/math.Sqrt2) }

// strength is what c brings against e: its level at home, its miracles,
// its allies who would actually join, less what its other wars take.
func (w *World) strength(c, e *Civ) float64 {
	s := c.Mil + c.warBonus()
	for _, pid := range c.Pacts {
		p := w.Pacts[pid]
		if p.Over || p.Kind == Defensive && !c.Wars[e.ID] {
			continue
		}
		for _, mid := range p.Members {
			m := w.Civs[mid]
			if m != c && m.Active() && len(w.front(m, e)) > 0 {
				s += 0.5 * m.Mil
			}
		}
	}
	for eid := range c.Wars {
		if eid != e.ID {
			s -= 0.3
		}
	}
	return s
}

// appraise estimates c's fight against e at a world: the nearest front
// world, or the given one, or e's home if nothing is in reach.
func (w *World) appraise(c, e *Civ, target int) Appraisal {
	a := Appraisal{Target: target}
	a.Front = w.front(c, e)
	if a.Target < 0 {
		if len(a.Front) > 0 {
			a.Target = a.Front[0]
		} else {
			_, a.Target = w.nearestEnemy(c, e) // what a fleet would go for
		}
	}
	mil, spread := w.believe(c, e)
	d := mil + e.warBonus() + 1
	if a.Target == e.Home {
		d += 2.5
	}
	if i := c.Intel[e.ID]; i != nil {
		if i.Grid {
			d += 0.5
		}
		if i.Star == a.Target {
			d += i.Relief
		}
	}
	if e.Plagued || w.Now-e.LastDark < 50_000 {
		d -= 1
	}
	for eid := range e.Wars {
		if eid != c.ID {
			d -= 0.3
		}
	}
	a.Margin = w.strength(c, e) - d
	a.Spread = spread
	from, dist := w.nearest(c, a.Target)
	_ = from
	a.Lag = dist * c.Speed
	a.Odds = phi(a.Margin / 2)
	a.Low = phi((a.Margin - spread) / 2)
	a.High = phi((a.Margin + spread) / 2)
	a.Acted = phi((a.Margin + (2*c.Dials.Risk-1)*spread) / 2)
	return a
}

// nearest is c's holding nearest a star, and the distance.
func (w *World) nearest(c *Civ, star int) (int, float64) {
	best, bd := c.Home, w.G.Dist(c.Home, star)
	for _, s := range c.Systems {
		if d := w.G.Dist(s, star); d < bd {
			best, bd = s, d
		}
	}
	return best, bd
}

// bar is the odds a posture needs before it strikes at e, and whether it
// would strike at all; far says whether it would send a fleet beyond the
// front to do it.
func (w *World) bar(c, e *Civ) (bar float64, wants, far bool) {
	p := c.posture()
	if p == "pacifist" {
		return 0, false, false
	}
	if c.hates(e) {
		return 0.35, true, true
	}
	switch p {
	case "opportunist":
		bar, wants = 0.75, true
	case "conqueror":
		bar, wants, far = 0.4, true, true
	case "vengeful":
		if c.Grudge[e.ID] > 0 {
			bar, wants, far = 0.3, true, true
		}
	}
	if wants && c.Grudge[e.ID] > 0 && p != "vengeful" {
		bar -= 0.1
	}
	return
}
