package history

import "worldgen/internal/mind"

// Exploration. Every star is seen: where it is, its colour, whether it is
// dying. What a star holds, worlds and who is on them, is known only by
// reading it: a surveyor's visit, a colony ship's arrival, a nomad fleet
// basing there, the Sight turned outward, or a telescope a few light years
// out, which shows worlds but not who is on them. Remains are found by a
// visit; the old lottery is what is left for the incurious.

// staleAfter is how long a reading of a star stands before the surveyors
// and the Sight think it worth another look: who is there changes.
const staleAfter = 1_000_000

// fresh says whether a people has read a star lately.
func (w *World) fresh(c *Civ, t int) bool {
	y, ok := c.Charted[t]
	return ok && w.Now-y <= staleAfter
}

// read says whether a people knows what worlds a star holds.
func (w *World) read(c *Civ, t int) bool {
	if _, ok := c.Charted[t]; ok {
		return true
	}
	for _, s := range w.holdings(c) {
		if s == t || w.G.Dist(s, t) <= w.watchAt(c, s) {
			return true
		}
	}
	return false
}

// knownTaken says whether a people knows a star is somebody's: charted,
// its own, or the owner loud enough to hear from one of its holdings.
func (w *World) knownTaken(c *Civ, t int) bool {
	o := w.Owner[t]
	if o < 0 {
		return false
	}
	if o == c.ID {
		return true
	}
	if !w.perceives(c, w.Civs[o]) {
		return false // its holdings read as empty
	}
	if _, ok := c.Charted[t]; ok {
		return true
	}
	if o >= 0 {
		e := w.Civs[o]
		if c.Met[o] && e.signal() > 0 {
			for _, s := range w.holdings(c) {
				if w.G.Dist(s, t) <= e.signal() {
					return true
				}
			}
		}
	}
	return false
}

// chart reads a star for a people. how is who was there: "survey",
// "ship", "settle" or "fleet"; "" is the Sight, which reads without a
// visit. A visit finds what is buried there and meets whoever holds it.
func (w *World) chart(c *Civ, t int, how string) {
	c.Tally.Charted++ // readings, not stars
	c.Charted[t] = w.Now
	visit := how != ""
	if visit {
		delete(c.Marked, t)
		w.readRuins(c, t)
	}
	for _, l := range w.Legacies {
		if l.Star != t || (l.State != Buried && l.State != Sealed) || c.Found[l.ID] {
			continue
		}
		if l.Maker >= 0 && c.Known[l.Node] && l.ships() == 0 {
			continue // nothing to learn; a structure is taken over on settling; a field with ships is worth the ships
		}
		if l.People == c.ID {
			continue // a sleeper does not find itself
		}
		if !visit {
			if !c.Marked[t] {
				c.Marked[t] = true
				if w.R.Float64() < 0.3 {
					w.log("The Sight shows the %s something at %s that nobody made in this age. They mean to go and see.", c.Name, w.star(t))
				}
			}
			continue
		}
		w.discover(c, l, how)
		if !c.Active() {
			return
		}
	}
	if !visit || !c.Active() {
		return
	}
	if w.lurks(t) && how == "survey" && w.R.Float64() < 0.3 {
		w.log("Surveyors of the %s find %s held by something that is not a people as they know one, and do not go closer.", c.Name, w.star(t))
	}
	if o := w.Owner[t]; o >= 0 && o != c.ID {
		e := w.Civs[o]
		if e.Active() && !(c.Reached[o] && e.Reached[c.ID]) {
			switch how {
			case "survey":
				c.Tally.MetSurvey++
			case "ship":
				c.Tally.MetShip++
			}
			w.meet(c, e, t)
		}
	}
}

