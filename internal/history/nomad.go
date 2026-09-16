package history

import (
	"worldgen/internal/names"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// Nomads are a people that takes to the sky when it can and lives as
// fleets, with no worlds. A fleet has a strength and a base, any star; a
// star feeds it for a few thousand years and then it must move on, which is
// the engine of wandering. The fleets are the people's Military, split and
// merge, carry what they know between the settled, and when they win a
// world they strip it: its share of the loser's strength becomes a fleet of
// theirs. They cannot be enslaved, only broken fleet by fleet, and they
// never capitulate; their peace is leaving. A few come to rest.

// nomad says whether a people has the way and has not settled for good.
func (c *Civ) nomad() bool { return c.Has("nomadic") && !c.Rested }

// fleets are a people's roaming fleets.
func (w *World) fleets(c *Civ) []*Expedition {
	var out []*Expedition
	for _, x := range w.Expeditions {
		if !x.Over && x.Kind == Roam && x.Owner == c.ID {
			out = append(out, x)
		}
	}
	return out
}

// ships is the strength of all a people's fleets.
func (w *World) ships(c *Civ) float64 {
	s := 0.0
	for _, x := range w.fleets(c) {
		s += x.Mil
	}
	return s
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
		w.takeSky(c, "")
	}
}

// takeSky turns a settled nomad people into fleets and empties its worlds.
func (w *World) takeSky(c *Civ, why string) {
	if c.Aloft || len(c.Systems) == 0 {
		return
	}
	total := max(1, c.Mil+c.Away)
	worlds := append([]int(nil), c.Systems...)
	each := total / float64(len(worlds))
	c.Aloft = true
	c.Dying = false
	c.Voyages = nil
	for _, s := range worlds {
		x := &Expedition{ID: len(w.Expeditions), Owner: c.ID, Target: -1, Kind: Roam, Star: s, From: s, Mil: each,
			Launched: w.Now, Arrive: w.Now, Base: s, Fed: w.Now, Seen: map[int]bool{}}
		w.Expeditions = append(w.Expeditions, x)
		w.loseSystem(c, s, "empty cradle of the "+c.Name, "")
	}
	c.Record = append(c.Record, "took to the sky")
	w.fact(FExodus, c, nil, c.Home)
	if why == "" {
		w.log("The %s take to the sky. %s is left empty behind them, and everything they are is in the fleets now.", c.Name, c.HomeName)
	} else {
		w.log("The %s take to the sky rather than %s. %s is left empty behind them.", c.Name, why, c.HomeName)
	}
	w.seat(c)
	w.recompute(c)
}

// flee is a settled people whose last world is gone taking to the sky with
// what escaped: refugees under the nomad rules, without the way. Returns
// false if nothing could get away.
func (w *World) flee(c *Civ, lost int, cause string) bool {
	if c.Reach < 1 || c.Species.Kind == species.PlanetaryMind || c.Aloft {
		return false
	}
	strength := max(0.5, 0.3*(c.Mil+c.Away))
	base := lost
	for _, t := range w.G.Near(lost, min(max(c.Reach, 3), 20)) {
		if w.Owner[t] < 0 && w.Held[t] < 0 {
			base = t
			break
		}
	}
	c.Aloft = true
	c.Dying = false
	c.Voyages = nil
	c.Away = 0
	x := &Expedition{ID: len(w.Expeditions), Owner: c.ID, Target: -1, Kind: Roam, Star: base, From: lost, Mil: strength,
		Launched: w.Now, Arrive: w.Now, Base: base, Fed: w.Now, Seen: map[int]bool{}}
	w.Expeditions = append(w.Expeditions, x)
	c.Record = append(c.Record, "took to the sky")
	c.Morale -= 1
	w.log("The %s %s. What got away is a fleet at %s, and it is all of them now.", c.Name, cause, w.star(base))
	w.fact(FExodus, c, nil, lost)
	w.seat(c)
	w.recompute(c)
	return true
}

// seat keeps a nomad people's seat where its greatest fleet is.
func (w *World) seat(c *Civ) {
	var best *Expedition
	for _, x := range w.fleets(c) {
		if best == nil || x.Mil > best.Mil {
			best = x
		}
	}
	if best == nil {
		return
	}
	s := best.Base
	if s < 0 {
		s = best.Star
	}
	if s != c.Home {
		c.Home = s
		c.HomeName = w.star(s)
	}
}

