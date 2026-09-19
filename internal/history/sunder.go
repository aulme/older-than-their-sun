package history

import (
	"worldgen/internal/names"
	"worldgen/internal/species"
)

// Sundering: the break of an old people into more peoples, not fewer. A
// civil war deals the realm out to two or three heirs, each a new people
// of the same blood with the whole telling, a line back to the old people,
// a claim on every world of the old realm and a war with every other heir.
// A shattering, the dark age that takes the stars, makes every world its
// own people, kin to the others and at war with nobody. No heir is the old
// people: it ends here, Sundered or Shattered, its holdings handed over
// without a trace left, and everything the galaxy held against it passes
// to each heir. Kinship is descent through Line and never fades; what
// changes is whether it is acted on (kin below).

// heir raises a new people of an old one's blood at a world, dressed from
// the old people: its tree, its scars and boons, its telling, its line.
// It holds nothing until the dealing gives it something.
func (w *World) heir(old *Civ, home int, origin string) *Civ {
	nc := w.newCiv(home, old.Species, -1, names.Civ(w.R))
	nc.HomeName, nc.CradleName = w.star(home), w.star(home)
	nc.Origin = origin
	nc.Line = append(append([]int(nil), old.Line...), old.ID)
	nc.Systems, nc.Peak = nil, 0
	nc.Own = old.Own
	nc.Morality = old.Morality
	nc.KnowsCycle = old.KnowsCycle
	nc.Starfaring = old.Starfaring
	nc.DarkAges, nc.Renaissances = old.DarkAges, old.Renaissances // the institutions are the old ones', however new the name
	nc.LastDark = old.LastDark
	nc.NextDrift = old.NextDrift
	nc.Word = old.Word
	for k := range old.Known {
		nc.Known[k] = true
	}
	for _, m := range []struct{ to, from map[string]bool }{
		{nc.Scars, old.Scars}, {nc.Boons, old.Boons}, {nc.Faced, old.Faced}, {nc.Locked, old.Locked}, {nc.Lifted, old.Lifted},
	} {
		for k, v := range m.from {
			m.to[k] = v
		}
	}
	for k, v := range old.Miracles {
		nc.Miracles[k] = v
	}
	for _, m := range []struct{ to, from map[int]bool }{
		{nc.Met, old.Met}, {nc.Reached, old.Reached}, {nc.Trade, old.Trade}, {nc.Fathomed, old.Fathomed}, {nc.Watched, old.Watched}, {nc.Marked, old.Marked}, {nc.Ridden, old.Ridden},
	} {
		for k, v := range m.from {
			m.to[k] = v
		}
	}
	for k, v := range old.FathomTried {
		nc.FathomTried[k] = v
	}
	for k, v := range old.Charted {
		nc.Charted[k] = v
	}
	for k, v := range old.Intel {
		i := *v
		nc.Intel[k] = &i
	}
	for k, v := range old.Truce {
		nc.Truce[k] = v
	}
	for k, v := range old.Grudge {
		nc.Grudge[k] = v * w.Cfg.Tuning.Ossify.HeirGrudge
	}
	for k, v := range old.Fought {
		nc.Fought[k] = v
	}
	for _, m := range []struct{ to, from *map[string]bool }{{&nc.Had, &old.Had}, {&nc.Harnessed, &old.Harnessed}} {
		if *m.from != nil {
			*m.to = map[string]bool{}
			for k, v := range *m.from {
				(*m.to)[k] = v
			}
		}
	}
	w.inherit(nc, old, 0)
	return nc
}

// dealing is who gets what of the old realm: worlds and fleets to heirs
// by index.
type dealing struct {
	heirs  []*Civ
	world  map[int]int // star to heir
	fleet  map[int]int // expedition to heir
	seat   int         // the heir with the old seat
	random func() int
}

func (d *dealing) of(star int) *Civ {
	if i, ok := d.world[star]; ok {
		return d.heirs[i]
	}
	return d.heirs[d.seat]
}

