package history

import (
	"worldgen/internal/species"
)

// Contact happens when reach spheres overlap. What follows depends on stance
// traits and relative military: peace and trade, submission, war, and after
// war either extermination, enslavement or contraction.

func (w *World) contacts() {
	for i, a := range w.Civs {
		if !a.Active() || a.Asleep {
			continue
		}
		for j := i + 1; j < len(w.Civs); j++ {
			b := w.Civs[j]
			if !b.Active() || b.Asleep || (a.Reached[b.ID] && b.Reached[a.ID]) {
				continue // a sleeper is all but impossible to contact
			}
			if pa, pb := w.perceives(a, b), w.perceives(b, a); !pa || !pb {
				// one side cannot hold the other in mind: the other finds it, and that is all the meeting there is
				if (w.touch(a, b) && (w.knowsOf(a, b) || w.knowsOf(b, a))) || w.hear(a, b) {
					if pa {
						w.notice(a, b)
					} else if pb {
						w.notice(b, a)
					}
				}
				continue
			}
			if !w.touch(a, b) {
				if !(a.Met[b.ID] && b.Met[a.ID]) && w.hear(a, b) {
					a.Met[b.ID], b.Met[a.ID] = true, true
					a.Tally.MetHeard++
					b.Tally.MetHeard++
					w.hearing(a, b)
				}
				continue
			}
			if !w.knowsOf(a, b) && !w.knowsOf(b, a) {
				continue // territories overlap, but neither has looked
			}
			a.Tally.MetTouch++
			b.Tally.MetTouch++
			w.meet(a, b, -1)
		}
	}
}

// knowsOf says whether a people has any idea b is there: heard, watched,
// or a holding of b read.
func (w *World) knowsOf(a, b *Civ) bool {
	if a.Met[b.ID] {
		return true
	}
	for _, s := range w.holdings(b) {
		if _, ok := a.Charted[s]; ok {
			return true
		}
	}
	return false
}

// meet is two peoples coming to know each other in the flesh: at a star
// where one came upon the other, or -1 for a border met by touch.
func (w *World) meet(a, b *Civ, at int) {
	if !a.Active() || !b.Active() || (a.Reached[b.ID] && b.Reached[a.ID]) {
		return
	}
	if pa, pb := w.perceives(a, b), w.perceives(b, a); !pa || !pb {
		if pa {
			w.notice(a, b)
		} else if pb {
			w.notice(b, a)
		}
		return
	}
	// a young species found by an old one is not a contact between equals
	// a people holding a miracle is nobody's primitive, whatever their era
	if young, old := a, b; (young.Era < 2 && len(young.held()) == 0) || (old.Era < 2 && len(old.held()) == 0) {
		if old.Era < 2 && len(old.held()) == 0 {
			young, old = b, a
		}
		if (young.Era >= 2 || len(young.held()) > 0) || old.Met[young.ID] {
			return // both young, or already watched
		}
		if !w.primitives(old, young) {
			return // watched from orbit; they will meet properly later
		}
	}
	watched := (a.Met[b.ID] || b.Met[a.ID]) && !(a.Met[b.ID] && b.Met[a.ID])
	heard := a.Met[b.ID] && b.Met[a.ID]
	a.Met[b.ID], b.Met[a.ID] = true, true
	a.Reached[b.ID], b.Reached[a.ID] = true, true
	w.encounter(a, b, watched, heard, at)
}

// touch says whether two peoples' territories overlap: a holding of one
// within the joint reach of a holding of the other.
func (w *World) touch(a, b *Civ) bool {
	r := a.Reach + b.Reach
	if r < 1 {
		return false
	}
	ha, hb := w.holdings(a), w.holdings(b)
	ea, eb := 0.0, 0.0
	for _, s := range ha {
		ea = max(ea, w.G.Dist(a.Home, s))
	}
	for _, s := range hb {
		eb = max(eb, w.G.Dist(b.Home, s))
	}
	if w.G.Dist(a.Home, b.Home) > r+ea+eb {
		return false
	}
	for _, sa := range ha {
		for _, sb := range hb {
			if w.G.Dist(sa, sb) <= r {
				return true
			}
		}
	}
	return false
}

// signal is how far a people's noise carries: radio at the atomic age,
// louder as it grows.
func (c *Civ) signal() float64 {
	switch {
	case c.Era >= 3:
		return 40
	case c.Era >= 2:
		return 20
	}
	return 0
}