// explore is the tick step: the Sight's mode and reading, then the surveyors.
func (w *World) explore(c *Civ) {
	if !c.Active() {
		return
	}
	w.sightMode(c)
	if c.miracle("foresight") {
		c.Tally.Sighted += w.dt
	}
	if c.Searching {
		// five surveyors' worth: a couple of stars a millennium, nearest
		// first, out to twice the reach, since the Sight does not travel;
		// the never-read before the stale
		c.Tally.Searched += w.dt
		n := max(1, w.count(w.Cfg.Tuning.Sight.Reads))
		near := w.G.Near(c.Home, w.sightRange(c))
		for pass := 0; pass < 2 && n > 0; pass++ {
			for _, t := range near {
				if n == 0 {
					break
				}
				if _, ok := c.Charted[t]; ok == (pass == 0) || w.fresh(c, t) {
					continue
				}
				w.chart(c, t, "")
				n--
			}
		}
	}
	w.survey(c)
	w.picket(c)
}

// sightMode turns the Sight outward when nothing threatens: it reads the
// unvisited stars, and while it does it is not watching the borders. A war,
// or for the fearful a hostile neighbour in reach, turns it back.
func (w *World) sightMode(c *Civ) {
	if !c.miracle("foresight") {
		c.Searching = false
		return
	}
	menaced := func() bool {
		for _, eid := range sortedInts(c.Met) {
			e := w.Civs[eid]
			if e.Active() && e.Free() && mind.Menaces(e.hostile(), w.monster(c, e), c.Grudge[eid], w.Cfg.Tuning) && w.inReach(e, c.Home) {
				return true
			}
		}
		return false
	}
	unread := func() bool { return w.unread(c, w.sightRange(c)) } // nothing left to look for; it comes back when reach grows
	m := mind.Sight(mind.SightInput{AtWar: len(c.Wars) > 0, Fear: c.Dials.Fear, Menaced: menaced, Unread: unread}, w.Cfg.Tuning)
	if m.Outward == c.Searching {
		return
	}
	w.explain(c, "the Sight", m)
	c.Searching = m.Outward
	switch {
	case m.Outward && c.Tally.Searched == 0:
		w.log("The %s turn the Sight outward, to the stars nobody has visited.", c.Name)
	case !m.Outward && m.Threat:
		w.log("The %s turn the Sight back to their own borders.", c.Name)
	}
}

// neverRead says whether any star within reach has never been read.
func (w *World) neverRead(c *Civ) bool {
	for _, t := range w.G.Near(c.Home, max(c.Reach, w.Cfg.Tuning.Survey.NearMin)) {
		if _, ok := c.Charted[t]; !ok && !w.read(c, t) {
			return true
		}
	}
	return false
}

// sightRange is how far the Sight reads: twice the reach, at least a little.
func (w *World) sightRange(c *Civ) float64 { return mind.SightRange(c.Reach, w.Cfg.Tuning) }

// unread says whether any star within a range is unread or stale.
func (w *World) unread(c *Civ, within float64) bool {
	for _, t := range w.G.Near(c.Home, within) {
		if !w.fresh(c, t) {
			return true
		}
	}
	return false
}

// survey keeps surveyors out among the stars: as many as hunger and greed
// want and the level can spare, none in wartime, at least one when there
// is nothing read to settle and stars still unread.
func (w *World) survey(c *Civ) {
	tn := w.Cfg.Tuning
	want := mind.Survey(mind.SurveyInput{
		Free: c.Free() && !c.Aloft, Reach: c.Reach, Ships: w.standing(c), AtWar: len(c.Wars) > 0, Era: c.Era, Dials: c.Dials,
		ToSettle:  func() bool { return w.anyToSettle(c) },
		Unread:    func() bool { return w.unread(c, max(c.Reach, tn.Survey.NearMin)) },
		NeverRead: func() bool { return w.neverRead(c) },
	}, tn)
	c.SurveyWant = want.Asked
	out := 0
	for _, x := range w.fleetsOf(c) {
		if x.Kind == Survey {
			out++
		}
	}
	if out >= want.Want || !w.chance(tn.Survey.Rate) {
		return
	}
	t := w.surveyTarget(c, -1)
	if t < 0 {
		return
	}
	w.explain(c, "surveying", want)
	x := w.launch(c, Survey, nil, t, 1)
	if x == nil {
		return
	}
	c.Tally.Surveys++
	if c.Tally.Surveys == 1 {
		w.log("The %s send their first surveyors out: a ship of a few, bound for %s, to see what the stars hold.", c.Name, w.star(t))
	} else if w.Cfg.TraceAI {
		w.log("[the %s send surveyors to %s, %s away]", c.Name, w.star(t), span(x.Arrive-w.Now))
	}
}