// roam is a nomad people's tick: fleets merge, feed, grow, thin, move on
// and split; the seat moves; what they know moves with them.
func (w *World) roam(c *Civ) {
	fl := w.fleets(c)
	if len(fl) == 0 {
		w.endCiv(c, Extinct, "lost the last of their fleets")
		return
	}
	// merge at a shared base
	for i := 0; i < len(fl); i++ {
		for j := i + 1; j < len(fl); j++ {
			if fl[i].Base >= 0 && fl[i].Base == fl[j].Base && !fl[j].Over {
				fl[i].Mil += fl[j].Mil
				fl[j].Over = true
			}
		}
	}
	fl = w.fleets(c)
	hop := min(max(c.Reach, 3), 20)
	cap := c.Quality + 2
	for _, x := range fl {
		if x.Base < 0 {
			continue // in flight
		}
		// a fleet feeds where it is and on the way; a star is grazed out in a
		// few thousand years, and a fleet that cannot move on thins
		if w.ships(c) < cap {
			x.Mil += 0.02 * w.dt * max(1, c.Quality/5)
		}
		if float64(w.Now-x.Fed) > 4000 {
			if t := w.nextStar(c, x, hop); t >= 0 {
				w.moveFleet(c, x, t)
				continue
			}
			x.Mil *= 1 - 0.03*w.dt
		}
		if x.Mil < 0.3 {
			x.Over = true
			continue
		}
		if x.Mil > 4 && len(fl) < 8 && w.chance(0.1) {
			if t := w.nextStar(c, x, hop); t >= 0 {
				nx := &Expedition{ID: len(w.Expeditions), Owner: c.ID, Target: -1, Kind: Roam, Star: t, From: x.Base, Mil: x.Mil / 2,
					Launched: w.Now, Base: -1, Fed: w.Now, Seen: map[int]bool{}}
				x.Mil /= 2
				nx.Arrive = w.Now + Year(w.G.Dist(x.Base, t)*c.Speed)
				w.Expeditions = append(w.Expeditions, nx)
			}
		}
	}
	if len(w.fleets(c)) == 0 {
		w.endCiv(c, Extinct, "lost the last of their fleets")
		return
	}
	w.seat(c)
	w.carry(c, hop)
	// refugees want a home; the way does not
	if !c.Has("nomadic") && w.chance(0.02) {
		w.rest(c, "the road")
	}
}

// nextStar is where a fleet goes next: a star within a hop, not the one it
// is at, not held by a horror, an unowned one or a trade partner's by
// preference.
func (w *World) nextStar(c *Civ, x *Expedition, hop float64) int {
	var good, any []int
	for _, t := range w.G.Near(x.Base, hop) {
		if w.Held[t] >= 0 {
			continue
		}
		o := w.Owner[t]
		if o < 0 || c.Trade[o] {
			good = append(good, t)
		} else if !c.Wars[o] || true {
			any = append(any, t)
		}
	}
	if len(good) > 0 {
		return good[w.R.IntN(len(good))]
	}
	if len(any) > 0 {
		return any[w.R.IntN(len(any))]
	}
	return -1
}

func (w *World) moveFleet(c *Civ, x *Expedition, t int) {
	x.From, x.Star = x.Base, t
	x.Base = -1
	x.Arrive = w.Now + Year(w.G.Dist(x.From, t)*c.Speed)
	x.Launched = w.Now
	if w.R.Float64() < 0.02 {
		w.log("The fleets of the %s move on, to %s.", c.Name, w.star(t))
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
			w.log("The fleets of the %s bring the %s %s.", c.Name, e.Name, tech.Get(k).Name)
		} else {
			w.log("The %s learn %s from the %s, and carry it on.", c.Name, tech.Get(k).Name, e.Name)
		}
	}
}

// strip is a nomad victory at a world: its ships and people join the horde
// and the world is left empty.
func (w *World) strip(wr *War, c, e *Civ, t int) {
	i := wr.side(c.ID)
	share := max(0.5, e.Mil/float64(max(1, len(e.Systems))))
	home := t == e.Home
	w.loseSystem(e, t, "stripped by the horde", sprintf("were swallowed by the horde of the %s", c.Name))
	x := &Expedition{ID: len(w.Expeditions), Owner: c.ID, Target: -1, Kind: Roam, Star: t, From: t, Mil: share,
		Launched: w.Now, Arrive: w.Now, Base: t, Fed: w.Now, Seen: map[int]bool{}}
	w.Expeditions = append(w.Expeditions, x)
	wr.Taken[i]++
	wr.Lost[1-i]++
	c.Tally.Taken++
	e.Tally.Lost++
	e.Morale -= 0.5
	wr.Will[i] += 0.3
	wr.Will[1-i] -= 0.3
	w.fact(FStripped, c, e, t)
	if home {
		w.log("The horde of the %s strips %s, the home of the %s, of its ships and its people.", c.Name, w.star(t), e.Name)
	} else {
		w.log("The %s strip %s of its ships and its people. The horde grows.", c.Name, w.star(t))
	}
	if wr.Name == "" {
		wr.Name = "the war of " + w.star(t)
	}
	if e.Active() && (wr.Lost[1-i] == 1 || wr.Lost[1-i]%3 == 0) {
		w.face(e, "hold", 0)
	}
	if !e.Active() {
		w.endWar(wr, "destroyed")
	}
}

