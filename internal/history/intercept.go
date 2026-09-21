package history

import (
	"sort"

	"worldgen/internal/battle"
	"worldgen/internal/mind"
)

// Meeting in the dark. A people holding a sighting of a fleet may meet
// it on its line before it arrives: at war with its people, or when it
// is relief bound to stand against this one and this one's posture would
// declare on the sender. An interceptor mustered fifty years after the
// sighting from a guard at a holding reaches the fleet's line where the
// two timetables cross, if they do before the fleet arrives; a fleet
// cannot be chased. The battle is a strike with no world behind either
// side. A fleet that loses the roll cannot go on: it turns into the back
// hemisphere, to its nearest holding within a hop, else any star, else
// where it came from. A fleet that wins keeps its course and its
// timetable and arrives with what it has left. Every ship lost lies
// adrift where it fell.

// Meeting is one battle in the dark, for the batch.
type Meeting struct {
	Year                Year
	Quarry, Interceptor int // fleets
	Owner, Seer         int // peoples
	Ships, Sent         int // the quarry's and the interceptor's ships going in
	Won                 bool
	Broken              bool // the quarry was broken
	Lost                [2]int
}

// considerIntercept is the council meeting on a sighting: feasibility
// first, then the will, then the sizing, and the interceptor sent.
func (w *World) considerIntercept(o *Civ, s *Sighting) {
	x := w.Expeditions[s.Fleet]
	c := w.Civs[x.Owner]
	if x.Over || x.Base >= 0 || x.Launched != s.Leg || !o.Active() || o.Starfaring == 0 || w.allied(o, c) || o.Master == c.ID || c.Master == o.ID {
		return
	}
	tn := w.Cfg.Tuning
	m := max(s.Year+Year(tn.Intercept.Muster), w.Now)
	from, t, ok := w.meetingPoint(o, x, m)
	relief := x.Kind == Relief && x.Target >= 0 && o.Wars[x.Target]
	in := mind.InterceptInput{Feasible: ok, AtWar: o.Wars[c.ID], ReliefAgainst: relief, BoundForUs: w.Owner[x.Star] == o.ID, Sending: w.sendingFleet(o), Posture: o.posture()}
	s.Feasible = ok && (in.AtWar || relief)
	ch := mind.Intercept(in, tn)
	if in.AtWar || relief {
		w.explain(o, "the fleet of the "+c.Name+" sighted", ch)
	}
	if !ch.Try {
		return
	}
	g := w.guardAt(o, from)
	k := w.sizeIntercept(o, c, s, g, t)
	w.explain(o, "meeting the fleet of the "+c.Name, k)
	if !k.Send {
		return
	}
	w.launchIntercept(o, g, x, s, m, t, k.Share)
}

// sendingFleet says whether a people has a campaign or relief fleet in
// flight.
func (w *World) sendingFleet(o *Civ) bool {
	for _, x := range w.fleetsOf(o) {
		if (x.Kind == Campaign || x.Kind == Relief) && x.Base < 0 && !x.Returning {
			return true
		}
	}
	return false
}

// meetingPoint is the earliest year an interceptor sailing at m from one
// of a people's guards reaches the fleet's line before the fleet
// arrives, and the guard's star; the rest of the line is tried at a fixed
// number of points.
func (w *World) meetingPoint(o *Civ, x *Expedition, m Year) (star int, t Year, ok bool) {
	n := w.Cfg.Tuning.Intercept.Samples
	star = -1
	for _, g := range w.fleetsOf(o) {
		if !g.atBase() || g.LaidUp || g.Ships <= 0 || g.Base == x.Star {
			continue // the guard at the fleet's destination fights at the world, not on its doorstep
		}
		h := w.pos(g.Base)
		for k := 0; k < n; k++ {
			y := m + Year(float64(x.Arrive-m)*float64(k)/float64(n))
			if y >= x.Arrive || (ok && y >= t) {
				break
			}
			if float64(m)+h.dist(w.posAtF(x, float64(y)))*o.Speed <= float64(y) {
				star, t, ok = g.Base, y, true
				break
			}
		}
	}
	return
}