// anyToSettle says whether a read, free, livable star lies within reach.
func (w *World) anyToSettle(c *Civ) bool {
	for _, t := range w.G.Near(c.Home, c.Reach) {
		if w.read(c, t) && !w.knownTaken(c, t) && w.canLive(c, t) {
			return true
		}
	}
	return false
}

// surveyTarget is the next star for surveyors: from a star, the nearest
// unread within a hop; from home, the nearest unread within reach. A star
// the Sight marked comes first. Never one another survey is bound for.
func (w *World) surveyTarget(c *Civ, from int) int {
	origin, within := c.Home, c.Reach
	if from >= 0 {
		origin, within = from, mind.SurveyHop(c.Reach, c.miracle("ftl"), w.Cfg.Tuning)
	}
	var stars []mind.Star
	for _, t := range w.G.Near(origin, within) {
		if w.G.Dist(c.Home, t) > c.Reach || w.surveyBound(c, t) || w.dread(c, t) {
			continue
		}
		_, charted := c.Charted[t]
		stars = append(stars, mind.Star{ID: t, Marked: c.Marked[t], Charted: charted, Fresh: w.fresh(c, t)})
	}
	return mind.SurveyTarget(stars)
}

// surveyBound says whether surveyors of a people are already on their way to a star.
func (w *World) surveyBound(c *Civ, t int) bool {
	for _, x := range w.fleetsOf(c) {
		if x.Kind == Survey && !x.Returning && x.Star == t {
			return true
		}
	}
	return false
}

// surveyArrive is surveyors reaching a star: read it, and go on to the
// next, or home after a tour, or when called back by a war. A star held by
// a monster keeps half the ships that reach it.
func (w *World) surveyArrive(x *Expedition) {
	c := w.Civs[x.Owner]
	t := x.Star
	w.chart(c, t, "survey")
	if !c.Active() {
		x.Over = true
		return
	}
	if w.lurks(t) && w.R.Float64() < 0.5 {
		o := w.Civs[w.Owner[t]]
		w.log("The surveyors of the %s do not come back from %s. What they sent before the end says enough: the %s are there.", c.Name, w.star(t), o.Name)
		w.fact(FSurveyLost, c, o, t)
		c.Morale -= 0.3
		x.Over = true
		w.fleetLost(x, t)
		return
	}
	if o := w.Owner[t]; o >= 0 && o != c.ID && w.Civs[o].Active() && !w.perceives(c, w.Civs[o]) && w.R.Float64() < 0.5 {
		// a world of what cannot be held in mind: the surveyors read it as empty, or do not come back, and either way the record has nothing in it
		w.log("The surveyors of the %s do not come back from %s. What they sent before the end says the star is empty. The %s are there.", c.Name, w.star(t), w.Civs[o].Name)
		w.fact(FSurveyLost, c, w.Civs[o], t)
		c.Morale -= 0.3
		x.Over = true
		w.fleetLost(x, t)
		return
	}
	x.Tour++
	next := -1
	if mind.TourOn(x.Tour, float64(w.Now-x.Launched), x.Recalled, w.Cfg.Tuning) {
		next = w.surveyTarget(c, t)
	}
	if next < 0 {
		w.goHome(x)
		return
	}
	x.From, x.Star = t, next
	x.Arrive = w.Now + Year(w.G.Dist(t, next)*c.Speed)
}