// hear says whether two peoples can detect each other's signals: each
// must be loud enough to reach the other, from any holding.
func (w *World) hear(a, b *Civ) bool {
	if a.signal() == 0 || b.signal() == 0 {
		return false
	}
	for _, sa := range w.holdings(a) {
		for _, sb := range w.holdings(b) {
			if d := w.G.Dist(sa, sb); d <= a.signal() && d <= b.signal() {
				return true
			}
		}
	}
	return false
}

// hearing is a contact by signal only: each learns the other is there,
// and what it can, and the councils weigh a war of fleets.
func (w *World) hearing(a, b *Civ) {
	w.observe(a, b, b.Home, 0.8)
	w.observe(b, a, a.Home, 0.8)
	_, d := w.nearest(a, b.Home)
	told := len(a.Met) == 1 || len(b.Met) == 1 || w.R.Float64() < 0.15
	w.meeting(a, b, -1, "signal").with(P{"distance": d, "told": told})
	w.renew(a, 0.1) // a stranger is something new
	w.renew(b, 0.1)
	w.mirrored(a, b)
	w.fathomPair(a, b)
	if w.warm(a, b) {
		w.kinMeet(a, b)
		return
	}
	if w.consider(a, b) || w.consider(b, a) {
		return
	}
	if w.mutual(a, b) {
		w.openPair(a, b)
	}
}

// primitives: an old civilisation finds a pre-atomic one. Returns true if
// something happened that counts as their meeting.
func (w *World) primitives(old, young *Civ) bool {
	switch {
	case old.hates(young) && w.R.Float64() < 0.3:
		old.Met[young.ID], young.Met[old.ID] = true, true
		w.meeting(old, young, young.Home, "touch").with(P{"way": "scoured"})
		w.fact(FScoured, old, young, young.Home).with(P{"way": "primitives"})
		w.Bio[young.Home] = BioSimple
		w.endCiv(young, Extinct, because("scoured_young").At(young.Home).By(old))
		return true
	case old.hostile() && !young.Has("swarming") && !young.Species.Is(species.Planetary) && young.treats() && w.R.Float64() < 0.3*(0.5+old.Dials.Greed):
		old.Met[young.ID], young.Met[old.ID] = true, true
		w.meeting(old, young, young.Home, "touch").with(P{"way": "taken"})
		if old.Own >= 0 {
			w.ride(old, young)
		} else {
			w.enslave(old, young)
		}
		return true
	}
	if !old.Met[young.ID] {
		old.Met[young.ID] = true // one-sided: the old know, the young do not
		w.noticed(old, young, young.Home).with(P{"way": "watched"})
		w.observe(old, young, young.Home, 0.2)
	}
	return false
}

// encounter is a first meeting between equals: the miracles and the
// postures that settle it without a council, each side's roll to fathom
// the other, then each side's council on the other, and the tales and
// trade if nobody strikes and each understands the other.
func (w *World) encounter(a, b *Civ, watched, heard bool, at int) {
	w.observe(a, b, b.Home, 0.5)
	w.observe(b, a, a.Home, 0.5)
	w.fathomPair(a, b) // each rolls once at once, whatever follows
	w.expose(a, b, "landing")
	w.expose(b, a, "landing")
	finder, found := a, b
	met := w.meeting(finder, found, at, "touch") // whatever follows, they have met
	// the stronger side is the one with the initiative
	if b.Mil > a.Mil {
		a, b = b, a
	}
	met.with(P{"strong": a.ID, "heard": heard, "watched": watched})
	gap := a.Mil - b.Mil
	switch {
	case a.miracle("chorus") && !b.miracle("chorus") && !b.Species.Is(species.Hive) && !b.Species.Is(species.Unconscious) && b.treats() && w.R.Float64() < 0.6:
		met.P["way"] = "chorus"
		w.vassal(a, b)
		return
	case b.miracle("chorus") && !a.miracle("chorus") && !a.Species.Is(species.Hive) && !a.Species.Is(species.Unconscious) && a.treats() && w.R.Float64() < 0.6:
		met.P["way"] = "chorus_weak"
		w.vassal(b, a)
		return
	case a.miracle("unmaking") && b.hostile() && !b.miracle("unmaking") && b.treats():
		met.P["way"] = "unmaking"
		w.vassal(a, b)
		return
	case b.Has("pacifist") && a.hostile() && gap >= 1 && !a.Has("pacifist") && w.inReach(a, b.Home):
		met.P["way"] = "pacifist"
		w.enslave(a, b)
		return
	case b.Has("submissive") && a.hostile() && gap >= 2 && w.inReach(a, b.Home):
		met.P["way"] = "submissive"
		w.vassal(a, b)
		return
	}
	met.P["way"] = "equals"
	w.renew(a, 0.1) // a stranger is something new
	w.renew(b, 0.1)
	if !heard {
		w.mirrored(a, b)
	}
	if a.Wars[b.ID] {
		return // already at war by fleet; now there is a front
	}
	if w.warm(a, b) {
		w.kinMeet(a, b)
		return
	}
	if w.consider(a, b) || w.consider(b, a) {
		return
	}
	if w.mutual(a, b) {
		w.openPair(a, b) // the tales and the trade, when each understands the other
	}
}

