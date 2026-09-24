package history

import (
	"sort"

	"worldgen/internal/battle"
	"worldgen/internal/mind"
)

// Garrisons: the council places ships as it places everything, by a
// score. Each holding has a want in ships, from the strongest fleet a
// hostile or unknown neighbour in reach is believed to have, counted in
// the people's own ships; ships move from where they exceed the want to
// where they fall short, at ship speed, one order per council. A campaign
// is drawn from the guard at the holding nearest the target if it has the
// ships, else the council orders ships from other guards to that holding
// and the campaign launches when they have arrived: the muster.

// Muster is a campaign gathering at a holding: the council's standing
// order until the guard there has the ships, or the odds go.
type Muster struct {
	Star   int    // where the ships gather
	Ships  int    // the campaign's ships
	Target int    // the enemy
	World  int    // the world the campaign is for
	Cause  reason // why the war will be declared when the fleet is gathered
	Since  Year
	War    int // the war it gathers for, or -1 for a war to be declared when it sails
}

// garrison is the civ step after the council: the muster it ordered, the
// wants of the holdings, and one move of the guards toward where they are
// wanted, at the council's cadence. A horde's fleets are not placed.
func (w *World) garrison(c *Civ) {
	if !c.Active() || c.Starfaring == 0 {
		return
	}
	w.musterStep(c)
	if c.Aloft {
		return
	}
	hs := w.holdingsOf(c)
	g := mind.Garrisons(mind.GarrisonInput{Holdings: hs, Fear: c.Dials.Fear}, w.Cfg.Tuning)
	c.GarrisonWant = g.Total
	if c.Muster != nil || !w.chance(w.Cfg.Tuning.Council.Cadence) {
		return
	}
	w.explain(c, "placing the guards", g)
	if g.From < 0 {
		// a guard stranded where nothing is wanted goes home
		for i, h := range hs {
			if w.Owner[h.Star] != c.ID && h.Ships > 0 && g.Wants[i] == 0 {
				w.goHome(w.guardAt(c, h.Star))
			}
		}
		return
	}
	from, to := hs[g.From], hs[g.To]
	w.sendGuard(c, w.guardAt(c, from.Star), to.Star, g.Ships)
	c.Tally.Garrisons++
	w.event(KGarrisoned, c, nil, to.Star, P{"ships": g.Ships})
}

// holdingsOf is a people's holdings as the garrison policy sees them, and
// any star a guard of its sits at that is not one.
func (w *World) holdingsOf(c *Civ) []mind.Holding {
	threats := w.threats(c)
	var out []mind.Holding
	seen := map[int]bool{}
	add := func(s int) {
		if seen[s] {
			return
		}
		seen[s] = true
		h := mind.Holding{Star: s, Home: s == c.Home, Threat: threats[s], Guns: w.gunsAt(c, s), Coming: w.comingTo(c, s)}
		if g := w.guardAt(c, s); g != nil && !g.LaidUp {
			h.Ships = g.Ships
		}
		out = append(out, h)
	}
	for _, s := range c.Systems {
		add(s)
	}
	for _, x := range w.fleetsOf(c) {
		if x.Kind == Guard && x.Base >= 0 {
			add(x.Base)
		}
	}
	return out
}

// threats is the strongest fleet a hostile or unknown neighbour whose
// reach covers each of a people's worlds is believed to have, in the
// people's own ships: their believed ships at their believed quality over
// ours. A neighbour is hostile when it strikes first, is at war with the
// people, or has never been looked at. A fleet in a world's sky is a
// threat to that world that needs no belief.
func (w *World) threats(c *Civ) map[int]float64 {
	out := map[int]float64{}
	q := w.quality(c)
	for _, x := range w.liveFleets() {
		if x.Kind != Campaign || x.Target != c.ID || x.Base < 0 || x.LaidUp || x.Ships <= 0 || w.Owner[x.Base] != c.ID {
			continue
		}
		out[x.Base] = max(out[x.Base], battle.Strength(x.Ships, w.quality(w.Civs[x.Owner]))/q)
	}
	for _, eid := range metOf(c) {
		e := w.Civs[eid]
		if !e.Active() || e.ID == c.ID || e.Master == c.ID || c.Master == e.ID || w.allied(c, e) {
			continue
		}
		if !(e.hostile() || c.Wars[eid] || c.Intel[eid] == nil) {
			continue
		}
		mil, _ := w.believe(c, e)
		ships := w.believeShips(c, e) * battle.Quality(mil+e.warBonus()) / q
		for _, s := range c.Systems {
			if w.inReach(e, s) {
				out[s] = max(out[s], ships)
			}
		}
	}
	return out
}

// comingTo is the ships of a people's guards on their way to a star.
func (w *World) comingTo(c *Civ, star int) int {
	n := 0
	for _, x := range w.fleetsOf(c) {
		if x.Kind == Guard && x.Base < 0 && x.Star == star {
			n += x.Ships
		}
	}
	return n
}

