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
	Cause     string
	Name      string
	Nth       int // the nth war between these two
	Will      [2]float64
	Taken     [2]int // worlds taken by each side
	Glassed   [2]int // worlds destroyed by each side
	Lost      [2]int // worlds lost by each side
	Result    string
	Pact      int // pact this war was joined under, -1
	Principal int // the ally whose war this is, -1
	Contested map[int]int
	Burn      bool         // a host burns what a parasite has converted
	Called    map[int]bool // allies already called to this war
	Hire      int          // the contract this war was declared for, or -1; see contract.go
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
func (w *World) declare(c, e *Civ, cause string) *War {
	if wr := w.warBetween(c.ID, e.ID); wr != nil {
		return wr
	}
	if !c.Active() || !e.Active() {
		return nil
	}
	c.Fought[e.ID]++
	e.Fought[c.ID]++
	wr := &War{ID: len(w.Wars), Sides: [2]int{c.ID, e.ID}, Began: w.Now, Cause: cause, Nth: c.Fought[e.ID], Pact: -1, Principal: -1, Hire: -1, Contested: map[int]int{}, Called: map[int]bool{},
		Slights: map[int]float64{}, Sent: map[int]float64{}, Slighted: map[int]float64{}, SlightTold: map[int]bool{}}
	wr.Will = [2]float64{w.initialWill(c, e, true), w.initialWill(e, c, false)}
	w.Wars = append(w.Wars, wr)
	c.Wars[e.ID], e.Wars[c.ID] = true, true
	w.cutTrade(c, e, "the war")
	w.cutTrade(e, c, "the war")
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
	w.factOf(FWar, c, e, -1, cause)
	w.slighted(c, e, wr)
	switch {
	case cause == "infection":
		// the infection line is already written
	case wr.Nth > 1:
		w.log("The %s go to war with the %s again, the %s time, over %s.", c.Name, e.Name, ordinal(wr.Nth), cause)
	default:
		w.log("The %s declare war on the %s, over %s.", c.Name, e.Name, cause)
	}
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
			w.endWar(wr, "the fall of a side")
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
	if t == e.Home {
		w.homeFalls(wr, c, e)
		return
	}
	i := wr.side(c.ID)
	colony := e.Species.Flavour().Colony
	first := wr.Taken[i]+wr.Glassed[i] == 0
	converted := false
	w.carryOff(c, e, t, w.fleet) // what is mobile leaves with the taker before the world is lost
	switch {
	case c.miracle("unmaking"):
		w.Bio[t] = BioNone
		w.loseSystem(e, t, "unmade world", "")
		wr.Glassed[i]++
		w.log("The %s unmake %s, a %s of the %s. There is nothing left to glass.", c.Name, w.star(t), colony, e.Name)
		w.fact(FBurned, c, e, t)
	case wr.Burn && e.Species.Sub == species.Parasite:
		w.Bio[t] = BioSimple
		w.loseSystem(e, t, "burned host-world", "")
		wr.Glassed[i]++
		wr.Will[1-i] -= 0.4
		w.log("The %s burn %s to be rid of what the %s put there.", c.Name, w.star(t), e.Name)
		w.fact(FBurned, c, e, t)
	case c.Species.Sub == species.Parasite:
		w.loseSystem(e, t, "host-world", "")
		w.Owner[t] = c.ID
		c.Systems = append(c.Systems, t)
		wr.Taken[i]++
		converted = true
		w.fact(FTaken, c, e, t)
		if !c.Ridden[e.ID] {
			c.Ridden[e.ID] = true
			c.Hosts = 1 + len(c.Ridden)
			w.log("The %s of %s are riders now. The %s wear them.", e.Name, w.star(t), c.Name)
		}
	case c.Has("swarming"):
		w.loseSystem(e, t, "overrun "+colony, "")
		w.Owner[t] = c.ID
		c.Systems = append(c.Systems, t)
		wr.Taken[i]++
		converted = true
		w.fact(FTaken, c, e, t)
		if first {
			w.log("The %s overrun %s. Where the %s were there is a nest.", c.Name, w.star(t), e.Name)
		}
	case c.hates(e) || !w.canLive(c, t):
		w.Bio[t] = BioSimple
		w.loseSystem(e, t, "glassed world", "")
		wr.Glassed[i]++
		w.fact(FBurned, c, e, t)
		if first || w.R.Float64() < 0.3 {
			w.log("The %s glass %s, a %s of the %s.", c.Name, w.star(t), colony, e.Name)
		}
	default:
		w.loseSystem(e, t, "conquered "+colony, "")
		w.Owner[t] = c.ID
		c.Systems = append(c.Systems, t)
		wr.Taken[i]++
		w.fact(FTaken, c, e, t)
		switch {
		case w.emptySky && (first || w.R.Float64() < 0.5):
			w.log("The %s take %s from the %s. There was nothing in its sky.", c.Name, w.star(t), e.Name)
		case first || w.R.Float64() < 0.3:
			w.log("The %s take %s from the %s.", c.Name, w.star(t), e.Name)
		}
	}
	_ = converted
	if w.Owner[t] == c.ID {
		w.takeOver(c, t) // the works there come back to use if the taker knows the art
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
	if wr.Name == "" {
		best, bn := t, 0
		for _, s := range sortedInts(boolKeys(wr.Contested)) {
			if wr.Contested[s] > bn {
				best, bn = s, wr.Contested[s]
			}
		}
		wr.Name = "the war of " + w.star(best)
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
		w.takeSky(e, "lose "+e.HomeName+" to the "+c.Name)
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
		w.Bio[e.Home] = BioNone
		w.log("The %s unmake %s, homeworld of the %s. It is not there any more.", c.Name, e.HomeName, e.Name)
		w.fact(FHomeBroken, c, e, e.Home)
		w.endCiv(e, Extinct, sprintf("were unmade by the %s", c.Name))
		w.endWar(wr, "extinction")
	case c.hates(e):
		w.Bio[e.Home] = BioNone
		w.log("The %s scour %s clean of the %s. They were too different to be let live.", c.Name, e.HomeName, e.Name)
		w.fact(FScoured, c, e, e.Home)
		if e.Reach >= 1 && len(e.Systems) == 1 && w.R.Float64() < 0.5 {
			w.loseSystem(e, e.Home, "scoured world", sprintf("were scoured from %s by the %s", e.HomeName, c.Name))
		} else {
			w.endCiv(e, Extinct, sprintf("were scoured from %s by the %s", e.HomeName, c.Name))
		}
		w.endWar(wr, "extinction")
	case e.Has("unyielding"):
		w.Bio[e.Home] = BioNone
		w.log("A relativistic strike from the %s shatters %s, homeworld of the %s. They never surrendered.", c.Name, e.HomeName, e.Name)
		w.fact(FHomeBroken, c, e, e.Home)
		w.endCiv(e, Extinct, sprintf("were annihilated in war with the %s", c.Name))
		w.endWar(wr, "extinction")
	case e.Has("swarming"):
		w.Bio[e.Home] = BioSimple
		w.log("The %s burn out the last nest of the %s. A swarm cannot be held; it can only be ended.", c.Name, e.Name)
		w.fact(FScoured, c, e, e.Home)
		w.endCiv(e, Extinct, sprintf("were burned out nest by nest by the %s", c.Name))
		w.endWar(wr, "extinction")
	case e.Species.Is(species.Planetary):
		w.Bio[e.Home] = BioSimple
		w.log("The %s take %s, and there is nothing to rule. The %s were the world, and the world is dead.", c.Name, e.HomeName, e.Name)
		w.fact(FHomeBroken, c, e, e.Home)
		w.endCiv(e, Extinct, sprintf("died when %s was taken by the %s", e.HomeName, c.Name))
		w.endWar(wr, "extinction")
	case c.Has("pacifist"):
		w.log("The %s defeat the %s and, having no use for a conquest, leave them be.", c.Name, e.Name)
		w.fact(FYield, c, e, e.Home)
		w.endWar(wr, "peace")
	case c.Species.Sub == species.Parasite:
		w.log("The %s break the last defences of %s.", c.Name, e.HomeName)
		w.ride(c, e)
		w.endWar(wr, "enslaved")
	case e.Has("submissive") || c.Dials.Greed < 0.3:
		w.log("The %s break the last defences of %s, and the %s bend the knee. They are vassals now.", c.Name, e.HomeName, e.Name)
		w.vassal(c, e)
		w.endWar(wr, "vassal")
	default:
		w.log("The %s break the last defences of %s.", c.Name, e.HomeName)
		w.enslave(c, e)
		w.endWar(wr, "enslaved")
	}
}

// ride is a parasite taking a people as hosts: slaves, and half their tree.
func (w *World) ride(p, h *Civ) {
	w.enslave(p, h)
	if !p.Ridden[h.ID] {
		p.Ridden[h.ID] = true
		p.Hosts = 1 + len(p.Ridden)
	}
	for _, k := range knownOf(h) {
		if !p.Known[k] && w.R.Float64() < 0.5 {
			if mode, _ := w.aptitude(p, tech.Get(k)); mode == aptDear {
				p.Known[k] = true
			}
		}
	}
	w.recompute(p)
	w.log("The %s are still there, and still themselves, mostly. They do what the %s want now, and what they knew, the %s know.", h.Name, p.Name, p.Name)
}

// judge decides whether a war goes on. While a fleet is in flight nothing
// is decided: the war it was sent to has not begun. Peace with terms
// needs each side to understand the other; a war between two peoples
// neither of whom fathoms the other ends only when a side falls or both
// wills run out.
func (w *World) judge(wr *War) {
	a, b := w.Civs[wr.Sides[0]], w.Civs[wr.Sides[1]]
	if w.fleetInFlight(a, b) || w.fleetInFlight(b, a) {
		return
	}
	switch {
	case wr.Will[0] <= 0 && wr.Will[1] <= 0 && w.canTreat(wr):
		w.peace(wr, "both sides tired of it")
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
	return wr.Cause == "infection" || w.mutual(w.Civs[wr.Sides[0]], w.Civs[wr.Sides[1]])
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
		w.log("The %s and the %s stop fighting, both sides tired of it, after %s. Neither ever understood what the other wanted, and nothing is signed.", a.Name, b.Name, w.warSpan(wr))
		w.endWar(wr, "exhaustion")
	}
}

// truce is a people that understands its enemy, and is not understood,
// suing for the one thing it knows how to ask for: no new war for a
// time. Nothing with terms.
func (w *World) truce(wr *War, l, v *Civ) {
	l.Tally.Misunderstood++
	v.Tally.Misunderstood++
	w.log("The %s sue the %s for a truce, which is all they know how to ask for, after %s. The fighting stops; nothing is settled.", l.Name, v.Name, w.warSpan(wr))
	w.factOf(FPeace, l, v, -1, "truce")
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
		w.peace(wr, sprintf("the %s moving on", l.Name))
		return
	case l.nomad() && l.Reach >= 1 && !l.Aloft:
		w.takeSky(l, "yield to the "+v.Name)
		w.peace(wr, sprintf("the %s gone to the sky", l.Name))
		return
	case v.Aloft:
		for _, t := range w.front(v, l) {
			if wr.Over || !l.Active() {
				break
			}
			w.strip(wr, v, l, t)
		}
		if !wr.Over {
			w.peace(wr, sprintf("the %s taking what they wanted and moving on", v.Name))
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
		w.peace(wr, sprintf("the %s tired of it, and the %s had nothing left to take", l.Name, v.Name))
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
		w.loseSystem(l, t, "ceded "+l.Species.Flavour().Colony, "")
		if !l.Active() {
			break
		}
		w.Owner[t] = v.ID
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
		w.log("The %s yield to the %s and cede %s, after %s.", l.Name, v.Name, worlds(ceded), w.warSpan(wr))
		w.factN(FYield, v, l, -1, ceded)
		w.endWar(wr, "capitulation")
		return
	}
	switch {
	case v.hates(l):
		w.log("The %s yield to the %s, who want no terms. %s is scoured clean of them; they were too different to be let live.", l.Name, v.Name, l.HomeName)
		w.fact(FScoured, v, l, l.Home)
		w.Bio[l.Home] = BioNone
		w.endCiv(l, Extinct, sprintf("were scoured from %s by the %s", l.HomeName, v.Name))
		w.endWar(wr, "extinction")
	case v.Has("pacifist"):
		w.log("The %s yield to the %s, who take %s and want nothing more, after %s.", l.Name, v.Name, worlds(ceded), w.warSpan(wr))
		w.factN(FYield, v, l, -1, ceded)
		w.endWar(wr, "capitulation")
	case v.Species.Sub == species.Parasite:
		w.log("The %s yield to the %s, after %s.", l.Name, v.Name, w.warSpan(wr))
		w.ride(v, l)
		w.endWar(wr, "enslaved")
	case l.Has("submissive") || v.Dials.Greed < 0.3 || l.Has("swarming") || l.Species.Is(species.Planetary):
		w.log("The %s yield to the %s and bend the knee, after %s. They are vassals now.", l.Name, v.Name, w.warSpan(wr))
		w.vassal(v, l)
		w.endWar(wr, "vassal")
	default:
		w.log("The %s yield to the %s, after %s.", l.Name, v.Name, w.warSpan(wr))
		w.enslave(v, l)
		w.endWar(wr, "enslaved")
	}
}

// peace ends a war with nobody yielding.
func (w *World) peace(wr *War, why string) {
	a, b := w.Civs[wr.Sides[0]], w.Civs[wr.Sides[1]]
	terms := "Neither side is sure who won."
	switch net := (wr.Taken[0] + wr.Glassed[0]) - (wr.Taken[1] + wr.Glassed[1]); {
	case net > 0:
		terms = sprintf("The %s keep what they took.", a.Name)
	case net < 0:
		terms = sprintf("The %s keep what they took.", b.Name)
	}
	w.log("The %s and the %s make peace, %s, after %s. %s", a.Name, b.Name, why, w.warSpan(wr), terms)
	w.fact(FPeace, a, b, -1)
	w.endWar(wr, "peace")
}

// warSpan says how long a war ran and what it cost.
func (w *World) warSpan(wr *War) string {
	worldsGone := wr.Taken[0] + wr.Taken[1] + wr.Glassed[0] + wr.Glassed[1]
	if w.Now-wr.Began < Year(w.dt*1000) {
		return sprintf("a short war; %s changed hands or burned", worlds(worldsGone))
	}
	return sprintf("%s of war; %s changed hands or burned", span(w.Now-wr.Began), worlds(worldsGone))
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
	a, b := w.Civs[wr.Sides[0]], w.Civs[wr.Sides[1]]
	if wr.Name == "" && wr.Taken[0]+wr.Taken[1]+wr.Glassed[0]+wr.Glassed[1] > 0 {
		// named for the home that fell, or the loser's home
		switch {
		case wr.Lost[1] > wr.Lost[0]:
			wr.Name = "the war of " + b.HomeName
		case wr.Lost[0] > wr.Lost[1]:
			wr.Name = "the war of " + a.HomeName
		}
	}
	delete(a.Wars, b.ID)
	delete(b.Wars, a.ID)
	truce := Year(5000 + w.R.IntN(10000))
	a.Truce[b.ID] = w.Now + truce
	b.Truce[a.ID] = w.Now + truce
	a.Grudge[b.ID] += 0.3 + 0.3*float64(wr.Lost[0])
	b.Grudge[a.ID] += 0.3 + 0.3*float64(wr.Lost[1])
	b.Grudge[a.ID] += 0.5 // being struck first is the deeper wrong
	w.warEnded(wr)
}

// hasFleetAgainst says whether c has a campaign fleet at or bound for e,
// or one gathering.
func (w *World) hasFleetAgainst(c, e *Civ) bool {
	return w.fleetInFlight(c, e) || w.fleetAtBase(c, e) || (c.Muster != nil && c.Muster.Target == e.ID)
}

func (w *World) fleetInFlight(c, e *Civ) bool {
	for _, x := range w.Expeditions {
		if !x.Over && x.Owner == c.ID && x.Target == e.ID && x.Kind == Campaign && !x.Returning && x.Base < 0 {
			return true
		}
	}
	return false
}

// fleetFront is e's worlds within a hop of c's fleets at base.
func (w *World) fleetFront(c, e *Civ) []int {
	var out []int
	for _, x := range w.Expeditions {
		if x.Over || x.Owner != c.ID || x.Target != e.ID || x.Kind != Campaign || x.Returning || x.Base < 0 {
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
	for _, x := range w.Expeditions {
		if !x.Over && x.Owner == c.ID && x.Target == e.ID && x.Kind == Campaign && !x.Returning && x.Base >= 0 {
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