// hitFleet is a strike at a nomad fleet: broken by a third, or ended.
func (w *World) hitFleet(wr *War, c, e *Civ, t int) {
	i := wr.side(c.ID)
	for _, x := range w.fleets(e) {
		if x.Base != t {
			continue
		}
		x.Mil *= 0.7
		wr.Will[i] += 0.2
		wr.Will[1-i] -= 0.2
		if x.Mil < 1 {
			x.Over = true
			wr.Glassed[i]++
			w.log("The %s break a fleet of the %s at %s.", c.Name, e.Name, w.star(t))
		}
		if len(w.fleets(e)) == 0 {
			w.endCiv(e, Extinct, sprintf("were broken fleet by fleet by the %s", c.Name))
			w.endWar(wr, "extinction")
		}
		return
	}
}

// fleetAt is the strength of a nomad people's fleet at a star.
func (w *World) fleetAt(e *Civ, t int) float64 {
	for _, x := range w.fleets(e) {
		if x.Base == t {
			return x.Mil
		}
	}
	return 0
}

// rest is a nomad people settling for good, usually after a scar.
func (w *World) rest(c *Civ, why string) {
	hop := min(max(c.Reach, 3), 20)
	var best *Expedition
	for _, x := range w.fleets(c) {
		if x.Base >= 0 && (best == nil || x.Mil > best.Mil) {
			best = x
		}
	}
	if best == nil {
		return
	}
	t := -1
	if w.Owner[best.Base] < 0 && w.Held[best.Base] < 0 && w.canLive(c, best.Base) {
		t = best.Base
	} else {
		for _, s := range w.G.Near(best.Base, hop) {
			if w.Owner[s] < 0 && w.Held[s] < 0 && w.canLive(c, s) {
				t = s
				break
			}
		}
	}
	if t < 0 {
		return
	}
	for _, x := range w.fleets(c) {
		x.Over = true
	}
	c.Aloft, c.Rested = false, true
	c.Away = 0
	c.Systems = []int{t}
	w.Owner[t] = c.ID
	c.Home, c.HomeName = t, w.star(t)
	c.Record = append(c.Record, "came to rest")
	w.takeOver(c, t)
	w.recompute(c)
	w.fact(FRest, c, nil, t)
	if c.Has("nomadic") {
		w.log("The %s come to rest at %s, and are nomads no longer. It was %s that did it.", c.Name, c.HomeName, why)
	} else {
		w.log("The %s, refugees no longer, settle %s. It is home now.", c.Name, c.HomeName)
	}
}

// splitFleets is schism among the aloft: half the fleets go their own way.
func (w *World) splitFleets(c *Civ) {
	fl := w.fleets(c)
	if len(fl) < 2 {
		w.log("Unrest in the fleets of the %s. It passes, this time.", c.Name)
		c.Morale -= 0.5
		return
	}
	home := -1
	for _, x := range fl {
		if x.Base >= 0 && w.Owner[x.Base] < 0 {
			home = x.Base
			break
		}
	}
	if home < 0 {
		for i, x := range fl {
			if i%2 == 1 {
				x.Over = true
			}
		}
		w.log("Schism in the fleets of the %s. Half of them scatter and are not heard of again.", c.Name)
		return
	}
	sp := *c.Species
	sp.Name = names.Civ(w.R)
	sp.Traits = append([]*species.Trait(nil), c.Species.Traits...)
	sp.Add("branch")
	sp.Made = "a branch of the " + c.Name
	nc := w.spawnCiv(home, &sp, -1)
	nc.Master = -1
	w.Owner[home] = -1
	nc.Systems = nil
	nc.Aloft = true
	for i, x := range fl {
		if i%2 == 1 {
			x.Owner = nc.ID
		}
	}
	for _, k := range knownOf(c) {
		nc.Known[k] = true
	}
	w.seat(nc)
	w.recompute(nc)
	w.log("Schism in the fleets of the %s. Half of them go their own way as the %s.", c.Name, nc.Name)
}