// sizeIntercept sizes the interceptor against the fleet seen: its level
// as seen, its owner's war bonus and its ships as levels, from the guard
// that would sail.
func (w *World) sizeIntercept(o, c *Civ, s *Sighting, g *Expedition, t Year) mind.Campaign {
	ships := 0
	if g != nil && !g.LaidUp {
		ships = g.Ships
	}
	return mind.SizeCampaign(mind.CampaignInput{
		Appraisal: mind.Appraisal{Spread: w.Cfg.Tuning.Belief.Spread, Lag: float64(t - w.Now)},
		Strength:  s.Mil + c.warBonus() + mind.ShipLevels(float64(s.Ships)),
		Mil:       o.Mil, Bonus: o.warBonus(), Ships: ships, Total: w.ships(o),
		Risk: o.Dials.Risk, Fear: o.Dials.Fear, Conqueror: o.posture() == mind.Conqueror,
	}, w.Cfg.Tuning)
}

// launchIntercept sends n ships of a guard to meet a fleet at a year on
// its line, sailing at the muster year.
func (w *World) launchIntercept(o *Civ, g *Expedition, x *Expedition, s *Sighting, m, t Year, n int) *Expedition {
	if g == nil || n <= 0 {
		return nil
	}
	n = min(n, g.Ships)
	c := w.Civs[x.Owner]
	at := w.posAtF(x, float64(t))
	y := &Expedition{ID: len(w.Expeditions), Owner: o.ID, Target: x.Owner, Kind: Intercept, Star: x.Star, From: g.Base, Ships: n, Back: -1, Contract: -1, SoldBy: -1,
		Launched: m, Out: m, Arrive: t, Meet: t, Base: -1, Manned: w.Now, Seen: map[int]bool{}, Quarry: x.ID, Leg: x.Launched, Drive: o.Speed,
		Path: &[2]vec{w.pos(g.Base), at}}
	g.Ships -= n
	if g.Ships == 0 && g.Kind == Guard {
		g.Over = true
	}
	w.addExpedition(y)
	s.Intercept = y.ID
	o.Tally.Intercepts++
	w.logAt(m, "The %s send %s from %s to meet the fleet of the %s in the dark.", o.Name, shipsWord(n), w.star(g.Base), c.Name)
	w.timetable(y)
	w.newEye(o, eye{star: -1, r: max(fleetEye, o.watchRange()/2), fleet: y, kind: eyeFleet})
	return y
}

// meetingsDue resolves every meeting due before the next tick, in year
// order, before any fleet arrives.
func (w *World) meetingsDue() {
	until := w.Now + w.Cfg.Step
	var due []*Expedition
	for _, x := range w.liveFleets() {
		if x.Kind == Intercept && x.Base < 0 && !x.Returning && x.Meet < until {
			due = append(due, x)
		}
	}
	sort.SliceStable(due, func(i, j int) bool { return due[i].Meet < due[j].Meet })
	for _, x := range due {
		w.meetInDark(x)
	}
}

