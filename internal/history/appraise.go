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

// strength is what c brings against e, in levels: its level at home, its
// miracles, its manned ships and those already out against e, its allies
// who would actually join, less what its other wars take.
func (w *World) strength(c, e *Civ) float64 {
	in := mind.StrengthInput{Mil: c.Mil, Bonus: c.warBonus(), Ships: w.standing(c) + w.bodyGuns(c) + w.outAgainst(c, e)} // the body is what a world fights with
	for _, pid := range c.Pacts {
		p := w.Pacts[pid]
		if p.Over || p.Kind == Defensive && !c.Wars[e.ID] {
			continue
		}
		for _, mid := range p.Members {
			m := w.Civs[mid]
			if m != c && m.Active() && len(w.front(m, e)) > 0 {
				in.Allies = append(in.Allies, w.levelOf(m))
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

// outAgainst is c's ships on campaign against e, kept.
func (w *World) outAgainst(c, e *Civ) int {
	n := 0
	for _, x := range w.fleetsOf(c) {
		if x.Kind == Campaign && x.Target == e.ID && !x.Returning && !x.LaidUp {
			n += x.Ships
		}
	}
	return n
}

// appraise estimates c's fight against e at a world: the nearest front
// world, or the given one, or e's home if nothing is in reach. An enemy
// not fathomed carries a level more of spread.
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
	if w.unfathomed(c, e) {
		spread++ // it cannot tell what they are or want
	}
	in := mind.AppraiseInput{
		Strength: w.strength(c, e), Believed: mil, Spread: spread, EnemyBonus: e.warBonus(),
		Risk: c.Dials.Risk, Speed: c.Speed, Wis: c.Wis,
		Weakened: w.plagued(e) || float64(w.Now-e.LastDark) < w.Cfg.Tuning.Appraise.DarkAge || e.Stiff >= w.Cfg.Tuning.Appraise.Stiff, // their ways have set; a fleet sent there would find the answer late
	}
	in.Ships, in.Guns, in.Relief = w.believeSky(c, e, a.Target)
	for eid := range e.Wars {
		if eid != c.ID {
			in.OtherWars++
		}
	}
	_, in.Dist = w.nearest(c, a.Target)
	in.Prize = w.prize(c, a.Target)
	in.Loss = w.tradeLoss(c, e)
	in.Slights = w.offence(c, e)
	in.Conqueror = c.posture() == mind.Conqueror
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
	return mind.Bar(mind.BarInput{
		Posture: c.posture(), Hates: c.hates(e), Grudge: c.Grudge[e.ID] > 0 || e.Embargo[c.ID], Aloft: c.Aloft, NoShips: w.standing(c) == 0 && w.bodyGuns(c) == 0, Wis: c.Wis,
		Claim: w.claims(c, e), Kin: w.kin(c, e) && !w.feud(c, e),
		Stiff: c.Stiff, Fought: c.Fought[e.ID] > 0, Sailed: c.Tally.Fleets > 0,
		Appetite: mind.Appetite(c.Appetite, w.Cfg.Tuning), Wary: c.Wary[e.ID], Rival: w.rival(c, e),
	}, w.Cfg.Tuning)
}

// rival says whether two peoples are old enemies: wars enough fought
// between them and a grudge standing on either side. The grudge's
// decay ends the rivalry when nothing renews it.
func (w *World) rival(c, e *Civ) bool {
	return c.Fought[e.ID] >= w.Cfg.Tuning.War.RivalWars && (c.Grudge[e.ID] > 0 || e.Grudge[c.ID] > 0)
}

// explain logs a decision's reason under -ai, and hands it to a
// watcher's hook (Config.Reason) when one is set.
func (w *World) explain(c *Civ, what string, d interface{ Why() string }) {
	if w.tracing() {
		w.reason(c, what, d.Why())
	}
}

// tracing says whether anyone is listening for reasons: -ai, or a hook.
func (w *World) tracing() bool { return w.Cfg.TraceAI || w.Cfg.Reason != nil }

// reason is one decision's reason, to the hook and under -ai to the log.
func (w *World) reason(c *Civ, what, why string) {
	if w.Cfg.Reason != nil {
		w.Cfg.Reason(w, c, what, why)
	}
	if w.Cfg.TraceAI {
		w.event(KReason, c, nil, -1, P{"what": what, "why": why})
	}
}
