package history

import (
	"worldgen/internal/flow"
	"worldgen/internal/mind"
	"worldgen/internal/plague"
	"worldgen/internal/species"
)

// Parasites: a people that is a plague. A rider of the flesh is a
// biological one and a rider of the mind a memetic one; each parasite
// people is one Plague, and obeys every rule of plague.go but two. Its
// plague does not spread on its own: on every channel event where it could
// carry, the parasite's council decides whether to try, and the attempt is
// an engineered one (weapon.go), with the same blame. And its lethality
// does not empty worlds, it takes them: a world that fails the toll goes
// over to the parasite, and the home going over is the people ridden. A
// parasite is born when a plague that thinks sees its first host's home go
// over, or, one cradle in a hundred, with the people it rides. It holds no
// world of its own beyond what went over, lives on its hosts' income, and
// dies the tick its last host is dead or free.

// rides says whether a people is a parasite with a host to be in, which
// is what stands in for a world of its own.
func (w *World) rides(c *Civ) bool {
	return c.Own >= 0 && len(w.hostsOf(c)) > 0
}

// riderOf is the living parasite people a plague is, or nil.
func (w *World) riderOf(p *Plague) *Civ {
	if p.Rider < 0 || !w.Civs[p.Rider].Living() {
		return nil
	}
	return w.Civs[p.Rider]
}

// ridden says whether a people is a host worn by a parasite.
func (w *World) ridden(c *Civ) bool {
	return c.Master >= 0 && !c.Vassal && w.Civs[c.Master].Own >= 0
}

// hostsOf lists the living peoples a parasite is in: ridden, or fighting
// its plague.
func (w *World) hostsOf(p *Civ) []*Civ {
	if p.Own < 0 {
		return nil
	}
	var out []*Civ
	for _, c := range w.Civs {
		if c != p && c.Living() && ((c.Master == p.ID && !c.Vassal) || c.Infections[p.Own] != nil) {
			out = append(out, c)
		}
	}
	return out
}

// wake is a plague that thinks becoming a people: a parasite of the
// plague's kind, spawned at the host's home with no world of its own,
// riding the host. A plague made to think on purpose is a made people
// with its maker as master, if the maker is there to hold it.
func (w *World) wake(p *Plague, host *Civ) *Civ {
	sp := species.GenerateWith(w.R, w.G.Stars[host.Home].Mult, host.Species.World.Key, species.Parasite, 0)
	if p.Kind == plague.Memetic {
		sp.Replace("bodyrider", "mindrider")
	} else {
		sp.Replace("mindrider", "bodyrider")
	}
	sp.Made = "woke in " + p.Name
	nc := w.spawn(host.Home, sp, -1, "", host)
	nc.Own, p.Rider, p.Conscious = p.ID, nc.ID, true
	nc.Origin = "something that woke in " + p.Name
	if p.Maker >= 0 && p.Made && w.Civs[p.Maker].Active() {
		m := w.Civs[p.Maker]
		nc.Master, nc.Vassal, nc.Seen = m.ID, true, m.Declines
		m.Ruled++
		w.log("The %s made %s to think, and it does. It answers to them, for now.", m.Name, p.Name)
	}
	w.factOf(FWoke, nc, host, host.Home, p.Name)
	w.log("Something in %s has begun to think. At %s it takes the %s for its own, and calls itself the %s.", p.Name, host.HomeName, host.Name, nc.Name)
	w.ride(nc, host)
	if wr := w.warBetween(nc.ID, host.ID); wr != nil {
		w.endWar(wr, "enslaved")
	}
	return nc
}

// bornRider is the cradle roll: one people in a hundred is born with a
// rider already in it, its home already over.
func (w *World) bornRider(c *Civ) {
	t := &w.Cfg.Tuning.Plague
	if !w.chance(t.BornRider) || !w.bears(c, plague.Biological) {
		return
	}
	p := w.newPlague(plague.Biological, c, "born rider")
	p.Conscious = true
	w.infect(c, p, nil, "born")
	nc := w.wake(p, c)
	nc.Origin = "a rider born with the " + c.Name
	w.log("The %s have never known a time before %s. They grew up ridden.", c.Name, p.Name)
}

// freed is a ridden people winning the contest: free, immune, and never
// again quite the same.
func (w *World) freed(h, rider *Civ) {
	h.Master, h.Vassal = -1, false
	h.Scars[ScarChains] = true
	rider.Tally.Risen++
	w.fact(FFreed, h, rider, h.Home)
	if rider.Has("mindrider") {
		w.log("The %s learn to unthink the %s. What was in their heads is gone, and they are free, and never again quite trust a new idea.", h.Name, rider.Name)
	} else {
		w.log("The %s find a drug that kills what rides them. They are free, and careful about their blood ever after.", h.Name)
	}
}

