package history

import (
	"worldgen/internal/mind"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// Nomads are a people that takes to the sky when it can and lives as
// fleets, with no worlds. A fleet has ships and a base, any star; a star
// feeds it for a few thousand years and then it must move on, which is
// the engine of wandering. The fleets are the people: they keep their
// numbers in hulls and feed them by grazing, and a horde that cannot meet
// the keep lays ships up, which for a horde is thinning. They split and
// merge, carry what they know between the settled, and when they win a
// world they strip it: its ships and its people join the horde. They
// cannot be enslaved, only broken fleet by fleet by a campaign sent to a
// base, on the battle rule, and they never capitulate; their peace is
// leaving. A few come to rest.

// nomad says whether a people has the way and has not settled for good.
func (c *Civ) nomad() bool { return c.Has("nomadic") && !c.Rested }

// fleets are a people's roaming fleets.
func (w *World) fleets(c *Civ) []*Expedition {
	var out []*Expedition
	for _, x := range w.fleetsOf(c) {
		if x.Kind == Roam {
			out = append(out, x)
		}
	}
	return out
}

// holdings are the stars a people acts from: its worlds, and for the aloft
// the bases of its fleets.
func (w *World) holdings(c *Civ) []int {
	if !c.Aloft {
		return c.Systems
	}
	var out []int
	for _, x := range w.fleets(c) {
		if x.Base >= 0 {
			out = append(out, x.Base)
		}
	}
	return out
}

// aWorld is one of a people's worlds at random, or for the aloft one of its
// bases, or its seat if it has neither.
func (w *World) aWorld(c *Civ) int {
	if h := w.holdings(c); len(h) > 0 {
		return w.pick(h)
	}
	return c.Home
}

// wander is the tick step that sends a nomad people to the sky when it can.
func (w *World) wander(c *Civ) {
	if !c.nomad() || c.Aloft || !c.Free() {
		return
	}
	if c.Reach >= 10 || (c.Dying && c.Reach >= 1) {
		w.takeSky(c, reason{})
	}
}

// takeSky turns a settled nomad people into fleets and empties its
// worlds: the guards become the horde, and a world with no guard sends
// one ship of what it had. A people with no ships cannot go.
func (w *World) takeSky(c *Civ, why reason) bool {
	if c.Aloft || len(c.Systems) == 0 {
		return false
	}
	worlds := append([]int(nil), c.Systems...)
	c.Aloft = true
	c.Dying = false
	c.Voyages = nil
	w.aloftGuards(c)
	for _, s := range worlds {
		if w.guardAt(c, s) == nil {
			w.addGuard(c, s, 1)
		}
		w.loseSystem(c, s, "empty_cradle", reason{})
	}
	c.Record = append(c.Record, Record{Kind: "sky", Legacy: -1})
	w.told(FExodus, c, nil, c.Home).with(P{"way": "sky"}).with(why.params("why"))
	w.seat(c)
	w.recompute(c)
	return true
}

// aloftGuards turns a people's guards into a horde's fleets, wherever
// they are, and its fleets in flight come home to the horde.
func (w *World) aloftGuards(c *Civ) {
	for _, x := range w.fleetsOf(c) {
		if x.Kind == Guard {
			x.Kind = Roam
			x.Fed = w.Now
		}
	}
}

// flee is a settled people whose last world is gone taking to the sky with
// what escaped: refugees under the nomad rules, without the way. The guard
// at the lost world is what got away, and one ship at the least. Returns
// false if nothing could get away.
func (w *World) flee(c *Civ, lost int, cause reason) bool {
	if c.Reach < 1 || !c.Species.Profile().Can(species.Flees) || c.Aloft {
		return false
	}
	ships := 1
	if g := w.guardAt(c, lost); g != nil {
		ships = max(1, g.Ships)
		g.Over = true
	}
	base := lost
	for _, t := range w.G.Near(lost, min(max(c.Reach, 3), 20)) {
		if w.Owner[t] < 0 {
			base = t
			break
		}
	}
	c.Aloft = true
	c.Dying = false
	c.Voyages = nil
	w.aloftGuards(c)
	w.addGuard(c, base, ships)
	c.Record = append(c.Record, Record{Kind: "sky", Legacy: -1})
	c.Morale -= 1
	w.fact(FExodus, c, nil, lost).with(P{"way": "fled", "base": base}).with(cause.params("why"))
	w.seat(c)
	w.recompute(c)
	return true
}

// greatestFleet is a nomad people's strongest fleet, or nil.
func (w *World) greatestFleet(c *Civ) *Expedition {
	var best *Expedition
	for _, x := range w.fleets(c) {
		if best == nil || x.Ships > best.Ships {
			best = x
		}
	}
	return best
}

// seat keeps a nomad people's seat where its greatest fleet is, and its
// mobile rarities aboard it.
func (w *World) seat(c *Civ) {
	best := w.greatestFleet(c)
	if best == nil {
		return
	}
	w.stow(c, best)
	s := best.Base
	if s < 0 {
		s = best.Star
	}
	if s != c.Home {
		c.Home = s
	}
}

// roam is a nomad people's tick: fleets merge, feed, thin, move on and
// split; the seat moves; what they know moves with them. The horde grows
// at its docks, which are its fleets at base, and thins by the keep it
// cannot pay; a laid-up fleet stays where it is.
func (w *World) roam(c *Civ) {
	fl := w.fleets(c)
	if len(fl) == 0 {
		w.endCiv(c, Extinct, because("last_fleet"))
		return
	}
	// merge at a shared base
	for i := 0; i < len(fl); i++ {
		for j := i + 1; j < len(fl); j++ {
			if fl[i].Base >= 0 && fl[i].Base == fl[j].Base && !fl[j].Over {
				fl[i].Ships += fl[j].Ships
				fl[i].LaidUp = fl[i].LaidUp && fl[j].LaidUp
				fl[j].Over = true
			}
		}
	}
	fl = w.fleets(c)
	hop := mind.RoamHop(c.Reach, w.Cfg.Tuning)
	for _, x := range fl {
		if x.Base < 0 || x.LaidUp {
			continue // in flight, or not kept
		}
		// a star is grazed out in a few thousand years, and a fleet that
		// cannot move on thins
		if float64(w.Now-x.Fed) > 4000 {
			if t := w.nextStar(c, x, hop); t >= 0 {
				w.moveFleet(c, x, t)
				continue
			}
			x.Ships -= w.count(0.03 * float64(x.Ships))
		}
		if x.Ships <= 0 {
			x.Over = true
			w.fleetLost(x, x.Base)
			continue
		}
		if x.Ships > 4 && len(fl) < 8 && w.chance(0.1) {
			if t := w.nextStar(c, x, hop); t >= 0 {
				nx := &Expedition{ID: len(w.Expeditions), Owner: c.ID, Target: -1, Kind: Roam, Star: t, From: x.Base, Ships: x.Ships / 2, Back: -1, Contract: -1, SoldBy: -1,
					Launched: w.Now, Out: w.Now, Base: -1, Fed: w.Now, Manned: w.Now, Seen: map[int]bool{}}
				x.Ships -= nx.Ships
				nx.Arrive = w.Now + Year(w.G.Dist(x.Base, t)*c.Speed)
				w.addExpedition(nx)
			}
		}
	}
	if len(w.fleets(c)) == 0 {
		w.endCiv(c, Extinct, because("last_fleet"))
		return
	}
	w.seat(c)
	w.carry(c, hop)
	// refugees want a home; the way does not
	if !c.Has("nomadic") && w.chance(0.02) {
		w.rest(c, because("road"))
	}
}

// nextStar is where a fleet goes next: a star within a hop, not the one it
// is at, not held by a monster, an unowned one or a trade partner's by
// preference.
func (w *World) nextStar(c *Civ, x *Expedition, hop float64) int {
	var ports []mind.Port
	for _, t := range w.G.Near(x.Base, hop) {
		o := w.Owner[t]
		ports = append(ports, mind.Port{ID: t, Held: w.lurks(t), Good: o < 0 || c.Trade[o]})
	}
	return mind.NextStar(ports, w.R)
}

func (w *World) moveFleet(c *Civ, x *Expedition, t int) {
	w.sail(x, t)
	if w.R.Float64() < 0.02 {
		w.event(KMovedOn, c, nil, t, P{"fleet": x.ID})
	}
}

// carry is what nomads do to the tree of everyone they trade with: a node
// passes now and then from one to the other.
func (w *World) carry(c *Civ, hop float64) {
	for _, eid := range sortedInts(c.Trade) {
		e := w.Civs[eid]
		if !e.Active() || !w.chance(0.003) {
			continue
		}
		near := false
		for _, b := range w.holdings(c) {
			if _, d := w.nearest(e, b); d <= hop {
				near = true
				break
			}
		}
		if !near {
			continue
		}
		from, to := c, e
		if w.R.Float64() < 0.5 {
			from, to = e, c
		}
		var cands []string
		for _, k := range knownOf(from) {
			n := tech.Get(k)
			if !to.Known[k] && !n.Miracle && w.canPursue(to, n) {
				cands = append(cands, k)
			}
		}
		if len(cands) == 0 {
			continue
		}
		k := cands[w.R.IntN(len(cands))]
		w.learn(to, tech.Get(k), false)
		if from == c {
			w.event(KCarried, c, e, -1, P{"node": k, "way": "brought"})
		} else {
			w.event(KCarried, c, e, -1, P{"node": k, "way": "learned"})
		}
	}
}

// strip is a nomad victory at a world: its ships and people join the horde
// and the world is left empty. The guard there, if any, changes hands
// whole; the people are one ship more.
func (w *World) strip(wr *War, c, e *Civ, t int) {
	i := wr.side(c.ID)
	share := 1
	if g := w.guardAt(e, t); g != nil {
		share += g.Ships
		g.Over = true
	}
	home := t == e.Home
	c.Loot.Add(w.yieldAt(e, t)) // the rest, once
	w.loseSystem(e, t, "horde", because("horde").By(c))
	w.addGuard(c, t, share)
	wr.Taken[i]++
	wr.Lost[1-i]++
	c.Tally.Taken++
	e.Tally.Lost++
	e.Morale -= 0.5
	wr.Will[i] += 0.3
	wr.Will[1-i] -= 0.3
	w.told(FStripped, c, e, t).with(P{"home": home})
	if wr.Named < 0 {
		wr.Named = t
	}
	if e.Active() && (wr.Lost[1-i] == 1 || wr.Lost[1-i]%3 == 0) {
		w.face(e, "hold", 0)
	}
	if !e.Active() {
		w.endWar(wr, "destroyed")
	}
}

// rest is a nomad people settling for good, usually after a scar. Its
// fleets become the guard of the world it rests at.
func (w *World) rest(c *Civ, why reason) {
	hop := min(max(c.Reach, 3), 20)
	var best *Expedition
	for _, x := range w.fleets(c) {
		if x.Base >= 0 && (best == nil || x.Ships > best.Ships) {
			best = x
		}
	}
	if best == nil {
		return
	}
	t := -1
	if w.Owner[best.Base] < 0 && w.canLive(c, best.Base) {
		t = best.Base
	} else {
		for _, s := range w.G.Near(best.Base, hop) {
			if w.Owner[s] < 0 && w.canLive(c, s) {
				t = s
				break
			}
		}
	}
	if t < 0 {
		return
	}
	c.Aloft, c.Rested = false, true
	for _, x := range w.fleets(c) {
		if x.Base >= 0 {
			w.mergeInto(x, t)
		} else {
			x.Kind = Guard // in flight: it lands where it was going and finds its way to the guard
			x.Returning, x.Star = true, t
		}
	}
	c.Systems = []int{t}
	w.Owner[t] = c.ID
	c.Home = t
	c.Record = append(c.Record, Record{Kind: "rest", Legacy: -1})
	w.takeOver(c, t)
	w.recompute(c)
	rest := w.unplaced(FRest, c, nil, t).with(P{"nomad": c.Has("nomadic")}).with(why.params("why"))
	w.wakeReservoir(c, t)
	w.place(rest)
}