// deal hands the old people's holdings to the heirs: worlds with the works,
// the guns and the guards on them, the wielded at them and the mobile
// rarities there; fleets in flight to whoever drew them, a campaign
// keeping its errand; colony ships likewise; slaves and vassals with the
// world they are ruled from; the miracles with the works or the objects
// that hold them, else to the seat. Then it ends the old people without
// a trace and passes what the galaxy held against it to each heir.
func (w *World) deal(old *Civ, d *dealing, fate Fate, cause string) {
	for _, s := range old.Systems {
		h := d.of(s)
		w.Owner[s] = h.ID
		h.Systems = append(h.Systems, s)
		if g, ok := old.Guns[s]; ok {
			if h.Guns == nil {
				h.Guns = map[int]int{}
			}
			h.Guns[s] = g
		}
		if old.GridBroken[s] {
			if h.GridBroken == nil {
				h.GridBroken = map[int]bool{}
			}
			h.GridBroken[s] = true
		}
		if r, ok := old.DockRate[s]; ok {
			if h.DockRate == nil {
				h.DockRate = map[int]float64{}
			}
			h.DockRate[s] = r
		}
	}
	for _, wk := range old.Works {
		h := d.of(wk.Star)
		h.Works = append(h.Works, wk)
		h.Structures[wk.Key]++
	}
	for _, x := range w.fleetsOf(old) {
		var h *Civ
		if i, ok := d.fleet[x.ID]; ok {
			h = d.heirs[i]
		} else if x.Base >= 0 {
			h = d.of(x.Base)
		} else {
			h = d.heirs[d.random()]
		}
		w.reown(x, h)
		if x.Back >= 0 && w.Owner[x.Back] != h.ID {
			x.Back, _ = w.nearest(h, x.Star)
		}
	}
	for _, x := range w.Expeditions {
		if !x.Over && x.Target == old.ID && x.Owner != old.ID {
			x.Target = d.of(x.Star).ID // a fleet against the old people is against whoever holds its mark
		}
	}
	for _, v := range old.Voyages {
		h := d.heirs[d.random()]
		h.Voyages = append(h.Voyages, v)
	}
	for _, id := range w.mobile {
		s := w.Sources[id]
		if s.Holder != old.ID {
			continue
		}
		switch {
		case s.Carried >= 0:
			s.Holder = w.Expeditions[s.Carried].Owner
		case s.Star >= 0:
			s.Holder = d.of(s.Star).ID
		default:
			s.Holder = d.heirs[d.seat].ID
		}
	}
	for _, l := range old.Wielded {
		h := d.of(l.Star)
		if l.Source >= 0 && w.Sources[l.Source].Holder >= 0 {
			h = w.Civs[w.Sources[l.Source].Holder]
		}
		h.Wielded = append(h.Wielded, l)
		l.Finder = h.ID
	}
	for _, key := range old.held() {
		for _, h := range d.heirs {
			if h == d.heirs[d.seat] || w.holdsObject(h, key) {
				continue
			}
			wielded := false
			for _, l := range h.Wielded {
				wielded = wielded || l.Node == key
			}
			if !wielded {
				delete(h.Known, key)
				delete(h.Miracles, key)
			}
		}
	}
	for _, o := range w.Civs {
		if o.Living() && o.Master == old.ID {
			h := d.of(o.Home)
			if _, dist := w.nearest(h, o.Home); dist > 0 {
				for _, x := range d.heirs {
					if _, dx := w.nearest(x, o.Home); len(x.Systems) > 0 && dx < dist {
						h, dist = x, dx
					}
				}
			}
			o.Master = h.ID
			h.Ruled++
		}
	}
	for _, h := range d.heirs {
		h.Peak = len(h.Systems)
		if h.Aloft {
			h.Peak = len(w.fleets(h))
		}
	}
	w.sunder(old, d.heirs, fate, cause)
}