// sendGuard sends n ships of a guard to another star, at ship speed: the
// whole guard if that is all of it, else a fleet split off it. A guard
// lands into the guard where it arrives; a horde's fleet lands as a base.
func (w *World) sendGuard(c *Civ, g *Expedition, to, n int) *Expedition {
	if g == nil || n <= 0 || g.Base < 0 {
		return nil
	}
	n = min(n, g.Ships)
	x := g
	if n < g.Ships || g.Kind == Roam {
		x = &Expedition{ID: len(w.Expeditions), Owner: c.ID, Target: -1, Kind: g.Kind, Star: to, From: g.Base, Ships: n, Back: -1, Contract: -1, SoldBy: -1,
			Launched: w.Now, Out: w.Now, Base: g.Base, Fed: w.Now, Manned: w.Now, Seen: map[int]bool{}}
		g.Ships -= n
		w.addExpedition(x)
	}
	w.sail(x, to)
	x.Returning = x.Kind == Guard
	return x
}

// guardWith is a people's guard nearest a star with at least n ships
// manned, or nil.
func (w *World) guardWith(c *Civ, star, n int) *Expedition {
	var best *Expedition
	bd := 0.0
	for _, x := range w.fleetsOf(c) {
		if !x.atBase() || x.LaidUp || x.Ships < n {
			continue
		}
		if d := w.G.Dist(x.Base, star); best == nil || d < bd {
			best, bd = x, d
		}
	}
	return best
}

// muster orders a campaign of n ships to gather at the holding nearest
// its target: the guards nearest the holding send what they have until
// the count is covered. The war, if none runs, is declared when the
// fleet sails: a muster that stands down has started nothing.
func (w *World) muster(c, e *Civ, world int, cause reason, n int) {
	star, _ := w.nearest(c, world)
	c.Muster = &Muster{Star: star, Ships: n, Target: e.ID, World: world, Cause: cause, Since: w.Now, War: -1}
	if wr := w.warBetween(c.ID, e.ID); wr != nil {
		c.Muster.War = wr.ID
	}
	c.Tally.Musters++
	have := 0
	if g := w.guardAt(c, star); g != nil && !g.LaidUp {
		have = g.Ships
	}
	w.gather(c, star, n-have-w.comingTo(c, star))
	w.event(KGathered, c, nil, star, P{})
}

// gather sends n ships to a star from the people's other guards, nearest
// first.
func (w *World) gather(c *Civ, star, n int) {
	var guards []*Expedition
	for _, x := range w.fleetsOf(c) {
		if x.atBase() && !x.LaidUp && x.Ships > 0 && x.Base != star {
			guards = append(guards, x)
		}
	}
	sort.SliceStable(guards, func(i, j int) bool { return w.G.Dist(guards[i].Base, star) < w.G.Dist(guards[j].Base, star) })
	for _, g := range guards {
		if n <= 0 {
			return
		}
		k := min(g.Ships, n)
		w.sendGuard(c, g, star, k)
		n -= k
	}
}

// musterStep is a muster's tick: it launches when the guard has the
// ships, is re-sized while it waits, and stands down when the odds go,
// when nothing more is coming, or when it has waited too long.
func (w *World) musterStep(c *Civ) {
	m := c.Muster
	if m == nil {
		return
	}
	e := w.Civs[m.Target]
	have := 0
	if g := w.guardAt(c, m.Star); g != nil && !g.LaidUp {
		have = g.Ships
	}
	coming := w.comingTo(c, m.Star)
	holds := false
	over := m.War >= 0 && w.Wars[m.War].Over // gathered for a war that has ended: it declares no new one (step 11's batches, seed 12: a muster outliving each war it was called in opened the next, twenty-nine times)
	if !over && e.Active() && e.Free() && w.holds(e, m.World) && (w.warBetween(c.ID, e.ID) != nil || c.Truce[e.ID] <= w.Now) {
		if k := w.sizeAt(c, e, m.World); k.Send {
			holds = true
			m.Ships = k.Share
		}
	}
	ch := mind.Muster(mind.MusterInput{Have: have, Need: m.Ships, Coming: coming, Age: float64(w.Now - m.Since), Holds: holds}, w.Cfg.Tuning)
	w.explain(c, "mustering at "+w.star(m.Star), ch)
	switch {
	case ch.Launch:
		c.Muster = nil
		if w.warBetween(c.ID, e.ID) == nil && w.openWar(c, e, m.Cause, m.World) == nil {
			return
		}
		w.launch(c, Campaign, e, m.World, m.Ships)
	case ch.StandDown:
		c.Muster = nil
		if w.Cfg.TraceAI {
			w.event(KDebug, c, nil, m.Star, P{"text": sprintf("[the %s stand down at %s]", c.Tok(), w.star(m.Star))})
		}
	}
}
