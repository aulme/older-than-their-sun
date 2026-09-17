package history

import (
	"math"

	"worldgen/internal/mind"
	"worldgen/internal/names"
	"worldgen/internal/tech"
)

// An expedition is the army far from home: a share of a people's Military
// that leaves for the duration, crosses at ship speed, and fights from a
// base in enemy territory with no reinforcement. A relief fleet is the
// same thing sent to stand with an ally. A scout is a fleet of one level
// that only looks.

// ExpKind is what a fleet is for.
type ExpKind uint8

const (
	Campaign ExpKind = iota
	Relief
	Scout
	Roam   // a nomad people's fleet; see nomad.go
	Survey // surveyors reading the stars; see explore.go
)

func (k ExpKind) String() string {
	return [...]string{"campaign", "relief", "scout", "roam", "survey"}[k]
}

// Expedition is one fleet.
type Expedition struct {
	ID        int
	Owner     int
	Target    int // the people fought, or stood with, or looked at; -1 if none
	Kind      ExpKind
	Star      int // destination
	From      int // where it set out from
	Mil       float64
	Launched  Year
	Arrive    Year
	Base      int // where it operates from; -1 in flight
	Returning bool
	Over      bool
	Held      []int // worlds it took and its people still hold
	Seen      map[int]bool
	Report    *Intel // a scout's report, carried home without the Voice
	Battles   int
	Wins      int
	Turned    bool
	Fed       Year // for a roaming fleet: when it reached its base
	Tour      int  // for surveyors: stars read this trip
	Recalled  bool // for surveyors: called home by a war; they finish the leg and turn
}

// launch sends a fleet. The strength leaves the home level at once.
func (w *World) launch(c *Civ, kind ExpKind, target *Civ, star int, mil float64) *Expedition {
	from, d := w.nearest(c, star)
	x := &Expedition{ID: len(w.Expeditions), Owner: c.ID, Target: -1, Kind: kind, Star: star, From: from, Mil: mil,
		Launched: w.Now, Arrive: w.Now + Year(d*c.Speed), Base: -1, Seen: map[int]bool{}}
	if target != nil {
		x.Target = target.ID
	}
	total := c.Mil + c.Away
	c.Away += mil
	w.recompute(c)
	w.Expeditions = append(w.Expeditions, x)
	switch kind {
	case Campaign:
		c.Tally.Fleets++
		w.log("The %s send %s of their strength against the %s: a fleet bound for %s, %s away.", c.Name, shareWord(mil, total), target.Name, w.star(star), span(x.Arrive-w.Now))
	case Relief:
		c.Tally.Relief++
		w.log("The %s send a fleet to stand with the %s at %s, %s away.", c.Name, target.Name, w.star(star), span(x.Arrive-w.Now))
	case Scout:
		c.Tally.Scouts++
		if w.Cfg.TraceAI {
			w.log("[the %s send a scout to %s, %s away]", c.Name, w.star(star), span(x.Arrive-w.Now))
		}
	case Survey:
		// logged by survey, which knows whether it is the first
	}
	return x
}

func shareWord(mil, total float64) string {
	switch f := mil / max(total, 0.01); {
	case f >= 0.6:
		return "most"
	case f >= 0.4:
		return "half"
	case f >= 0.25:
		return "a third"
	}
	return "a part"
}

// tickExpeditions moves every fleet a tick.
func (w *World) tickExpeditions() {
	for _, x := range w.Expeditions {
		if x.Over {
			continue
		}
		c := w.Civs[x.Owner]
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
		if x.Returning {
			if w.Now >= x.Arrive {
				c.Away = max(0, c.Away-x.Mil)
				if x.Report != nil && x.Target >= 0 {
					if c.receive(x.Target, x.Report) {
						c.Scouted[x.Target] = w.Now
						delete(c.Watched, x.Target)
						w.forward(c, w.Civs[x.Target], x.Report)
					}
					c.Summoned = true
				}
				x.Over = true
				w.recompute(c)
			}
			continue
		}
		if x.Base < 0 {
			w.watchSky(x)
			if w.Now < x.Arrive {
				continue
			}
			w.arrive(x)
			continue
		}
		switch x.Kind {
		case Campaign:
			w.campaign(x)
		case Relief:
			w.station(x)
		}
	}
}

