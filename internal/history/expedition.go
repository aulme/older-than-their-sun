package history

import (
	"worldgen/internal/mind"
	"worldgen/internal/species"
)

// Every ship is in a fleet, and this is the fleet: one object for the
// guard at a star, the campaign sent against an enemy, the relief sent to
// stand with a host, the scout that only looks, the surveyors, and a
// horde's fleet. A fleet has a count of ships, a base or a line in flight,
// an owner and a kind; it changes kind by order, splits when ships are
// taken out of it and merges with another of its owner and kind at a
// base. Its strength is its ships at its owner's quality; the arithmetic
// of a fight is internal/battle.

// ExpKind is what a fleet is for.
type ExpKind uint8

const (
	Campaign ExpKind = iota
	Relief
	Scout
	Roam      // a nomad people's fleet; see nomad.go
	Survey    // surveyors reading the stars; see explore.go
	Guard     // the ships at one of the people's stars: its garrison and its reserve; see ships.go
	Intercept // sent to meet a fleet in the dark; see intercept.go
)

func (k ExpKind) String() string {
	return [...]string{"campaign", "relief", "scout", "roam", "survey", "guard", "intercept"}[k]
}

// Expedition is one fleet.
type Expedition struct {
	ID        int
	Owner     int
	Target    int // the people fought, or stood with, or looked at; -1 if none
	Kind      ExpKind
	Star      int // destination
	From      int // where it set out from
	Ships     int
	Launched  Year
	Arrive    Year
	Base      int // where it operates from; -1 in flight
	Returning bool
	Over      bool
	LaidUp    bool  // the flow does not keep it: it neither fights nor moves, and it rots
	Laid      Year  // when it was last laid up
	Manned    Year  // when it was last manned again
	Held      []int // worlds it took and its people still hold
	Seen      map[int]bool
	Report    *Intel // a scout's report, carried home without the Voice
	Battles   int
	Wins      int
	Turned    bool
	Fed       Year    // for a roaming fleet: when it reached its base
	Tour      int     // for surveyors: stars read this trip
	Recalled  bool    // for surveyors: called home by a war; they finish the leg and turn
	Out       Year    // when the fleet first set out; Launched is the start of its current leg
	Back      int     // for a campaign fleet that fell back: the world it returns to, or -1
	Sieges    int     // times it fell back and came again
	Withdrawn bool    // for a relief fleet: moving between its host's worlds, not arriving anew
	Drive     float64 // years per light year on this leg
	Path      *[2]vec // the leg's ends when either is a point in the dark, not a star; see geom.go
	Warning   Year    // the most warning any people had of it: its arrival less the sighting
	// interception: see intercept.go
	Quarry int  // the fleet an interceptor was sent to meet
	Meet   Year // when
	Leg    Year // the quarry's Launched when it was seen: the leg the meeting was computed on
	Picket bool // for a scout: it stays at its star and watches
	// contracts: see contract.go
	Contract int // the contract this fleet does a guard or a strike under, or -1
	SoldBy   int // the people that last sold a sighting of it, or -1
}