// sunder ends the old people without a trace: its holdings are already
// dealt, its wars end (the heirs' are already opened or not, by the
// caller), its pacts dissolve, and what every other people held against
// it passes to each heir: grudges at half, truces whole, intelligence,
// watchfulness, the memory of it as a monster, and the trade it had.
func (w *World) sunder(old *Civ, heirs []*Civ, fate Fate, cause string) {
	t := &w.Cfg.Tuning.Ossify
	for _, o := range w.Civs {
		if o == old || !o.Living() {
			continue
		}
		for _, h := range heirs {
			if g, ok := o.Grudge[old.ID]; ok {
				o.Grudge[h.ID] = g * t.HeldGrudge
			}
			if y, ok := o.Truce[old.ID]; ok {
				o.Truce[h.ID] = y
			}
			if i := o.Intel[old.ID]; i != nil {
				c := *i
				o.Intel[h.ID] = &c
			}
			for _, m := range []map[int]bool{o.Watched, o.Met, o.Reached, o.Fathomed, o.Trade, o.Dependent, o.Suspect, o.Closed, o.Barred, o.Ridden, o.monsters} {
				if m != nil && m[old.ID] {
					m[h.ID] = true
				}
			}
			if y, ok := o.FathomTried[old.ID]; ok {
				o.FathomTried[h.ID] = y
			}
			if n, ok := o.Fought[old.ID]; ok {
				o.Fought[h.ID] = n
			}
		}
		for _, m := range []map[int]bool{o.Trade, o.Dependent, o.Watched} {
			delete(m, old.ID)
		}
	}
	for _, wr := range w.Wars {
		if !wr.Over && (wr.Sides[0] == old.ID || wr.Sides[1] == old.ID) {
			wr.Over, wr.Ended, wr.Result = true, w.Now, cause
			e := w.Civs[wr.Sides[1-wr.side(old.ID)]]
			delete(e.Wars, old.ID)
		}
	}
	for _, pid := range old.Pacts {
		p := w.Pacts[pid]
		if p.Over {
			continue
		}
		p.Members = remove(p.Members, old.ID)
		if len(p.Members) < 2 {
			p.Over, p.Ended = true, w.Now
		}
	}
	old.Systems, old.Works, old.Wielded, old.Voyages, old.Guns, old.Muster = nil, nil, nil, nil, nil, nil
	old.Wars, old.Trade = map[int]bool{}, map[int]bool{}
	old.Stage, old.Fate, old.Cause, old.Ended, old.Fell = Dead, fate, cause, w.Now, w.Now
	old.FellDependent = len(old.Dependent) > 0
	old.Into = "the " + heirs[0].Name
	for _, h := range heirs[1:] {
		old.Into += ", the " + h.Name
	}
}

// inheritWar carries an open war of the old people on against an heir:
// the same cause, the old side's will, a fresh count.
func (w *World) inheritWar(wr *War, old, h *Civ) {
	e := w.Civs[wr.Sides[1-wr.side(old.ID)]]
	if !e.Active() {
		return
	}
	i := wr.side(old.ID)
	nw := &War{ID: len(w.Wars), Sides: [2]int{h.ID, e.ID}, Began: wr.Began, Cause: wr.Cause, Nth: 1, Pact: -1, Principal: -1, Hire: -1, Contested: map[int]int{}, Called: map[int]bool{},
		Slights: map[int]float64{}, Sent: map[int]float64{}, Slighted: map[int]float64{}, SlightTold: map[int]bool{}}
	nw.Will = [2]float64{wr.Will[i], wr.Will[1-i]}
	w.Wars = append(w.Wars, nw)
	h.Wars[e.ID], e.Wars[h.ID] = true, true
	h.Fought[e.ID], e.Fought[h.ID] = 1, e.Fought[h.ID]+1
	h.Tally.Fought++
	e.Tally.Fought++
}