// arrive is a fleet reaching its destination.
func (w *World) arrive(x *Expedition) {
	c := w.Civs[x.Owner]
	switch x.Kind {
	case Survey:
		w.surveyArrive(x)
	case Scout:
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
			w.log("The fleet of the %s arrives at %s to find the war over.", c.Name, w.star(x.Star))
			w.resolve(x)
			return
		}
		x.Base = x.Star
		t := w.campaignTarget(x, e)
		if t < 0 {
			w.log("The fleet of the %s arrives at %s to find nothing of the %s left to fight.", c.Name, w.star(x.Star), e.Name)
			w.resolve(x)
			return
		}
		x.Base = t
		w.log("The fleet of the %s arrives at %s, %s after it set out.", c.Name, w.star(t), span(w.Now-x.Launched))
	case Relief:
		h := w.Civs[x.Target]
		if !h.Active() || len(h.Wars) == 0 {
			w.log("The fleet of the %s arrives at %s to find the war over.", c.Name, w.star(x.Star))
			w.resolve(x)
			return
		}
		x.Base = x.Star
		if w.Owner[x.Base] != h.ID {
			x.Base = w.nearestOf(h, x.Star, 20)
			if x.Base < 0 {
				w.resolve(x)
				return
			}
		}
		w.faith(c, h, 0.5)
		w.log("A fleet of the %s arrives at %s to stand with the %s.", c.Name, w.star(x.Base), h.Name)
		w.fact(FRelief, c, h, x.Base)
	}
}

// nearestOf is e's world nearest a star within a range, or -1.
func (w *World) nearestOf(e *Civ, star int, within float64) int {
	best, bd := -1, within
	for _, s := range e.Systems {
		if d := w.G.Dist(star, s); d <= bd {
			best, bd = s, d
		}
	}
	return best
}

// campaignTarget is what a fleet fights next: its base if the enemy still
// holds it, else the enemy's nearest world within a fleet's hop.
func (w *World) campaignTarget(x *Expedition, e *Civ) int {
	if w.Owner[x.Base] == e.ID {
		return x.Base
	}
	return w.nearestOf(e, x.Base, 20)
}

// campaign is a fleet's tick in enemy territory: attrition, then battles.
func (w *World) campaign(x *Expedition) {
	c, e := w.Civs[x.Owner], w.Civs[x.Target]
	wr := w.warBetween(c.ID, e.ID)
	if !e.Active() || wr == nil {
		w.resolve(x)
		return
	}
	if !w.canLive(c, x.Base) {
		x.Mil *= 1 - 0.03*w.dt
	}
	if x.Mil < 1 {
		w.log("The fleet of the %s wastes away at %s, far from anything it could live on.", c.Name, w.star(x.Base))
		w.resolve(x)
		return
	}
	for n := min(w.count(1), 3); n > 0 && !x.Over; n-- {
		t := w.campaignTarget(x, e)
		if t < 0 {
			w.resolve(x)
			return
		}
		x.Battles++
		i := wr.side(c.ID)
		atk := x.Mil + c.warBonus() + w.R.NormFloat64()*1.5
		def := w.defence(e, t) + w.R.NormFloat64()*1.5
		w.observe(c, e, t, 0.3)
		if atk < def {
			x.Mil *= 0.7
			wr.Will[i] -= 0.1
			wr.Will[1-i] += 0.1
			if x.Mil < 1 {
				w.log("The fleet of the %s is broken at %s.", c.Name, w.star(t))
				w.fact(FDefeat, c, e, t)
				w.resolve(x)
				return
			}
			continue
		}
		x.Wins++
		w.takeWorld(wr, c, e, t)
		if w.Owner[t] == c.ID {
			x.Held = append(x.Held, t)
			x.Base = t
		}
		if wr.Over || !e.Active() {
			w.resolve(x)
			return
		}
	}
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
		away := float64(w.Now - x.Launched)
		p := clamp(0.15+away/60_000+w.G.Dist(x.Base, c.Home)/80, 0, 0.9)
		if !c.Active() || w.R.Float64() < p {
			w.goNative(x)
			return
		}
	}
	if x.Mil < 1 && len(held) == 0 {
		c.Away = max(0, c.Away-x.Mil)
		c.Morale -= 0.5
		x.Over = true
		w.recompute(c)
		return
	}
	w.goHome(x)
}