// launch sends a fleet of n ships toward a star: the ships are taken
// from the people's guard nearest the star that has them, and the fleet
// sets out from there. With no guard that has n ships manned it does not
// go, and nil is returned; a campaign that cannot go is mustered by the
// council instead. The ships were kept already; a fleet in flight keeps
// them still.
func (w *World) launch(c *Civ, kind ExpKind, target *Civ, star int, n int) *Expedition {
	if !c.Species.Profile().Can(species.Launches) {
		return nil // a world does not sail
	}
	total := w.ships(c)
	g := w.guardWith(c, star, n)
	if g == nil || n <= 0 {
		if w.Cfg.TraceAI {
			w.log("[the %s cannot man a %s of %s for %s]", c.Tok(), kind, shipsWord(n), w.star(star))
		}
		return nil
	}
	from := g.Base
	g.Ships -= n
	if g.Ships == 0 && g.Kind == Guard {
		g.Over = true // an empty guard is nothing; a horde's fleet is its base and stays
	}
	d := w.G.Dist(from, star)
	x := &Expedition{ID: len(w.Expeditions), Owner: c.ID, Target: -1, Kind: kind, Star: star, From: from, Ships: n, Back: -1, Contract: -1, SoldBy: -1,
		Launched: w.Now, Out: w.Now, Arrive: w.Now + Year(d*c.Speed), Base: -1, Manned: w.Now, Seen: map[int]bool{}, Drive: c.Speed}
	if target != nil {
		x.Target = target.ID
	}
	w.addExpedition(x)
	w.timetable(x)
	if kind != Scout && kind != Survey {
		w.newEye(c, eye{star: -1, r: max(fleetEye, c.watchRange()/2), fleet: x, kind: eyeFleet})
	}
	switch kind {
	case Campaign:
		c.Tally.Fleets++
		c.WantShips = 0
		if wr := w.warBetween(c.ID, target.ID); wr != nil && wr.Gap != nil && wr.Sides[0] == c.ID {
			w.log("The %s send %s of their ships into the hole around %s, where the ledger says something is: a fleet of %s bound for %s, %s away.", c.Tok(), shareWord(n, total), w.star(wr.Gap.Star), shipsWord(n), w.star(star), span(x.Arrive-w.Now))
		} else {
			w.log("The %s send %s of their ships against the %s: a fleet of %s bound for %s, %s away.", c.Tok(), shareWord(n, total), target.Tok(), shipsWord(n), w.star(star), span(x.Arrive-w.Now))
		}
	case Relief:
		c.Tally.Relief++
		w.log("The %s send %s to stand with the %s at %s, %s away.", c.Tok(), shipsWord(n), target.Tok(), w.star(star), span(x.Arrive-w.Now))
	case Scout:
		c.Tally.Scouts++
		if w.Cfg.TraceAI {
			w.log("[the %s send a scout to %s, %s away]", c.Tok(), w.star(star), span(x.Arrive-w.Now))
		}
	case Survey:
		// logged by survey, which knows whether it is the first
	}
	return x
}

func shareWord(n, total int) string {
	switch f := float64(n) / float64(max(total, 1)); {
	case f >= 0.6:
		return "most"
	case f >= 0.4:
		return "half"
	case f >= 0.25:
		return "a third"
	}
	return "a part"
}

// tickExpeditions moves every fleet a tick: first the sightings and the
// meetings due before the next tick, in year order, so a crossing shorter
// than a tick is seen and met before it arrives; then every fleet. A
// guard at its star has nothing to do here, its keep and its rot being in
// the civ tick; a guard in flight is on its way to another guard.
func (w *World) tickExpeditions() {
	w.sightingsDue()
	w.meetingsDue()
	for _, x := range w.Expeditions {
		if x.Over {
			continue
		}
		c := w.Civs[x.Owner]
		if x.Kind == Guard {
			if !c.Active() {
				x.Over = true
				continue
			}
			if x.Base >= 0 {
				continue
			}
		}
		if x.Kind == Roam {
			if !c.Active() {
				x.Over = true
			} else if x.Base < 0 && w.Now >= x.Arrive {
				x.Base, x.Fed = x.Star, w.Now
				w.chart(c, x.Star, "fleet")
			}
			continue
		}
		if !c.Active() {
			if len(x.Held) > 0 && x.Base >= 0 {
				w.goNative(x)
			} else {
				x.Over = true
			}
			continue
		}
		if x.Base >= 0 && w.rarityAt(x.Base, "horizon") != nil {
			x.Ships -= w.count(horizonLoss * float64(x.Ships)) // a fleet based at a black hole is lost a little at a time
		}
		if x.Returning {
			if w.Now >= x.Arrive {
				if o := w.Owner[x.Star]; (x.Kind == Campaign || x.Kind == Relief || x.Kind == Intercept) && o >= 0 && o != c.ID {
					w.goHome(x) // turned back to a star somebody else holds: on from there
					continue
				}
				if x.Report != nil && x.Target >= 0 {
					if c.receive(x.Target, x.Report) {
						c.Scouted[x.Target] = w.Now
						delete(c.Watched, x.Target)
						w.forward(c, w.Civs[x.Target], x.Report)
					}
					c.Summoned = true
				}
				w.mergeInto(x, x.Star)
			}
			continue
		}
		if x.Base < 0 {
			if w.Now < x.Arrive {
				continue
			}
			w.arrive(x)
			continue
		}
		if x.LaidUp {
			continue // inoperable until the flow comes back
		}
		switch x.Kind {
		case Campaign:
			w.campaign(x)
		case Relief:
			w.station(x)
		case Scout:
			w.picketStep(x)
		}
	}
}