// meetInDark is the battle in the dark: the interceptor at the meeting point
// against the fleet, if it is still on the line it was seen on.
func (w *World) meetInDark(x *Expedition) {
	o := w.Civs[x.Owner]
	q := w.Expeditions[x.Quarry]
	c := w.Civs[q.Owner]
	at := w.posAtF(x, float64(x.Meet))
	if !o.Active() {
		x.Over = true
		return
	}
	if q.Over || q.Base >= 0 || q.Launched != x.Leg || q.Ships <= 0 {
		if w.Cfg.TraceAI {
			w.logAt(x.Meet, "[the %s find nothing where the fleet of the %s should have been]", o.Name, c.Name)
		}
		w.homeFrom(x, at, x.Meet)
		return
	}
	wr := w.warBetween(o.ID, c.ID)
	if wr == nil {
		// no war any more: the fleets pass each other
		w.homeFrom(x, at, x.Meet)
		return
	}
	near, far := w.nearerEnd(q, at), q.Star
	if near == far {
		near = q.From
	}
	between := "between " + w.star(near) + " and " + w.star(far)
	atk, def := battle.Strength(x.Ships, w.quality(o)), battle.Strength(q.Ships, w.quality(c))
	won, la, ld := battle.Fight(w.R, atk, def)
	rec := &Meeting{Year: x.Meet, Quarry: q.ID, Interceptor: x.ID, Owner: c.ID, Seer: o.ID, Ships: q.Ships, Sent: x.Ships, Won: won}
	w.Meetings = append(w.Meetings, rec)
	lo := battle.ToShips(w.R, la, w.quality(o), x.Ships)
	lc := battle.ToShips(w.R, ld, w.quality(c), q.Ships)
	x.Ships -= lo
	q.Ships -= lc
	o.Tally.ShipsLost += lo
	c.Tally.ShipsLost += lc
	rec.Lost = [2]int{lo, lc}
	w.leaveField(o, lo, near, at, true)
	w.leaveField(c, lc, near, at, true)
	w.observeDark(o, c)
	w.observeDark(c, o)
	w.sold(q)
	o.Tally.Meetings++
	i := wr.side(o.ID)
	winner, loser := o, c
	if !won {
		winner, loser = c, o
	}
	w.factAt(x.Meet, FIntercept, winner, loser, near)
	w.factAt(x.Meet, FCaught, loser, winner, near)
	switch {
	case q.Ships <= 0:
		rec.Broken = true
		wr.Will[i] += 0.2
		wr.Will[1-i] -= 0.2
		c.Tally.Caught++
		w.logAt(x.Meet, "The %s meet the fleet of the %s %s, and break it. Nothing of it arrives.", o.Name, c.Name, between)
		w.resolve(q)
	case won:
		wr.Will[i] += 0.2
		wr.Will[1-i] -= 0.2
		c.Tally.Caught++
		w.logAt(x.Meet, "The %s meet the fleet of the %s %s, and turn it back.", o.Name, c.Name, between)
		w.turnBack(q, at, x.Meet)
	default:
		wr.Will[i] -= 0.2
		wr.Will[1-i] += 0.2
		weaker := ""
		if lc > 0 {
			weaker = ", " + shareWord(lc, lc+q.Ships) + " weaker"
		}
		w.logAt(x.Meet, "The fleet of the %s, met in the dark %s by the %s, goes on%s.", c.Name, between, o.Name, weaker)
	}
	if x.Ships <= 0 {
		x.Over = true
		w.fleetLost(x, near)
		return
	}
	w.homeFrom(x, at, x.Meet)
}

// turnBack is a fleet beaten in the dark leaving the meeting point into
// the back hemisphere: its people's nearest holding there within a hop,
// else the nearest star of any kind there within a hop, else where it
// came from; and from wherever it lands it turns for home. A horde's
// fleet bases where it lands.
func (w *World) turnBack(q *Expedition, at vec, y Year) {
	c := w.Civs[q.Owner]
	a, b := w.line(q)
	heading := b.sub(a)
	back := func(s int) bool { return w.pos(s).sub(at).dot(heading) < 0 }
	dest, bd := -1, float64(fleetHop)
	for _, s := range w.holdings(c) {
		if d := w.pos(s).dist(at); back(s) && d <= bd {
			dest, bd = s, d
		}
	}
	if dest < 0 {
		bd = fleetHop
		for s := range w.G.Stars {
			if d := w.pos(s).dist(at); back(s) && d <= bd {
				dest, bd = s, d
			}
		}
	}
	if dest < 0 {
		dest = q.From
	}
	q.From = w.nearerEnd(q, at)
	q.Back = -1
	w.legFrom(q, at, dest, y)
	if q.Kind != Roam {
		q.Returning = true
	}
}

// legFrom puts a fleet on a leg from a point to a star, sailing at a year.
func (w *World) legFrom(x *Expedition, at vec, dest int, y Year) {
	c := w.Civs[x.Owner]
	x.Star = dest
	x.Base = -1
	x.Launched = y
	x.Arrive = y + Year(at.dist(w.pos(dest))*c.Speed)
	x.Drive = c.Speed
	x.Path = &[2]vec{at, w.pos(dest)}
}

// homeFrom turns a fleet at a point for its people's nearest holding.
func (w *World) homeFrom(x *Expedition, at vec, y Year) {
	c := w.Civs[x.Owner]
	to, _ := w.nearestTo(c, at)
	x.From = w.nearerEnd(x, at)
	x.Returning = true
	w.legFrom(x, at, to, y)
}