// goHome turns a fleet for the nearest holding.
func (w *World) goHome(x *Expedition) {
	c := w.Civs[x.Owner]
	from := x.Base
	if from < 0 {
		from = x.Star
	}
	to, d := w.nearest(c, from)
	x.Returning = true
	x.Base = -1
	x.From, x.Star = from, to
	x.Arrive = w.Now + Year(d*c.Speed)
}

// goNative is a fleet that never comes home: its captains keep what they
// hold as a people of their own, a successor state with the parent's tree.
func (w *World) goNative(x *Expedition) {
	c := w.Civs[x.Owner]
	var held []int
	for _, s := range x.Held {
		if w.Owner[s] == c.ID && s != c.Home {
			held = append(held, s)
		}
	}
	if len(held) == 0 {
		x.Over = true
		c.Away = max(0, c.Away-x.Mil)
		return
	}
	home := held[len(held)-1]
	for _, s := range held {
		c.Systems = remove(c.Systems, s)
		w.Owner[s] = -1
	}
	w.log("The fleet of the %s never comes home. At %s its captains rule as their own people.", c.Name, w.star(home))
	nc := w.spawnCiv(home, c.Species, -1, names.Civ(w.R))
	nc.Origin = "the fleet of the " + c.Name + " that never came home"
	nc.Master = -1
	for _, s := range held {
		if s != home {
			w.Owner[s] = nc.ID
			nc.Systems = append(nc.Systems, s)
		}
	}
	nc.Peak = len(nc.Systems)
	for _, k := range knownOf(c) {
		nc.Known[k] = true
	}
	w.forget(nc, 0.1)
	w.recompute(nc)
	c.Away = max(0, c.Away-x.Mil)
	c.Morale -= 1
	c.Tally.Native++
	w.recompute(c)
	x.Over = true
}

// station is a relief fleet's tick with its host.
func (w *World) station(x *Expedition) {
	c, h := w.Civs[x.Owner], w.Civs[x.Target]
	if !h.Active() || len(h.Wars) == 0 {
		w.resolve(x)
		return
	}
	if !w.canLive(c, x.Base) {
		x.Mil *= 1 - 0.03*w.dt
	}
	if x.Mil < 1 {
		w.resolve(x)
		return
	}
	if w.Owner[x.Base] != h.ID {
		x.Base = w.nearestOf(h, x.Base, 20)
		if x.Base < 0 {
			w.resolve(x)
			return
		}
	}
	if w.wouldTurn(x) {
		w.turn(x)
	}
}

// reliefAt is the strength of others standing with a people at a world.
func (w *World) reliefAt(h *Civ, t int) float64 {
	r := 0.0
	for _, x := range w.Expeditions {
		if x.Over || x.Kind != Relief || x.Base < 0 || x.Target != h.ID || x.Returning {
			continue
		}
		if w.G.Dist(x.Base, t) <= 20 {
			r += x.Mil
		}
	}
	return r
}