// arrive is a fleet reaching its destination.
func (w *World) arrive(x *Expedition) {
	c := w.Civs[x.Owner]
	switch x.Kind {
	case Survey:
		w.surveyArrive(x)
	case Intercept:
		w.meetInDark(x)
	case Scout:
		w.scoutSeen(x)
		if x.Picket {
			w.picketPost(x)
			return
		}
		if x.Target >= 0 {
			if e := w.Civs[x.Target]; e.Living() {
				rep := w.look(c, e, x.Star, 0.2)
				if c.miracle("ansible") {
					if c.receive(e.ID, rep) {
						c.Scouted[e.ID] = w.Now
						delete(c.Watched, e.ID)
						w.forward(c, e, rep)
					}
					c.Summoned = true
				} else {
					x.Report = rep
				}
			}
		}
		w.goHome(x)
	case Campaign:
		e := w.Civs[x.Target]
		if !e.Active() || w.warBetween(c.ID, e.ID) == nil {
			w.log("The fleet of the %s arrives at %s to find the war over.", c.Tok(), w.star(x.Star))
			w.resolve(x)
			return
		}
		x.Base = x.Star
		if back := x.Back; back >= 0 {
			// fallen back a hop: it comes again while the enemy holds the world
			x.Back = -1
			if w.holds(e, back) {
				w.sail(x, back)
			}
			return
		}
		if w.holds(e, x.Base) {
			w.log("The fleet of the %s arrives at %s, %s after it set out.", c.Tok(), w.star(x.Base), span(w.Now-x.Out))
		} else if wr := w.warBetween(c.ID, e.ID); wr.Gap != nil && w.R.Float64() < 0.3 {
			w.log("The hunting fleet of the %s arrives at %s and finds nothing there, which is what it was told it would find.", c.Tok(), w.star(x.Base))
		}
	case Relief:
		h := w.Civs[x.Target]
		if k := w.contractOf(x); k != nil {
			// a hired guard: it stands whether or not its host is at war, and its clock starts now
			if k.State != Running || w.Owner[x.Star] != h.ID {
				w.resolve(x)
				return
			}
			x.Base = x.Star
			k.Until = w.Now + Year(k.Length*1000)
			w.log("A fleet of the %s arrives at %s to hold it for the %s, as paid.", c.Tok(), w.star(x.Base), h.Tok())
			w.landed(c, h)
			return
		}
		if !h.Active() || len(h.Wars) == 0 {
			w.log("The fleet of the %s arrives at %s to find the war over.", c.Tok(), w.star(x.Star))
			w.resolve(x)
			return
		}
		x.Base = x.Star
		w.landed(c, h)
		if x.Withdrawn {
			x.Withdrawn = false
			return
		}
		if w.Owner[x.Base] != h.ID {
			x.Base = w.nearestOf(h, x.Star, fleetHop, -1)
			if x.Base < 0 {
				w.resolve(x)
				return
			}
		}
		w.faith(c, h, 0.5)
		w.log("A fleet of the %s arrives at %s to stand with the %s.", c.Tok(), w.star(x.Base), h.Tok())
		w.fact(FRelief, c, h, x.Base)
	}
}