// civilWar is the realm breaking into two or three heirs at war with each
// other over it. It returns false where there is nothing to break: one
// world, or a people with no factions.
func (w *World) civilWar(c *Civ) bool {
	t := &w.Cfg.Tuning.Ossify
	if !c.Active() {
		return false
	}
	if !c.Species.Profile().Can(species.CivilWars) {
		w.log("The %s cannot split; a hive has no factions. The pressure goes elsewhere.", c.Name)
		c.Morale -= 1
		return false
	}
	var parts []int
	if c.Aloft {
		for _, x := range w.fleets(c) {
			parts = append(parts, x.ID)
		}
	} else {
		parts = append(parts, c.Systems...)
	}
	if len(parts) < 2 {
		if c.Aloft {
			w.log("Unrest in the fleets of the %s. It passes, this time.", c.Name)
		} else {
			w.log("Unrest among the %s on %s. It passes, this time.", c.Name, c.HomeName)
		}
		c.Morale -= 0.5
		return false
	}
	n := 2
	if len(parts) >= 6 && w.R.Float64() < 0.5 {
		n = 3
	}
	d := &dealing{world: map[int]int{}, fleet: map[int]int{}, random: func() int { return w.R.IntN(n) }}
	order := w.R.Perm(len(parts))
	first := make([]int, n)
	for k, idx := range order {
		i := d.random()
		if k < n {
			i = k
			first[i] = parts[idx]
		}
		if c.Aloft {
			d.fleet[parts[idx]] = i
		} else {
			d.world[parts[idx]] = i
		}
	}
	old := append([]int(nil), c.Systems...)
	oldWars := []*War{}
	for _, wr := range w.Wars {
		if !wr.Over && (wr.Sides[0] == c.ID || wr.Sides[1] == c.ID) {
			oldWars = append(oldWars, wr)
		}
	}
	for i := range n {
		home := c.Home
		if c.Aloft {
			x := w.Expeditions[first[i]]
			if x.Base >= 0 {
				home = x.Base
			} else {
				home = x.Star
			}
			d.fleet[first[i]] = i
		} else if d.world[c.Home] != i {
			home = first[i]
		} else {
			d.seat = i
		}
		h := w.heir(c, home, "heirs of the "+c.Name)
		h.Aloft = c.Aloft
		d.heirs = append(d.heirs, h)
	}
	if c.Aloft {
		d.seat = d.fleet[w.greatestFleet(c).ID]
	}
	w.deal(c, d, Sundered, "tore themselves apart")
	for _, h := range d.heirs {
		w.forget(h, 0.1) // the arsenals were in the other province
		for _, s := range old {
			if h.Claim == nil {
				h.Claim = map[int]bool{}
			}
			h.Claim[s] = true
		}
		w.branchMorality(h, c)
		for _, wr := range oldWars {
			w.inheritWar(wr, c, h)
		}
		if h.Aloft {
			w.seat(h)
		}
		w.recompute(h)
	}
	w.tearApart(c, d.heirs, d.heirs[d.seat])
	for i, a := range d.heirs {
		for _, b := range d.heirs[i+1:] {
			a.Met[b.ID], b.Met[a.ID], a.Reached[b.ID], b.Reached[a.ID] = true, true, true, true
			a.Fathomed[b.ID], b.Fathomed[a.ID] = true, true
			w.declare(a, b, "the sundering")
			a.Grudge[b.ID], b.Grudge[a.ID] = t.SunderGrudge, t.SunderGrudge
		}
	}
	return true
}

// tearApart is the record of a civil war: the facts and the line.
func (w *World) tearApart(old *Civ, heirs []*Civ, seat *Civ) {
	for _, h := range heirs {
		f := w.factN(FSundered, old, h, h.Home, len(heirs))
		f.What = "the true " + old.Name
	}
	var ns []string
	for _, h := range heirs {
		ns = append(ns, "the "+h.Name)
	}
	seatLine := ""
	if !old.Aloft {
		seatLine = sprintf(" The %s hold the old seat.", seat.Name)
	}
	w.log("The %s tear themselves in %s: %s, each the true %s by its own telling, each holding the others traitors.%s", old.Name, numberWord(len(heirs)), listOf(ns), old.Name, seatLine)
}

// shatter is the dark age that took the stars: every world its own people,
// the seat included, with the reduced tree, the telling a step more worn,
// the works and the garrison there, and kin to the rest. Past the cap the
// other worlds are abandoned.
func (w *World) shatter(c *Civ, why string) {
	t := &w.Cfg.Tuning.Ossify
	worlds := []int{c.Home}
	for _, s := range w.R.Perm(len(c.Systems)) {
		if c.Systems[s] != c.Home {
			worlds = append(worlds, c.Systems[s])
		}
	}
	for _, s := range worlds[min(len(worlds), t.Shards):] {
		w.loseSystem(c, s, "abandoned "+c.Species.Flavour().Colony, "")
	}
	worlds = worlds[:min(len(worlds), t.Shards)]
	d := &dealing{world: map[int]int{}, fleet: map[int]int{}, random: func() int { return 0 }}
	for i, s := range worlds {
		d.world[s] = i
		h := w.heir(c, s, "heirs of the "+c.Name)
		d.heirs = append(d.heirs, h)
	}
	w.deal(c, d, Shattered, "forgot how to reach the stars")
	for _, h := range d.heirs {
		for _, tl := range h.Lore {
			tl.Wear = min(2, tl.Wear+1) // a step more worn than the old people held it
		}
		h.Stage = Emergent
		w.recompute(h)
		f := w.factN(FShattered, c, h, h.Home, len(worlds))
		f.What = why
	}
	w.log("The %s forget how to reach the stars. On %s worlds %s peoples wake up alone: %s.", c.Name, numberWord(len(worlds)), numberWord(len(worlds)), w.shardList(d.heirs))
}

