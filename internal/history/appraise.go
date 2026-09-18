package history

import "worldgen/internal/mind"

// The appraisal is one estimate of a fight against a target, made from
// beliefs, that every choice routes through: whether to declare, to send a
// fleet, to scout, to accept a pact, to come when called, to sue. The
// arithmetic is mind.Appraise; this file gathers what it needs.

// Appraisal is what a people thinks of a fight, and where.
type Appraisal struct {
	mind.Appraisal
	Front  []int // the enemy's worlds within strike reach, nearest first
	Target int   // the world appraised, or -1
}

// strength is what c brings against e: its level at home, its miracles,
// its allies who would actually join, less what its other wars take.
func (w *World) strength(c, e *Civ) float64 {
	in := mind.StrengthInput{Mil: c.Mil, Bonus: c.warBonus()}
	for _, pid := range c.Pacts {
		p := w.Pacts[pid]
		if p.Over || p.Kind == Defensive && !c.Wars[e.ID] {
			continue
		}
		for _, mid := range p.Members {
			m := w.Civs[mid]
			if m != c && m.Active() && len(w.front(m, e)) > 0 {
				in.Allies = append(in.Allies, m.Mil)
			}
		}
	}
	for eid := range c.Wars {
		if eid != e.ID {
			in.OtherWars++
		}
	}
	return mind.Strength(in, w.Cfg.Tuning)
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
	in := mind.AppraiseInput{
		Strength: w.strength(c, e), Believed: mil, Spread: spread, EnemyBonus: e.warBonus(),
		AtHome: a.Target == e.Home, Risk: c.Dials.Risk, Speed: c.Speed,
		Weakened: e.Plagued || float64(w.Now-e.LastDark) < w.Cfg.Tuning.Appraise.DarkAge,
	}
	if i := c.Intel[e.ID]; i != nil {
		in.Grid = i.Grid
		if i.Star == a.Target {
			in.Relief = i.Relief
		}
	}
	for eid := range e.Wars {
		if eid != c.ID {
			in.OtherWars++
		}
	}
	_, in.Dist = w.nearest(c, a.Target)
	in.Prize = w.prize(c, a.Target)
	in.Loss = w.tradeLoss(c, e)
	a.Appraisal = mind.Appraise(in, w.Cfg.Tuning)
	return a
}

// prize is what a star is worth to a people that would take it: the yield
// there it could harness, as far as it wants that kind, and each rarity
// there it lacks.
func (w *World) prize(c *Civ, star int) float64 {
	if star < 0 {
		return 0
	}
	p := 0.0
	for _, id := range w.sourcesAt[star] {
		s := w.Sources[id]
		if s.Rarity {
			if !c.has(s.Key) {
				p += rarityWorth
			}
			continue
		}
		if !c.harnessed(s) {
			continue
		}
		for k := range s.Yield {
			p += min(s.Yield[k], c.Want[k])
		}
	}
	return p
}

// rarityWorth is what a rarity a people lacks counts for in the prize, in
// units of yield per tick.
const rarityWorth = 5

// nearest is c's holding nearest a star, and the distance.
func (w *World) nearest(c *Civ, star int) (int, float64) {
	best, bd := c.Home, w.G.Dist(c.Home, star)
	for _, s := range w.holdings(c) {
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
	return mind.Bar(mind.BarInput{Posture: c.posture(), Hates: c.hates(e), Grudge: c.Grudge[e.ID] > 0 || e.Embargo[c.ID], Aloft: c.Aloft}, w.Cfg.Tuning)
}

// explain logs a decision's reason under -ai.
func (w *World) explain(c *Civ, what string, d interface{ Why() string }) {
	if w.Cfg.TraceAI {
		w.log("[the %s, %s: %s]", c.Name, what, d.Why())
	}
}
