package history

import (
	"worldgen/internal/names"
	"worldgen/internal/species"
)

// Contact happens when reach spheres overlap. What follows depends on stance
// traits and relative military: peace and trade, submission, war, and after
// war either extermination, enslavement or contraction.

func (w *World) contacts() {
	for i, a := range w.Civs {
		if !a.Active() {
			continue
		}
		for j := i + 1; j < len(w.Civs); j++ {
			b := w.Civs[j]
			if !b.Active() || (a.Reached[b.ID] && b.Reached[a.ID]) {
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
	if len(a.Met) == 1 || len(b.Met) == 1 || w.R.Float64() < 0.15 {
		w.log("The %s hear the %s across %.0f light years: a signal, then a conversation %.0f years to the answer. Neither can reach the other yet.", a.Name, b.Name, d, 2*d)
	}
	w.fact(FMet, a, b, -1)
	w.exchange(a, b)
	if a.Has("mindrider") || b.Has("mindrider") {
		w.infection(a, b) // an idea needs no ship
		return
	}
	if a.Species.Sub == species.Parasite || b.Species.Sub == species.Parasite {
		return
	}
	if w.consider(a, b) || w.consider(b, a) {
		return
	}
	if w.monster(a, b) || w.monster(b, a) {
		return
	}
	a.Trade[b.ID], b.Trade[a.ID] = true, true
	w.fact(FTrade, a, b, -1)
}

// primitives: an old civilisation finds a pre-atomic one. Returns true if
// something happened that counts as their meeting.
func (w *World) primitives(old, young *Civ) bool {
	switch {
	case old.hates(young) && w.R.Float64() < 0.3:
		old.Met[young.ID], young.Met[old.ID] = true, true
		w.log("The %s find the %s on %s before they have looked up, and scour the world clean. They are thorough.", old.Name, young.Name, young.HomeName)
		w.fact(FScoured, old, young, young.Home)
		w.Bio[young.Home] = BioSimple
		w.endCiv(young, Extinct, sprintf("were scoured from %s by the %s before they had looked up", young.HomeName, old.Name))
		return true
	case old.hostile() && !young.Has("swarming") && !young.Species.Is(species.Planetary) && w.R.Float64() < 0.3*(0.5+old.Dials.Greed):
		old.Met[young.ID], young.Met[old.ID] = true, true
		w.log("The %s find the %s on %s, still at the plough, and take them. There is no war to speak of.", old.Name, young.Name, young.HomeName)
		if old.Species.Sub == species.Parasite {
			w.ride(old, young)
		} else {
			w.enslave(old, young)
		}
		return true
	}
	if !old.Met[young.ID] {
		old.Met[young.ID] = true // one-sided: the old know, the young do not
		w.observe(old, young, young.Home, 0.2)
		w.log("The %s find the %s on %s, still young, and watch from orbit.", old.Name, young.Name, young.HomeName)
	}
	return false
}

// encounter is a first meeting between equals: the miracles and the
// postures that settle it without a council, then each side's council on
// the other, and trade if nobody strikes.
func (w *World) encounter(a, b *Civ, watched, heard bool, at int) {
	w.observe(a, b, b.Home, 0.5)
	w.observe(b, a, a.Home, 0.5)
	finder, found := a, b
	// the stronger side is the one with the initiative
	if b.Mil > a.Mil {
		a, b = b, a
	}
	if a.Species.Sub == species.Parasite || b.Species.Sub == species.Parasite {
		w.infection(a, b)
		return
	}
	gap := a.Mil - b.Mil
	switch {
	case a.miracle("chorus") && !b.miracle("chorus") && !b.Species.Is(species.Hive) && !b.Species.Is(species.Unconscious) && w.R.Float64() < 0.6:
		w.log("The %s find the %s, and speak. Within a generation the %s ask to be ruled.", a.Name, b.Name, b.Name)
		w.vassal(a, b)
		return
	case b.miracle("chorus") && !a.miracle("chorus") && !a.Species.Is(species.Hive) && !a.Species.Is(species.Unconscious) && w.R.Float64() < 0.6:
		w.log("The %s find the %s, and the %s speak. Within a generation the %s ask to be ruled.", a.Name, b.Name, b.Name, a.Name)
		w.vassal(b, a)
		return
	case a.miracle("unmaking") && b.hostile() && !b.miracle("unmaking"):
		w.log("The %s meet the %s and learn what they hold. There is no war. The %s bend the knee.", b.Name, a.Name, b.Name)
		w.vassal(a, b)
		return
	case b.Has("pacifist") && a.hostile() && gap >= 1 && !a.Has("pacifist") && w.inReach(a, b.Home):
		w.log("The %s find the %s, who will not fight. They are taken without a war.", a.Name, b.Name)
		w.enslave(a, b)
		return
	case b.Has("submissive") && a.hostile() && gap >= 2 && w.inReach(a, b.Home):
		w.log("The %s meet the %s, and seeing what they face, bend the knee. They are vassals now.", b.Name, a.Name)
		w.vassal(a, b)
		return
	}
	switch {
	case at >= 0 && heard:
		w.log("Ships of the %s come upon the %s at %s, and the long conversation across the dark has a face at last.", finder.Name, found.Name, w.star(at))
	case at >= 0:
		w.log("Ships of the %s come upon the %s at %s.", finder.Name, found.Name, w.star(at))
	case heard:
		w.log("The %s and the %s, who have heard each other for a long time, at last meet in the flesh.", a.Name, b.Name)
	case watched:
		w.log("The %s, long watched from orbit, look up and find the %s.", b.Name, a.Name)
	default:
		w.log("The %s and the %s find each other.", a.Name, b.Name)
	}
	w.fact(FMet, finder, found, at)
	w.exchange(a, b)
	if a.Wars[b.ID] {
		return // already at war by fleet; now there is a front
	}
	if w.consider(a, b) || w.consider(b, a) {
		return
	}
	if w.monster(a, b) || w.monster(b, a) {
		return // what is remembered of them is not traded with
	}
	a.Trade[b.ID], b.Trade[a.ID] = true, true
	w.fact(FTrade, a, b, -1)
	w.log("Slow messages cross the dark between the %s and the %s for generations, and then trade.", a.Name, b.Name)
	if (a.Faced["plague"] || b.Faced["plague"]) && w.R.Float64() < 0.3 {
		a.Plagued, b.Plagued = true, true
		w.log("Something crosses with the messages and the trade. Both the %s and the %s begin to sicken.", a.Name, b.Name)
	}
}

// infection: a parasite meets a host species. The parasite always opens;
// the host's Infection filter is its first response, and the war that
// follows is fought by conversion and burning.
func (w *World) infection(a, b *Civ) {
	p, h := a, b
	if p.Species.Sub != species.Parasite {
		p, h = b, a
	}
	if h.Species.Sub == species.Parasite || h.Species.Sub == species.Machine {
		w.log("The %s and the %s find each other, and find nothing in the other worth having.", a.Name, b.Name)
		return
	}
	if p.Has("mindrider") {
		w.log("The %s reach the %s. Within a generation the %s are in their heads.", p.Name, h.Name, p.Name)
	} else {
		w.log("The %s find the %s. Within a generation the %s are inside them.", p.Name, h.Name, p.Name)
	}
	out := w.face(h, "infection", 0)
	wr := w.declare(p, h, "infection")
	if wr == nil {
		return
	}
	hi := wr.side(h.ID)
	switch out {
	case Overcome:
		wr.Will[hi] += 1
	case Scarred:
		wr.Burn = true
		w.log("The %s cannot cut it out. They will burn what it takes.", h.Name)
	case Declined:
		wr.Will[hi] = 0
		w.log("The %s do not fight it. World by world, they are ridden.", h.Name)
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
	s.Morale -= 1
	w.log("The %s are enslaved by the %s. They keep %s and little else.", s.Name, m.Name, s.HomeName)
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
	if c.Free() || !c.Active() {
		return
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
		w.log("The %s, who held the %s, are gone. The question of freedom answers itself, one way or the other.", m.Name, c.Name)
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
		if w.Bio[t] == BioComplex && w.Owner[t] < 0 && w.Held[t] < 0 && t != w.G.Sol {
			sp := species.Generate(w.R, w.G.Stars[t].Mult)
			sp.Add("uplifted")
			sp.Made = "uplifted by the " + c.Name
			c.Uplifts++
			nc := w.spawnCiv(t, sp, c.ID, "")
			nc.Vassal = true
			nc.Seen = c.Declines
			for _, k := range knownOf(c) {
				if w.R.Float64() < 0.5 {
					nc.Known[k] = true
				}
			}
			w.forget(nc, 0.3)
			w.recompute(nc)
			w.log("The %s raise the %s from the beasts of %s. They are %s, and grateful, for now.", c.Name, nc.Name, w.star(t), sp.Describe())
			w.fact(FUplift, c, nc, t)
			w.inherit(nc, c, 1)
			return
		}
	}
}

// breed turns a slave people into something the master wants: a made species.
func (w *World) breed(m, s *Civ) {
	sp := s.Species.Branch()
	sp.Name = names.Civ(w.R)
	sp.Add("bred")
	sp.Made = "bred by the " + m.Name + " from the " + s.Name
	home := s.Home
	w.endCiv(s, Transformed, sprintf("were bred by the %s into something else", m.Name))
	s.Into = "the " + sp.Name
	nc := w.spawnCiv(home, sp, m.ID, "")
	nc.Seen = m.Declines
	w.log("The %s remake the %s into the %s: %s.", m.Name, s.Name, nc.Name, sp.Describe())
	w.fact(FBred, m, s, home)
	w.inherit(nc, s, 1)
}

func init() {
	def(&Filter{
		Key: "infection", Name: "Infection", Levels: []string{"sur"}, Diff: 5, Domain: "biology",
		Overcome: func(w *World, c *Civ) {
			w.log("The %s find the thing inside them and cut it out. Then they go looking for where it came from.", c.Name)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarQuarantine] = true
		},
		Decline: func(w *World, c *Civ) {},
	})
}