// recallSurveys calls a people's surveyors home at the outbreak of a war.
func (w *World) recallSurveys(c *Civ) {
	for _, x := range w.fleetsOf(c) {
		if x.Kind == Survey && !x.Returning {
			x.Recalled = true
		}
	}
}

// picket keeps scouts out watching: one per enemy the policy names, at
// the star nearest the midpoint between the people's nearest holding and
// the enemy's nearest world, if that star is empty or its own. A picket
// holds its post for a tour and comes home. Wartime does not recall it.
func (w *World) picket(c *Civ) {
	tn := w.Cfg.Tuning
	if c.Aloft || c.Starfaring == 0 || !c.Free() || !w.chance(tn.Picket.Rate) {
		return
	}
	for _, eid := range sortedInts(c.Met) {
		e := w.Civs[eid]
		if !e.Active() || w.allied(c, e) || c.Master == e.ID || e.Master == c.ID || w.picketAgainst(c, e) != nil {
			continue
		}
		want := mind.WantPicket(mind.PicketInput{
			AtWar: c.Wars[eid], Front: len(w.front(c, e)) > 0, Caught: c.Tally.Caught > 0,
			Hostile: e.hostile(), InReach: w.inReach(e, c.Home), Fear: c.Dials.Fear, Ships: w.standing(c),
		}, tn)
		if !want.Send {
			continue
		}
		t := w.picketStar(c, e)
		if t < 0 {
			continue
		}
		w.explain(c, "a picket against the "+e.Name, want)
		x := w.launch(c, Scout, e, t, 1)
		if x == nil {
			return
		}
		x.Picket = true
		c.Tally.Pickets++
		if c.Tally.Pickets == 1 || w.Cfg.TraceAI {
			w.log("The %s send a ship to %s to sit and watch the sky toward the %s.", c.Name, w.star(t), e.Name)
		}
		return
	}
}

// picketAgainst is a people's picket sent against an enemy, or nil.
func (w *World) picketAgainst(c, e *Civ) *Expedition {
	for _, x := range w.fleetsOf(c) {
		if x.Kind == Scout && x.Picket && x.Target == e.ID && !x.Returning {
			return x
		}
	}
	return nil
}

// picketStar is where a picket against an enemy sits: the star nearest
// the midpoint between the people's nearest holding and the enemy's
// nearest world, empty or the people's own, within a hop of the midpoint.
func (w *World) picketStar(c, e *Civ) int {
	_, ew := w.nearestEnemy(c, e)
	h, _ := w.nearest(c, ew)
	mid := w.pos(h).lerp(w.pos(ew), 0.5)
	best, bd := -1, float64(fleetHop)
	for s := range w.G.Stars {
		if w.Owner[s] >= 0 && w.Owner[s] != c.ID {
			continue
		}
		if d := w.pos(s).dist(mid); d < bd {
			best, bd = s, d
		}
	}
	return best
}

// picketPost is a picket arriving: it stays, and it is an eye from now on.
func (w *World) picketPost(x *Expedition) {
	c := w.Civs[x.Owner]
	x.Base, x.Fed = x.Star, w.Now
	w.newEye(c, eye{star: x.Base, r: max(fleetEye, c.watchRange()/2), kind: eyePicket})
}

// picketStep is a picket at its post: home when its tour is done.
func (w *World) picketStep(x *Expedition) {
	if !x.Picket || float64(w.Now-x.Fed) < w.Cfg.Tuning.Picket.Tour {
		return
	}
	w.goHome(x)
}

// lurks says whether a star is held by something surveyors do not come
// back from: a people that is a monster by its nature, awake or asleep.
func (w *World) lurks(t int) bool {
	o := w.Owner[t]
	return o >= 0 && w.Civs[o].Species.Profile().Monster
}
