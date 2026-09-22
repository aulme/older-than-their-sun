package history

import (
	"sort"

	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// A war is an object with a front, a duration and a will per side. The
// front is the overlap: each side's worlds within the other's reach, where
// a campaign can be sent and what the appraisal looks at. Worlds change
// hands only by a fleet at them (battle.go). Will drains with time, twice
// as fast while nobody's fleet is against the other, and moves with each
// world taken; when it runs out there is peace, or capitulation, and a
// record that makes the next war between the two different from the
// first.

// War is one war between two peoples.
type War struct {
	ID        int
	Sides     [2]int
	Began     Year
	Ended     Year
	Over      bool
	Cause     string // why it was declared: a key of data/causes.json's war_causes; CauseOf the people the cause names, or -1
	CauseOf   int
	Named     int // the star the war is named for, or -1; the name itself is a row of the names pass
	Nth       int // the nth war between these two
	Will      [2]float64
	Taken     [2]int // worlds taken by each side
	Glassed   [2]int // worlds destroyed by each side
	Lost      [2]int // worlds lost by each side
	Result    string // how it ended: a key of data/causes.json's war_results
	Pact      int    // pact this war was joined under, -1
	Principal int    // the ally whose war this is, -1
	Contested map[int]int
	Called    map[int]bool // allies already called to this war
	Hire      int          // the contract this war was declared for, or -1; see contract.go
	Gap       *Gap         // set for a hunt: the first side fights a region, not a people it can name; see gap.go
	// the slights of the war: see slight.go
	Slights    map[int]float64 // the slight each partner of the target took at the declaration
	Sent       map[int]float64 // what the target was sending each the tick before
	Slighted   map[int]float64 // what each has taken in all, capped
	SlightTold map[int]bool    // the fact was written
}

func (wr *War) side(id int) int {
	if wr.Sides[0] == id {
		return 0
	}
	return 1
}

// warBetween finds the active war between two peoples, or nil.
func (w *World) warBetween(a, b int) *War {
	for i := len(w.Wars) - 1; i >= 0; i-- {
		wr := w.Wars[i]
		if !wr.Over && ((wr.Sides[0] == a && wr.Sides[1] == b) || (wr.Sides[0] == b && wr.Sides[1] == a)) {
			return wr
		}
	}
	return nil
}

// front is e's worlds within c's reach measured from c's nearest holding,
// nearest first. Only these can be struck.
func (w *World) front(c, e *Civ) []int {
	if c.Reach < 1 {
		return nil
	}
	type fw struct {
		s int
		d float64
	}
	var out []fw
	reach := c.Reach
	if c.Aloft {
		reach = min(max(c.Reach, 3), 20) // a hop from a fleet
	}
	for _, s := range w.holdings(e) {
		_, d := w.nearest(c, s)
		if d <= reach {
			out = append(out, fw{s, d})
		}
	}
	// nearest first; the home last whatever its distance, since everything
	// between defends it
	sort.SliceStable(out, func(i, j int) bool {
		if (out[i].s == e.Home) != (out[j].s == e.Home) {
			return out[j].s == e.Home
		}
		return out[i].d < out[j].d
	})
	res := make([]int, len(out))
	for i, f := range out {
		res[i] = f.s
	}
	return res
}

// inReach says whether c can strike a star from any holding.
func (w *World) inReach(c *Civ, star int) bool {
	_, d := w.nearest(c, star)
	return c.Reach >= 1 && d <= c.Reach
}

// initialWill is what a side brings to a war at its start.
func (w *World) initialWill(c, e *Civ, attacker bool) float64 {
	will := 0.0
	switch c.posture() {
	case "pacifist":
		will = 0.3
	case "defensive":
		will = 0.6
	case "submissive":
		will = 0.2
	case "opportunist":
		will = 0.8
		if attacker {
			will = 1.2
		}
	case "conqueror":
		will = 2
	case "vengeful":
		will = 1 + 0.5*c.Grudge[e.ID]
	case "confederate":
		will = 0.7
	case "unyielding":
		will = 3
	}
	if c.hates(e) {
		will += 1.5
	}
	if w.monster(c, e) {
		will += 1 // what is remembered of them
	}
	if !attacker && w.inReach(e, c.Home) {
		will += 0.7 // the home is at stake
	}
	if c.Aloft {
		will *= 0.5 // a horde's peace is leaving
	}
	return will
}

// declare opens a war, or returns the one already running.
func (w *World) declare(c, e *Civ, cause reason) *War {
	if wr := w.warBetween(c.ID, e.ID); wr != nil {
		return wr
	}
	if !c.Active() || !e.Active() || c == e {
		return nil
	}
	if e.Asleep {
		w.rouse(e, nil) // a war brought to a sleeper wakes it; the strike it wakes with is its own council's
	}
	c.Fought[e.ID]++
	e.Fought[c.ID]++
	wr := &War{ID: len(w.Wars), Sides: [2]int{c.ID, e.ID}, Began: w.Now, Cause: cause.key, CauseOf: cause.by, Nth: c.Fought[e.ID], Named: -1, Pact: -1, Principal: -1, Hire: -1, Contested: map[int]int{}, Called: map[int]bool{},
		Slights: map[int]float64{}, Sent: map[int]float64{}, Slighted: map[int]float64{}, SlightTold: map[int]bool{}}
	wr.Will = [2]float64{w.initialWill(c, e, true), w.initialWill(e, c, false)}
	w.addWar(wr)
	c.Wars[e.ID], e.Wars[c.ID] = true, true
	w.stir(c) // a war is not still
	w.stir(e)
	w.cutTrade(c, e, because("war"))
	w.cutTrade(e, c, because("war"))
	delete(c.Trade, e.ID)
	delete(e.Trade, c.ID)
	delete(c.Watched, e.ID)
	w.recallSurveys(c)
	w.recallSurveys(e)
	c.Tally.Declared++
	c.Tally.Fought++
	e.Tally.Fought++
	c.Focus[tech.Weapons] = max(c.Focus[tech.Weapons], 2)
	e.Focus[tech.Weapons] = max(e.Focus[tech.Weapons], 2)
	w.observe(c, e, e.Home, 0.5)
	w.observe(e, c, c.Home, 0.5)
	declared := w.unplaced(FWar, c, e, -1).with(cause.params("cause")).with(P{"war": wr.ID, "nth": wr.Nth, "hunt": !w.perceives(c, e), "unseen": !w.perceives(e, c)})
	w.slighted(c, e, wr)
	w.place(declared) // told after the slight it gives
	w.callAllies(e, c, wr)
	w.joinAllies(c, e, wr)
	return wr
}

func ordinal(n int) string {
	switch n {
	case 2:
		return "second"
	case 3:
		return "third"
	case 4:
		return "fourth"
	case 5:
		return "fifth"
	}
	return sprintf("%dth", n)
}

// tickWars runs every war: will drains and peace is judged. The fighting
// is the fleets' own tick.
func (w *World) tickWars() {
	for _, wr := range w.Wars {
		if wr.Over {
			continue
		}
		a, b := w.Civs[wr.Sides[0]], w.Civs[wr.Sides[1]]
		if !a.Active() || !b.Active() {
			w.endWar(wr, "fall")
			continue
		}
		for i := 0; i < 2 && !wr.Over; i++ {
			w.drain(wr, i)
		}
		if wr.Over {
			continue
		}
		w.slightTick(wr)
		w.judge(wr)
	}
}

// drain is what a tick of war costs a side's will.
func (w *World) drain(wr *War, i int) {
	c, e := w.Civs[wr.Sides[i]], w.Civs[wr.Sides[1-i]]
	d := 0.05 * w.dt
	if w.Now-wr.Began > 50_000 {
		d *= 2
	}
	switch c.posture() {
	case "unyielding":
		d = 0
	case "vengeful":
		d *= 0.5
	case "conqueror":
		if wr.Taken[i] > wr.Lost[i] {
			d *= 0.7
		}
	case "pacifist":
		if !w.inReach(e, c.Home) {
			wr.Will[i] = 0 // nothing worth fighting for
			return
		}
	}
	if c.hates(e) {
		d *= 0.25
	}
	switch {
	case w.fleetInFlight(c, e) || w.fleetInFlight(e, c):
		d = 0 // a fleet is on its way; nobody tires of a war that has not begun
	case w.hasFleetAgainst(c, e) || w.hasFleetAgainst(e, c):
		d *= 0.5
	default:
		d *= 2 // nobody's fleet is against the other: a war nobody sends ships to ends quickly
	}
	wr.Will[i] -= d
}

// takeWorld is a world with nothing left in its sky to hold it: conquered,
// glassed or converted, by the winner's nature; the home falling is its
// own matter.
func (w *World) takeWorld(wr *War, c, e *Civ, t int) {
	w.sourceSlight(wr, c, e, t)
	c.LastTaken, e.LastTaken = w.Now, w.Now
	w.expose(e, c, "occupation")
	w.expose(c, e, "occupation")
	if c.Species.Profile().Eats {
		w.carryOff(c, e, t, w.fleet)
		w.consume(wr, c, e, t) // stripped, home or not, and held empty: see replicator.go
		return
	}
	if t == e.Home {
		w.homeFalls(wr, c, e)
		return
	}
	i := wr.side(c.ID)
	colony := e.Species.ID
	first := wr.Taken[i]+wr.Glassed[i] == 0
	converted := false
	w.carryOff(c, e, t, w.fleet) // what is mobile leaves with the taker before the world is lost
	switch {
	case c.miracle("unmaking"):
		w.setBio(t, BioNone)
		w.loseSystem(e, t, "unmade", reason{})
		wr.Glassed[i]++
		w.fact(FBurned, c, e, t).with(P{"way": "unmade", "species": colony, "told": true})
	case c.Own >= 0:
		w.loseSystem(e, t, "host", reason{})
		w.setOwner(t, c.ID)
		c.Systems = append(c.Systems, t)
		wr.Taken[i]++
		converted = true
		w.told(FTaken, c, e, t).with(P{"way": "host_war", "told": true})
	case c.Has("swarming"):
		w.loseSystem(e, t, "overrun", reason{})
		w.setOwner(t, c.ID)
		c.Systems = append(c.Systems, t)
		wr.Taken[i]++
		converted = true
		w.told(FTaken, c, e, t).with(P{"way": "overrun", "told": first})
	case c.Claim[t]:
		w.loseSystem(e, t, "reclaimed", reason{})
		w.setOwner(t, c.ID)
		c.Systems = append(c.Systems, t)
		wr.Taken[i]++
		w.reclaimed(c, e, t) // restored to the realm, in the taker's own telling: nobody glasses what was theirs
	case c.hates(e) || !w.canLive(c, t):
		w.setBio(t, BioSimple)
		w.loseSystem(e, t, "glassed", reason{})
		wr.Glassed[i]++
		f := w.told(FBurned, c, e, t).with(P{"way": "glassed", "species": colony})
		f.P["told"] = first || w.R.Float64() < 0.3
	default:
		w.loseSystem(e, t, "conquered", reason{})
		w.setOwner(t, c.ID)
		c.Systems = append(c.Systems, t)
		wr.Taken[i]++
		f := w.told(FTaken, c, e, t).with(P{"way": "war", "empty_sky": w.emptySky, "told": false})
		switch {
		case w.emptySky && (first || w.R.Float64() < 0.5):
			f.P["told"] = true
		case first || w.R.Float64() < 0.3:
			f.P["told"] = true
		}
	}
	_ = converted
	if w.Owner[t] == c.ID {
		w.takeOver(c, t) // the works there come back to use if the taker knows the art
		w.wakeReservoir(c, t)
	}
	c.Peak = max(c.Peak, len(c.Systems))
	wr.Lost[1-i]++
	c.Tally.Taken++
	e.Tally.Lost++
	wr.Will[i] += 0.3
	switch e.posture() {
	case "conqueror", "unyielding":
		wr.Will[1-i] += 0.2
	default:
		wr.Will[1-i] -= 0.3
	}
	if wr.Named < 0 {
		best, bn := t, 0
		for _, s := range sortedInts(boolKeys(wr.Contested)) {
			if wr.Contested[s] > bn {
				best, bn = s, wr.Contested[s]
			}
		}
		wr.Named = best
	}
	if e.Active() && (wr.Lost[1-i] == 1 || wr.Lost[1-i]%3 == 0) {
		w.face(e, "hold", 0)
	}
	if !e.Active() {
		w.endWar(wr, "destroyed")
	}
}

func boolKeys(m map[int]int) map[int]bool {
	out := map[int]bool{}
	for k := range m {
		out[k] = true
	}
	return out
}

// homeFalls is the last defence of a home broken.
func (w *World) homeFalls(wr *War, c, e *Civ) {
	if e.nomad() && e.Reach >= 1 && !e.Aloft {
		w.takeSky(e, because("lose_home_to").At(e.Home).By(c))
		w.endWar(wr, "peace")
		return
	}
	i := wr.side(c.ID)
	wr.Lost[1-i]++
	wr.Taken[i]++
	c.Tally.Taken++
	e.Tally.Lost++
	switch {
	case c.miracle("unmaking") && !c.Has("pacifist") && (e.Has("unyielding") || w.R.Float64() < 0.5):
		w.setBio(e.Home, BioNone)
		w.fact(FHomeBroken, c, e, e.Home).with(P{"way": "unmade"})
		w.endCiv(e, Extinct, because("unmade_by").By(c))
		w.endWar(wr, "extinction")
	case c.hates(e):
		w.setBio(e.Home, BioNone)
		w.fact(FScoured, c, e, e.Home).with(P{"way": "hated"})
		if e.Reach >= 1 && len(e.Systems) == 1 && w.R.Float64() < 0.5 {
			w.loseSystem(e, e.Home, "scoured", because("scoured_by").At(e.Home).By(c))
		} else {
			w.endCiv(e, Extinct, because("scoured_by").At(e.Home).By(c))
		}
		w.endWar(wr, "extinction")
	case e.Has("unyielding"):
		w.setBio(e.Home, BioNone)
		w.fact(FHomeBroken, c, e, e.Home).with(P{"way": "shattered"})
		w.endCiv(e, Extinct, because("annihilated").By(c))
		w.endWar(wr, "extinction")
	case !e.treats():
		w.setBio(e.Home, BioSimple)
		w.fact(FScoured, c, e, e.Home).with(P{"way": "eater"})
		w.endCiv(e, Extinct, because("burned_out").At(e.Home).By(c))
		w.endWar(wr, "extinction")
	case e.Has("swarming"):
		w.setBio(e.Home, BioSimple)
		w.fact(FScoured, c, e, e.Home).with(P{"way": "nest"})
		w.endCiv(e, Extinct, because("burned_nests").By(c))
		w.endWar(wr, "extinction")
	case !e.Species.Profile().Can(species.Reseats):
		w.setBio(e.Home, BioSimple)
		w.fact(FHomeBroken, c, e, e.Home).with(P{"way": "world"})
		w.endCiv(e, Extinct, because("home_taken").At(e.Home).By(c))
		w.endWar(wr, "extinction")
	case e.Has("onequeen"):
		w.fact(FHomeBroken, c, e, e.Home).with(P{"way": "queen"})
		w.endCiv(e, Extinct, because("queen_taken").At(e.Home).By(c))
		w.endWar(wr, "extinction")
	case c.Has("pacifist"):
		w.fact(FYield, c, e, e.Home).with(P{"way": "spared"}).with(w.warSpanP(wr))
		w.endWar(wr, "peace")
	case c.Own >= 0:
		w.event(KDefencesBroken, c, e, e.Home, P{"outcome": "ridden"})
		w.ride(c, e)
		w.endWar(wr, "enslaved")
	case e.Has("submissive") || c.Dials.Greed < 0.3:
		w.event(KDefencesBroken, c, e, e.Home, P{"outcome": "vassal"})
		w.vassal(c, e)
		w.endWar(wr, "vassal")
	default:
		w.event(KDefencesBroken, c, e, e.Home, P{"outcome": "enslaved"})
		w.enslave(c, e)
		w.endWar(wr, "enslaved")
	}
}

// ride is a parasite taking a people as hosts: slaves, and half their
// tree, and the parasite's plague in them if it was not already.
func (w *World) ride(p, h *Civ) {
	w.enslave(p, h)
	if !p.Ridden[h.ID] {
		p.Ridden[h.ID] = true
		p.Tally.Ridden++
	}
	if p.Own >= 0 && h.Infections[p.Own] == nil && !h.Immune[p.Own] {
		w.infect(h, w.Plagues[p.Own], p, "ridden")
	}
	for _, eid := range sortedInts(h.Met) {
		if eid != p.ID && !p.Met[eid] {
			p.Met[eid] = true // the rider sees with the host's eyes
			w.noticed(p, w.Civs[eid], -1)
		}
	}
	p.Hosts = len(w.hostsOf(p))
	for _, k := range knownOf(h) {
		if !p.Known[k] && w.R.Float64() < 0.5 {
			if mode, _ := w.aptitude(p, tech.Get(k)); mode == aptDear {
				w.know(p, k)
			}
		}
	}
	w.recompute(p)
	w.event(KRiddenWar, h, p, -1, P{})
}

// judge decides whether a war goes on. While a fleet is in flight nothing
// is decided: the war it was sent to has not begun. Peace with terms
// needs each side to understand the other; a war between two peoples
// neither of whom fathoms the other ends only when a side falls or both
// wills run out; a war with a side that makes no terms (a thing that
// eats) ends only by exhaustion, both wills gone and nothing signed.
func (w *World) judge(wr *War) {
	a, b := w.Civs[wr.Sides[0]], w.Civs[wr.Sides[1]]
	if w.fleetInFlight(a, b) || w.fleetInFlight(b, a) {
		return
	}
	switch {
	case wr.Gap != nil && wr.Will[0] <= 0:
		w.event(KWarEnded, a, b, -1, P{"way": "hunt"}).with(w.warSpanP(wr))
		w.endWar(wr, "exhaustion")
	case wr.Gap != nil:
		// a hunt has nobody to treat with: it runs until the hunter's will is spent or the region is empty; the hunted side's will is nothing to it, since no offer of its can be received; see gap.go
	case w.noTerms(wr) && wr.Will[0] <= 0 && wr.Will[1] <= 0:
		w.event(KWarEnded, a, b, -1, P{"way": "spent"}).with(w.warSpanP(wr))
		w.endWar(wr, "exhaustion")
	case w.noTerms(wr):
		// no offer of terms is heard: the war goes on until the other side tires too
	case wr.Will[0] <= 0 && wr.Will[1] <= 0 && w.canTreat(wr):
		w.peace(wr, because("tired_both"))
	case wr.Will[0] <= 0 && wr.Will[1] <= 0:
		w.exhausted(wr)
	case wr.Will[0] <= 0:
		w.yield(wr, 0)
	case wr.Will[1] <= 0:
		w.yield(wr, 1)
	}
}

// canTreat says whether a war can end with terms: each side understands
// the other, or the war is an infection, which is no negotiation and
// needs no understanding.
func (w *World) canTreat(wr *War) bool {
	return w.mutual(w.Civs[wr.Sides[0]], w.Civs[wr.Sides[1]])
}

// exhausted is both wills gone between peoples that do not understand
// each other: the fighting stops, and nothing is signed. A side that
// fathoms the other sues it for a truce; otherwise the war is simply over.
func (w *World) exhausted(wr *War) {
	a, b := w.Civs[wr.Sides[0]], w.Civs[wr.Sides[1]]
	switch {
	case a.Fathomed[b.ID]:
		w.truce(wr, a, b)
	case b.Fathomed[a.ID]:
		w.truce(wr, b, a)
	default:
		a.Tally.Misunderstood++
		b.Tally.Misunderstood++
		w.event(KWarEnded, a, b, -1, P{"way": "misunderstood"}).with(w.warSpanP(wr))
		w.endWar(wr, "exhaustion")
	}
}

// truce is a people that understands its enemy, and is not understood,
// suing for the one thing it knows how to ask for: no new war for a
// time. Nothing with terms.
func (w *World) truce(wr *War, l, v *Civ) {
	l.Tally.Misunderstood++
	v.Tally.Misunderstood++
	w.fact(FPeace, l, v, -1).with(P{"way": "truce", "why": "", "net": 0}).with(w.warSpanP(wr))
	w.endWar(wr, "truce")
}

// yield is one side's will gone while the other's holds: capitulation if
// the winner still has something to take, peace otherwise. A loser that
// does not understand the winner cannot yield, and one that is not
// understood can only sue for a truce.
func (w *World) yield(wr *War, li int) {
	l, v := w.Civs[wr.Sides[li]], w.Civs[wr.Sides[1-li]]
	switch {
	case l.Aloft:
		w.peace(wr, because("moving_on").By(l))
		return
	case l.nomad() && l.Reach >= 1 && !l.Aloft:
		w.takeSky(l, because("yield_to").By(v))
		w.peace(wr, because("gone_to_sky").By(l))
		return
	case v.Aloft:
		for _, t := range w.front(v, l) {
			if wr.Over || !l.Active() {
				break
			}
			w.strip(wr, v, l, t)
		}
		if !wr.Over {
			w.peace(wr, because("took_and_left").By(v))
		}
		return
	case w.canTreat(wr):
	case !l.Fathomed[v.ID]:
		return // the offer would mean nothing: the war goes on until the other side tires too
	default:
		w.truce(wr, l, v)
		return
	}
	if len(w.front(v, l)) == 0 && len(w.fleetFront(v, l)) == 0 {
		w.peace(wr, because("tired_one").By(l))
		return
	}
	w.capitulate(wr, l, v)
}

// capitulate cedes the front, and the home too if it lies in reach.
func (w *World) capitulate(wr *War, l, v *Civ) {
	if w.tribute(wr, l, v) {
		return
	}
	vi := wr.side(v.ID)
	ceded := 0
	// the winner's gains and a little more; the rest of the front is safe by the truce
	limit := wr.Taken[vi] + wr.Glassed[vi] + 2
	for _, t := range append(w.front(v, l), w.fleetFront(v, l)...) {
		if t == l.Home || ceded >= limit || !contains(l.Systems, t) {
			continue
		}
		w.loseSystem(l, t, "ceded", reason{})
		if !l.Active() {
			break
		}
		w.setOwner(t, v.ID)
		v.Systems = append(v.Systems, t)
		ceded++
	}
	v.Peak = max(v.Peak, len(v.Systems))
	wr.Taken[vi] += ceded
	l.Tally.Capitulated = true
	if !l.Active() {
		w.endWar(wr, "destroyed")
		return
	}
	if !w.canStrikeHome(v, l) {
		w.factN(FYield, v, l, -1, ceded).with(P{"way": "ceded"}).with(w.warSpanP(wr))
		w.endWar(wr, "capitulation")
		return
	}
	switch {
	case v.hates(l):
		w.fact(FScoured, v, l, l.Home).with(P{"way": "yielded"})
		w.setBio(l.Home, BioNone)
		w.endCiv(l, Extinct, because("scoured_by").At(l.Home).By(v))
		w.endWar(wr, "extinction")
	case v.Has("pacifist"):
		w.factN(FYield, v, l, -1, ceded).with(P{"way": "took"}).with(w.warSpanP(wr))
		w.endWar(wr, "capitulation")
	case v.Own >= 0:
		w.event(KYielded, l, v, -1, P{"outcome": "ridden"}).with(w.warSpanP(wr))
		w.ride(v, l)
		w.endWar(wr, "enslaved")
	case l.Has("submissive") || v.Dials.Greed < 0.3 || l.Has("swarming") || !l.Species.Profile().Can(species.Reseats):
		w.event(KYielded, l, v, -1, P{"outcome": "vassal"}).with(w.warSpanP(wr))
		w.vassal(v, l)
		w.endWar(wr, "vassal")
	default:
		w.event(KYielded, l, v, -1, P{"outcome": "enslaved"}).with(w.warSpanP(wr))
		w.enslave(v, l)
		w.endWar(wr, "enslaved")
	}
}

// peace ends a war with nobody yielding.
func (w *World) peace(wr *War, why reason) {
	a, b := w.Civs[wr.Sides[0]], w.Civs[wr.Sides[1]]
	net := (wr.Taken[0] + wr.Glassed[0]) - (wr.Taken[1] + wr.Glassed[1])
	w.fact(FPeace, a, b, -1).with(P{"way": "peace", "net": net}).with(why.params("why")).with(w.warSpanP(wr))
	w.endWar(wr, "peace")
}

// warSpanP is how long a war ran and what it cost, as an event's
// parameters: the years, the worlds gone, and whether it was over within
// a tick.
func (w *World) warSpanP(wr *War) P {
	return P{"war": wr.ID, "years": w.Now - wr.Began, "gone": wr.Taken[0] + wr.Taken[1] + wr.Glassed[0] + wr.Glassed[1], "short": w.Now-wr.Began < Year(w.dt*1000)}
}

// warSpan says how long a war ran and what it cost, from those.
func warSpan(e *Event) string {
	if e.P["short"].(bool) {
		return sprintf("a short war; %s changed hands or burned", worlds(e.P["gone"].(int)))
	}
	return sprintf("%s of war; %s changed hands or burned", span(e.P["years"].(Year)), worlds(e.P["gone"].(int)))
}

func span(y Year) string {
	switch {
	case y < 1000:
		return "a few centuries"
	case y < 2000:
		return "a thousand years"
	default:
		return sprintf("%d thousand years", y/1000)
	}
}

func worlds(n int) string {
	switch n {
	case 0:
		return "no worlds"
	case 1:
		return "one world"
	}
	return sprintf("%d worlds", n)
}

// endWar closes a war and writes the record into both peoples.
func (w *World) endWar(wr *War, result string) {
	if wr.Over {
		return
	}
	wr.Over = true
	wr.Ended = w.Now
	wr.Result = result
	w.warOver(wr, result)
	a, b := w.Civs[wr.Sides[0]], w.Civs[wr.Sides[1]]
	if wr.Named < 0 && wr.Taken[0]+wr.Taken[1]+wr.Glassed[0]+wr.Glassed[1] > 0 {
		// named for the home that fell, or the loser's home
		switch {
		case wr.Lost[1] > wr.Lost[0]:
			wr.Named = b.Home
		case wr.Lost[0] > wr.Lost[1]:
			wr.Named = a.Home
		}
	}
	delete(a.Wars, b.ID)
	delete(b.Wars, a.ID)
	truce := Year(5000 + w.R.IntN(10000))
	a.Truce[b.ID] = w.Now + truce
	b.Truce[a.ID] = w.Now + truce
	a.resent(b.ID, 0.3+0.3*float64(wr.Lost[0]))
	b.resent(a.ID, 0.3+0.3*float64(wr.Lost[1]))
	b.resent(a.ID, 0.5) // being struck first is the deeper wrong
	w.renew(a, 0.1)     // a war fought to its end is something new
	w.renew(b, 0.1)
	w.spent(wr) // the long sleep, for a side with nothing left it wants
	w.warEnded(wr)
}

// hasFleetAgainst says whether c has a campaign fleet at or bound for e,
// or one gathering.
func (w *World) hasFleetAgainst(c, e *Civ) bool {
	return w.fleetInFlight(c, e) || w.fleetAtBase(c, e) || (c.Muster != nil && c.Muster.Target == e.ID)
}

func (w *World) fleetInFlight(c, e *Civ) bool {
	for _, x := range w.fleetsOf(c) {
		if x.Target == e.ID && x.Kind == Campaign && !x.Returning && x.Base < 0 {
			return true
		}
	}
	return false
}

// fleetFront is e's worlds within a hop of c's fleets at base.
func (w *World) fleetFront(c, e *Civ) []int {
	var out []int
	for _, x := range w.fleetsOf(c) {
		if x.Target != e.ID || x.Kind != Campaign || x.Returning || x.Base < 0 {
			continue
		}
		for _, s := range e.Systems {
			if w.G.Dist(x.Base, s) <= 20 && !contains(out, s) {
				out = append(out, s)
			}
		}
	}
	return out
}

// canStrikeHome says whether c can reach e's home, from a holding or a fleet.
func (w *World) canStrikeHome(c, e *Civ) bool {
	return w.inReach(c, e.Home) || contains(w.fleetFront(c, e), e.Home)
}

func (w *World) fleetAtBase(c, e *Civ) bool {
	for _, x := range w.fleetsOf(c) {
		if x.Target == e.ID && x.Kind == Campaign && !x.Returning && x.Base >= 0 {
			return true
		}
	}
	return false
}

// wartime is what a war costs a people each tick apart from the fighting.
func (w *World) wartime(c *Civ) {
	if len(c.Wars) == 0 {
		return
	}
	c.Morale -= 0.03 * w.dt * float64(len(c.Wars))
	c.Focus[tech.Weapons] = max(c.Focus[tech.Weapons], 1.5)
}