// campaign is a fleet's tick in enemy territory: attrition, then the
// battle at the world it is at, or the move to the enemy's nearest world
// within a hop. No battle happens without a fleet at the world.
func (w *World) campaign(x *Expedition) {
	c, e := w.Civs[x.Owner], w.Civs[x.Target]
	if !e.Active() || w.warBetween(c.ID, e.ID) == nil {
		w.resolve(x)
		return
	}
	if !w.canLive(c, x.Base) {
		x.Ships -= w.count(0.03 * float64(x.Ships))
	}
	if x.Ships <= 0 {
		w.log("The fleet of the %s wastes away at %s, far from anything it could live on.", c.Tok(), w.star(x.Base))
		w.resolve(x)
		return
	}
	if w.holds(e, x.Base) {
		w.fight(x, x.Base)
		return
	}
	t := -1
	if wr := w.warBetween(c.ID, e.ID); wr.Gap != nil && wr.Sides[0] == c.ID {
		t = w.huntNext(c, wr, x.Base) // a hunt goes where the hole is, not where the enemy is
	} else {
		t = w.nearestOf(e, x.Base, fleetHop, x.Base)
	}
	if t < 0 {
		w.resolve(x)
		return
	}
	w.sail(x, t)
}

// resolve is a fleet's campaign ending: it goes native, is lost, or turns for home.
func (w *World) resolve(x *Expedition) {
	if x.Over {
		return
	}
	c := w.Civs[x.Owner]
	var held []int
	for _, s := range x.Held {
		if w.Owner[s] == c.ID && s != c.Home {
			held = append(held, s)
		}
	}
	x.Held = held
	if len(held) > 0 && x.Base >= 0 {
		away := float64(w.Now - x.Out)
		p := clamp(0.15+away/60_000+w.G.Dist(x.Base, c.Home)/80, 0, 0.9)
		if !c.Active() || w.R.Float64() < p {
			w.goNative(x)
			return
		}
	}
	if x.Ships <= 0 && len(held) == 0 {
		c.Morale -= 0.5
		x.Over = true
		w.fleetLost(x, max(x.Base, x.Star))
		return
	}
	w.goHome(x)
}

// horizonLoss is the share of a fleet lost per tick based at a black hole.
const horizonLoss = 0.05

// goHome turns a fleet for the nearest holding, where it joins the guard.
func (w *World) goHome(x *Expedition) {
	c := w.Civs[x.Owner]
	from := x.Base
	if from < 0 {
		from = x.Star
	}
	to, d := w.nearest(c, from)
	x.Returning = true
	x.Base = -1
	x.Path = nil
	x.From, x.Star = from, to
	x.Launched = w.Now
	x.Arrive = w.Now + Year(d*c.Speed)
	x.Drive = c.Speed
}