func (w *World) enslave(m, s *Civ) {
	m.Ruled++
	s.Master, s.Vassal = m.ID, false
	s.Seen = m.Declines
	s.Voyages = nil
	delete(m.Wars, s.ID)
	delete(s.Wars, m.ID)
	// colonies pass to the master
	for _, x := range append([]int(nil), s.Systems...) {
		if x != s.Home {
			s.Systems = remove(s.Systems, x)
			w.Owner[x] = m.ID
			m.Systems = append(m.Systems, x)
		}
	}
	m.Peak = max(m.Peak, len(m.Systems))
	// the slave's fleets become the master's guards, where they stand
	for _, x := range w.fleetsOf(s) {
		w.reown(x, m)
		if x.Kind == Roam {
			x.Kind = Guard
		}
	}
	s.Morale -= 1
	w.fact(FEnslaved, m, s, s.Home)
}

func (w *World) vassal(m, s *Civ) {
	m.Ruled++
	s.Master, s.Vassal = m.ID, true
	s.Seen = m.Declines
	delete(m.Wars, s.ID)
	delete(s.Wars, m.ID)
	w.fact(FVassal, m, s, s.Home)
}

// revolt: slaves and vassals watch their master. A master's decline is the
// slaves' chance; a master's death forces the question.
func (w *World) revolt(c *Civ) {
	if c.Free() || !c.Active() || w.ridden(c) {
		return // a ridden people's rising is the cure contest; see plague.go
	}
	m := w.Civs[c.Master]
	if m.Living() && m.Declines == c.Seen {
		return
	}
	c.Seen = m.Declines
	adj := 0.0
	if c.Vassal {
		adj = -1
	}
	if !m.Living() {
		adj -= 1
		w.event(KMasterGone, m, c, -1, P{})
	}
	w.face(c, "revolt", adj+w.holdDiff(c))
}

// uplift: a strong civilisation makes a new species from complex life
// within its reach. The client relationship goes the way of vassalage.
func (w *World) uplift(c *Civ) {
	if c.Aloft {
		return
	}
	inclined := c.Has("curious") || c.Has("collective") || c.Has("contemplative")
	p := 0.0001
	if c.Species.Sub == species.Parasite {
		inclined, p = true, 0.0003 // a rider makes riders
	}
	if !c.Active() || c.Era < 3 || c.Soc < 5 || !c.Free() || c.Uplifts >= 2 || !inclined || !w.chance(p) {
		return
	}
	for _, t := range w.G.Near(c.Home, c.Reach) {
		if w.Bio[t] == BioComplex && w.Owner[t] < 0 && t != w.G.Sol {
			sp := species.Generate(w.R, w.G.Stars[t].Mult)
			sp.Add("uplifted")
			sp.Made = species.MadeBy("uplifted", c.ID)
			c.Uplifts++
			nc := w.spawnCiv(t, sp, c.ID)
			nc.Vassal = true
			nc.Seen = c.Declines
			for _, k := range knownOf(c) {
				if w.R.Float64() < 0.5 {
					nc.Known[k] = true
				}
			}
			w.forget(nc, 0.3)
			w.recompute(nc)
			at := w.slot() // the raising is told before the morality it gives, and recorded after
			w.upliftMorality(nc, c)
			w.fill(at, w.unplaced(FUplift, c, nc, t).with(P{"traits": sp.TraitKeys()}))
			w.inherit(nc, c, 1)
			return
		}
	}
}

// breed turns a slave people into something the master wants: a made species.
func (w *World) breed(m, s *Civ) {
	sp := s.Species.Branch()
	sp.Add("bred")
	sp.Made = species.Making{Key: "bred_from", By: m.ID, From: s.ID, Legacy: -1, Plague: -1}
	home := s.Home
	w.endCiv(s, Transformed, because("bred_into").By(m))
	s.Into, s.IntoCivs = "species", []int{sp.ID}
	nc := w.spawnCiv(home, sp, m.ID)
	nc.Seen = m.Declines
	w.fact(FBred, m, s, home).with(P{"into": nc.ID, "traits": sp.TraitKeys()})
	w.inherit(nc, s, 1)
}