// watchSky is everyone with a telescope seeing a fleet pass.
func (w *World) watchSky(x *Expedition) {
	if x.Kind == Scout || x.Kind == Roam || x.Kind == Survey {
		return
	}
	c := w.Civs[x.Owner]
	frac := clamp(float64(w.Now-x.Launched)/float64(max(1, x.Arrive-x.Launched)), 0, 1)
	a, b := &w.G.Stars[x.From], &w.G.Stars[x.Star]
	px, py, pz := a.X+frac*(b.X-a.X), a.Y+frac*(b.Y-a.Y), a.Z+frac*(b.Z-a.Z)
	for _, o := range w.Civs {
		if !o.Active() || o.ID == c.ID || x.Seen[o.ID] {
			continue
		}
		r := o.watchRange()
		if c.Speed <= 4 {
			r *= 2 // a relativistic drive is a torch
		}
		seen := o.miracle("foresight") && !o.Searching && w.Owner[x.Star] == o.ID
		if !seen && r > 0 {
			for _, s := range o.Systems {
				st := &w.G.Stars[s]
				if math.Sqrt((st.X-px)*(st.X-px)+(st.Y-py)*(st.Y-py)+(st.Z-pz)*(st.Z-pz)) <= r {
					seen = true
					break
				}
			}
		}
		if seen {
			x.Seen[o.ID] = true
			w.fleetSeen(x, o)
		}
	}
}

// fleetSeen is what a people does with a fleet sighted.
func (w *World) fleetSeen(x *Expedition, o *Civ) {
	c := w.Civs[x.Owner]
	switch {
	case x.Kind == Campaign && x.Target == o.ID:
		w.log("The %s see the fleet of the %s coming, %s out.", o.Name, c.Name, span(x.Arrive-w.Now))
		o.Focus[tech.Weapons] = max(o.Focus[tech.Weapons], 3)
		o.Summoned = true
		if wr := w.warBetween(c.ID, o.ID); wr != nil {
			wr.Will[wr.side(o.ID)] += 0.5
			w.callAllies(o, c, wr)
		}
	case x.Kind == Relief && x.Target >= 0 && o.Wars[x.Target]:
		h := w.Civs[x.Target]
		i := w.observe(o, h, x.Star, 0.5)
		i.Relief += x.Mil
	}
}

// wouldTurn is whether a relief fleet betrays its host this tick: honour
// sets the base, posture multiplies it, and there has to be an opening.
func (w *World) wouldTurn(x *Expedition) bool {
	c, h := w.Civs[x.Owner], w.Civs[x.Target]
	u := mind.Turn(mind.TurnInput{
		Honour: c.honour(), Posture: c.posture(), Betrayed: w.betrayed(h, c),
		HostMil: h.Mil + h.warBonus(), Relief: w.reliefAt(h, x.Base), Mil: x.Mil,
		AtHome: x.Base == h.Home, Grid: h.Known["defence_grid"], Greed: c.Dials.Greed,
	}, w.Cfg.Tuning)
	return u.Rate > 0 && w.chance(u.Rate)
}

// turn is the betrayal that makes history: the relief fleet seizes the
// world it was sent to keep.
func (w *World) turn(x *Expedition) {
	c, h := w.Civs[x.Owner], w.Civs[x.Target]
	w.log("The fleet of the %s, sent to keep %s for the %s, takes it for themselves.", c.Name, w.star(x.Base), h.Name)
	w.betray(c, h, "turned on the world they were sent to keep", 1)
	w.breakPacts(c, h)
	h.Grudge[c.ID] += 3
	x.Kind, x.Turned = Campaign, true
	wr := w.declare(c, h, "betrayal")
	if wr == nil {
		return
	}
	wr.Will[wr.side(c.ID)] += 1
	def := h.Mil + h.warBonus() + 1 + w.R.NormFloat64()*1.5
	if x.Base == h.Home {
		def += 2.5
	}
	if h.Known["defence_grid"] {
		def += 0.5
	}
	if x.Mil+c.warBonus()+w.R.NormFloat64()*1.5 >= def {
		w.takeWorld(wr, c, h, x.Base)
		if w.Owner[x.Base] == c.ID {
			x.Held = append(x.Held, x.Base)
		}
	} else {
		x.Mil *= 0.7
	}
}