// goNative is a fleet that never comes home: its captains keep what they
// hold as a people of their own, a successor state with the parent's tree,
// and the fleet is its guard.
func (w *World) goNative(x *Expedition) {
	c := w.Civs[x.Owner]
	if c.Species.Profile().Eats {
		// a fleet of a thing that eats has no captains: what holds the worlds is more of it, and stays where it is as a guard, or is lost with it
		if c.Active() && x.Base >= 0 {
			w.mergeInto(x, x.Base)
		} else {
			x.Over = true
		}
		return
	}
	var held []int
	for _, s := range x.Held {
		if w.Owner[s] == c.ID && s != c.Home {
			held = append(held, s)
		}
	}
	if len(held) == 0 {
		x.Over = true
		w.fleetLost(x, max(x.Base, x.Star))
		return
	}
	home := held[len(held)-1]
	for _, s := range held {
		c.Systems = remove(c.Systems, s)
		w.Owner[s] = -1
	}
	w.log("The fleet of the %s never comes home. At %s its captains rule as their own people.", c.Tok(), w.star(home))
	nc := w.spawnCiv(home, c.Species, -1)
	nc.Origin = "the fleet of the " + c.Tok() + " that never came home"
	nc.Master = -1
	for _, s := range held {
		if s != home {
			w.Owner[s] = nc.ID
			nc.Systems = append(nc.Systems, s)
		}
	}
	nc.Peak = len(nc.Systems)
	for _, s := range w.carriedBy(x) {
		w.transfer(s, c, nc)
		s.Carried, s.Star = -1, home
	}
	for _, k := range knownOf(c) {
		nc.Known[k] = true
	}
	w.forget(nc, 0.1)
	w.recompute(nc)
	w.reown(x, nc)
	x.Over = true
	if x.Ships > 0 {
		w.addGuard(nc, home, x.Ships)
	}
	c.Morale -= 1
	c.Tally.Native++
	w.recompute(c)
}

// station is a relief fleet's tick with its host.
func (w *World) station(x *Expedition) {
	c, h := w.Civs[x.Owner], w.Civs[x.Target]
	if k := w.contractOf(x); k != nil {
		if k.State != Running || !h.Active() {
			w.resolve(x)
			return
		}
	} else if !h.Active() || len(h.Wars) == 0 {
		w.resolve(x)
		return
	}
	if !w.canLive(c, x.Base) {
		x.Ships -= w.count(0.03 * float64(x.Ships))
	}
	if x.Ships <= 0 {
		w.resolve(x)
		return
	}
	if w.Owner[x.Base] != h.ID {
		// the world it stood at is lost: it moves to the host's nearest
		t := w.nearestOf(h, x.Base, fleetHop, x.Base)
		if t < 0 {
			w.resolve(x)
			return
		}
		w.sail(x, t)
		x.Withdrawn = true
		return
	}
	if w.wouldTurn(x) {
		w.turn(x)
	}
}

// wouldTurn is whether a relief fleet betrays its host this tick: honour
// sets the base, posture multiplies it, and there has to be an opening.
func (w *World) wouldTurn(x *Expedition) bool {
	c, h := w.Civs[x.Owner], w.Civs[x.Target]
	guard := 0
	if g := w.guardAt(h, x.Base); g != nil && !g.LaidUp {
		guard = g.Ships
	}
	u := mind.Turn(mind.TurnInput{
		Honour: c.honour(), Posture: c.posture(), Betrayed: w.betrayed(h, c),
		HostMil: h.Mil + h.warBonus(), HostShips: float64(guard), Guns: float64(w.gunsAt(h, x.Base)), Relief: float64(w.reliefAt(h, x.Base) - x.Ships),
		Mil: c.Mil + c.warBonus(), Ships: x.Ships, Greed: c.Dials.Greed,
	}, w.Cfg.Tuning)
	return u.Rate > 0 && w.chance(u.Rate)
}

// turn is the betrayal that makes history: the relief fleet turns on the
// world it was sent to keep, and fights for it as any campaign does; a
// host with nothing else in the sky loses it at once.
func (w *World) turn(x *Expedition) {
	c, h := w.Civs[x.Owner], w.Civs[x.Target]
	w.log("The fleet of the %s, sent to keep %s for the %s, takes it for themselves.", c.Tok(), w.star(x.Base), h.Tok())
	w.betray(c, h, "turned on the world they were sent to keep", 1)
	w.breakPacts(c, h)
	h.resent(c.ID, 3)
	x.Kind, x.Turned = Campaign, true
	wr := w.declare(c, h, "betrayal")
	if wr == nil {
		return
	}
	wr.Will[wr.side(c.ID)] += 1
	w.fight(x, x.Base)
}