// tryRide is a parasite on a channel event: its council on the people at
// the other end, and the attempt if it says so.
func (w *World) tryRide(c, e *Civ, roadKey string) {
	if c.Own < 0 || !c.Active() || !e.Active() || e.Own >= 0 || e.Master == c.ID {
		return
	}
	p := w.Plagues[c.Own]
	if p.Kind != roads[roadKey].Kind || !w.catchable(e, p) {
		return
	}
	odds := false
	if c.Met[e.ID] && e.Free() {
		bar, wants, far := w.bar(c, e)
		if wants {
			ap := w.appraise(c, e, -1)
			odds = ap.Acted >= bar && (len(ap.Front) > 0 || far)
		}
	}
	r := mind.TryRide(mind.RideInput{
		Posture: c.posture(), Hates: c.hates(e), AtWar: c.Wars[e.ID],
		Prey:  w.plagued(e) || (e.DarkAges > 0 && float64(w.Now-e.LastDark) <= w.Cfg.Tuning.Plague.DarkYears),
		Odds:  odds,
		Sworn: w.allied(c, e) && !w.betrayed(c, e),
		Hurry: len(w.hostsOf(c)) <= 1 && len(c.Systems) == 0,
	})
	w.explain(c, "riding the "+e.Name, r)
	if r.Try {
		w.attempt(c, e, p, roadKey)
	}
}

// rideAll is the parasite's every-tick channels, its own and its hosts',
// since the rider is the mind: goods moving on a trade link for a rider of
// the flesh, the signals of everyone in earshot for a rider of the mind.
func (w *World) rideAll(c *Civ) {
	if c.Own < 0 || !c.Active() {
		return
	}
	eyes := []*Civ{c}
	for _, h := range w.hostsOf(c) {
		if h.Master == c.ID && !h.Vassal && h.Active() {
			eyes = append(eyes, h)
		}
	}
	tried := map[int]bool{}
	if w.Plagues[c.Own].Kind == plague.Biological {
		for _, a := range eyes {
			for _, eid := range sortedInts(a.Trade) {
				e := w.Civs[eid]
				if e.Active() && !tried[eid] && a.From[eid].Total()+e.From[a.ID].Total() > 0 {
					tried[eid] = true
					w.tryRide(c, e, "goods")
				}
			}
		}
		return
	}
	for _, a := range eyes {
		for _, eid := range sortedInts(a.Met) {
			if e := w.Civs[eid]; e.Active() && !tried[eid] && w.hear(a, e) {
				tried[eid] = true
				w.tryRide(c, e, "signal")
			}
		}
	}
}

// converted is a world of a host that failed the toll going over to the
// parasite: the home ridden, a colony a host-world.
func (w *World) converted(rider, c *Civ, p *Plague, s int) {
	if s == c.Home {
		w.log("%s takes %s. World by world, the %s were ridden, and now they are.", upper(p.Name), c.HomeName, c.Name)
		w.ride(rider, c)
		if wr := w.warBetween(rider.ID, c.ID); wr != nil {
			w.endWar(wr, "enslaved")
		}
		return
	}
	w.loseSystem(c, s, "host-world", "")
	w.Owner[s] = rider.ID
	rider.Systems = append(rider.Systems, s)
	rider.Peak = max(rider.Peak, len(rider.Systems))
	w.fact(FTaken, rider, c, s)
	w.log("%s takes %s, a %s of the %s. It is a host-world of the %s now.", upper(p.Name), w.star(s), c.Species.Flavour().Colony, c.Name, rider.Name)
}

// burn is a contained host at war with its rider burning one host-world
// the rider holds within its reach each tick: a peace earned rather than
// rolled.
func (w *World) burn(c, rider *Civ) {
	if !c.Wars[rider.ID] || !c.Active() {
		return
	}
	for _, s := range append([]int(nil), rider.Systems...) {
		if !w.inReach(c, s) {
			continue
		}
		w.Bio[s] = BioSimple
		w.loseSystem(rider, s, "burned host-world", "")
		w.log("The %s burn %s to be rid of what the %s put there.", c.Name, w.star(s), rider.Name)
		w.fact(FBurned, c, rider, s)
		if wr := w.warBetween(c.ID, rider.ID); wr != nil {
			wr.Glassed[wr.side(c.ID)]++
			wr.Will[1-wr.side(c.ID)] -= 0.4
		}
		return
	}
}

// starve ends the parasites with nothing left to wear, and counts the
// rest's hosts.
func (w *World) starve() {
	for _, c := range w.Civs {
		if c.Own >= 0 && c.Living() {
			w.starveOne(c)
		}
	}
}

// starveOne is one parasite's count of hosts, and its end if there are
// none. A rider's worlds keep its plague in their dead cities.
func (w *World) starveOne(c *Civ) {
	t := &w.Cfg.Tuning.Plague
	hosts := w.hostsOf(c)
	c.Hosts = len(hosts)
	if len(hosts) > 0 || !c.Active() {
		return
	}
	p := w.Plagues[c.Own]
	if p.Kind == plague.Biological {
		for _, s := range c.Systems {
			w.Reservoir[s] = &Reservoir{Plague: p.ID, Until: w.Now + Year(p.Contagion*t.ReservoirMyr*1e6)}
		}
	}
	w.endCiv(c, Extinct, "had nothing left to wear")
}

// riderTithe is what a ridden people's income loses to its rider before
// it feeds anything: the rider's upkeep, shared among its hosts.
func (w *World) riderTithe(c *Civ, in flow.Income) flow.Income {
	if !w.ridden(c) {
		return in
	}
	m := w.Civs[c.Master]
	n := 0
	for _, h := range w.hostsOf(m) {
		if h.Master == m.ID && !h.Vassal {
			n++
		}
	}
	if n == 0 {
		return in
	}
	share := m.Upkeep.Scale(1 / float64(n))
	for k := range in {
		in[k] = max(0, in[k]-share[k])
	}
	return in
}