func (w *World) shardList(hs []*Civ) string {
	var ns []string
	for _, h := range hs {
		ns = append(ns, "the "+h.Name+" on "+h.HomeName)
	}
	return listOf(ns)
}

// kin says whether two peoples share a line: one in the other's line, or
// an ancestor in both. It is a fact of descent and never fades.
func (w *World) kin(a, b *Civ) bool {
	if a == b {
		return false
	}
	if contains(a.Line, b.ID) || contains(b.Line, a.ID) {
		return true
	}
	for _, x := range a.Line {
		if contains(b.Line, x) {
			return true
		}
	}
	return false
}

// ofLine says whether a people is this one or one it came out of: its
// deeds are ours, its crimes excused as ours are.
func (c *Civ) ofLine(id int) bool { return id == c.ID || contains(c.Line, id) }

// feud says whether a grudge above the enemy line stands between two
// peoples, which suspends kinship's warmth until it decays.
func (w *World) feud(a, b *Civ) bool {
	line := w.Cfg.Tuning.Ossify.KinLine
	return a.Grudge[b.ID] > line || b.Grudge[a.ID] > line
}

// warm says whether kinship is acted on: kin, with no feud and no monster
// reckoning either way.
func (w *World) warm(a, b *Civ) bool {
	return w.kin(a, b) && !w.feud(a, b) && !w.monster(a, b) && !w.monster(b, a)
}

// claims says whether c holds a claim on any world e holds.
func (w *World) claims(c, e *Civ) bool {
	for _, s := range e.Systems {
		if c.Claim[s] {
			return true
		}
	}
	return false
}

// reclaimed is a claimed world taken: restored to the realm in the taker's
// own telling, and the claims gone when the whole old realm is held.
func (w *World) reclaimed(c, e *Civ, t int) {
	w.fact(FReclaimed, c, e, t)
	w.log("The %s call %s restored to the realm.", c.Name, w.star(t))
	for s := range c.Claim {
		if w.Owner[s] != c.ID {
			return
		}
	}
	c.Claim = nil
	w.log("The %s hold every world the %s held. There is nothing left to claim, and they are the %s that hold what the %s held.", c.Name, w.Civs[c.Line[len(c.Line)-1]].Name, c.Name, w.Civs[c.Line[len(c.Line)-1]].Name)
}

// kinMeet is kin finding each other again with no feud between them: the
// tales and the trade at once, no council, and a pact of defence offered
// with the loyalty dial's odds doubled.
func (w *World) kinMeet(a, b *Civ) {
	w.log("The %s and the %s, both of the line of the %s, find each other again.", a.Name, b.Name, w.Civs[w.commonLine(a, b)].Name)
	w.openPair(a, b)
	if a.Free() && b.Free() && !w.allied(a, b) {
		w.send(a, b, &Message{Kind: MsgPact, PactKind: Defensive, Target: -1, Pact: -1})
	}
}

// commonLine is the nearest people in both lines, or the one in the
// other's line.
func (w *World) commonLine(a, b *Civ) int {
	if contains(a.Line, b.ID) {
		return b.ID
	}
	if contains(b.Line, a.ID) {
		return a.ID
	}
	for i := len(a.Line) - 1; i >= 0; i-- {
		if contains(b.Line, a.Line[i]) {
			return a.Line[i]
		}
	}
	return a.Line[len(a.Line)-1]
}

// numberWord is a small count in words.
func numberWord(n int) string {
	if n >= 0 && n < len(numberWords) {
		return numberWords[n]
	}
	return sprintf("%d", n)
}

var numberWords = []string{"no", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"}

func listOf(ns []string) string {
	switch len(ns) {
	case 0:
		return ""
	case 1:
		return ns[0]
	}
	out := ""
	for i, n := range ns {
		switch {
		case i == 0:
			out = n
		case i == len(ns)-1:
			out += " and " + n
		default:
			out += ", " + n
		}
	}
	return out
}
