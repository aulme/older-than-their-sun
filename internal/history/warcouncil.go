package history

import (
	"math"

	"worldgen/internal/battle"
	"worldgen/internal/flow"
	"worldgen/internal/mind"
	"worldgen/internal/species"
)

// The war council: each side of each running war asks, on the council's
// cadence and whenever the war summons it, whether to press, hold or sue.
// A war has an aim, from its cause and the declarer's posture, which says
// what the declarer's winning is; the side declared on holds, and presses
// to win back what it lost. Terms are offered by the side that sues, as
// the other's aim wants them, and the other's council answers. Every war
// has the council, however it began: the heir's inherited war, the
// sleeper's, the ally's. The judgment is mind.WarCouncil, mind.AnswerTerms
// and mind.AnswerYoke; this file gathers and executes.

// aimOf is what a war is declared for: its cause's aim, total for the
// hating and raised to submission by a conqueror; an ally's is defence.
func (w *World) aimOf(c, e *Civ, cause string) string {
	aim := mind.AimWorld
	if d := tables.warCauses[cause]; d != nil && d.Aim != "" {
		aim = d.Aim
	}
	switch {
	case aim == mind.AimDefence:
	case c.hates(e):
		aim = mind.AimEnding
	case c.posture() == mind.Conqueror && mind.Limited(aim):
		aim = mind.AimSubmission
	}
	return aim
}

// openWar is a war declared with its first fleet: a conqueror whose fleet
// sails for a border world and not the home fights for that world, and a
// war for submission is only one where the home is the first target —
// but between old enemies, whose third war is for the whole wherever it
// begins.
func (w *World) openWar(c, e *Civ, cause reason, target int) *War {
	wr := w.declare(c, e, cause)
	if wr != nil && wr.Aim == mind.AimSubmission && target != e.Home && !wr.Escalated {
		wr.Aim = mind.AimWorld
	}
	return wr
}

// aim is a side's aim in a war: the declarer's own, the other's to hold.
func (wr *War) aim(i int) string {
	if i == 0 {
		return wr.Aim
	}
	return mind.AimHold
}

// aimMet says whether a side has what it went to war for, short of the
// enemy's end: a world taken for a border or tribute, for redress the
// worlds lost in the last war, and for the side declared on more taken
// than it lost.
func (wr *War) aimMet(i int) bool {
	took := wr.Taken[i] + wr.Glassed[i]
	switch wr.aim(i) {
	case mind.AimWorld, mind.AimTribute:
		return took >= 1
	case mind.AimRedress:
		return took >= max(1, wr.Asks) // what was lost taken back (step 11's batches: met at one world whatever was lost, the rivals' wars were over as they began; at two for all, a third of the wars went)
	case mind.AimHold:
		return took > 0 && took >= wr.Lost[i]
	}
	return false
}

// aimTarget is the next world the aim names: the home for a total aim
// when a fleet can reach it; for the side declared on, the nearest of the
// worlds it lost in the war that the enemy still holds; otherwise the
// enemy's world nearest.
func (w *World) aimTarget(c, e *Civ, aim string, wr *War) int {
	_, near := w.nearestEnemy(c, e)
	switch {
	case (aim == mind.AimSubmission || aim == mind.AimEnding) && w.inReach(c, e.Home):
		return e.Home
	case aim == mind.AimHold && wr != nil:
		best, bd := -1, 0.0
		for _, s := range wr.Seized[1-wr.side(c.ID)] {
			if !w.holds(e, s) {
				continue
			}
			if _, d := w.nearest(c, s); best < 0 || d < bd {
				best, bd = s, d
			}
		}
		if best >= 0 {
			return best
		}
	}
	return near
}

// warCouncils is a people's step: the appetite wanes, and each of its
// wars sits a council if it is summoned or its turn has come.
func (w *World) warCouncils(c *Civ) {
	t := &w.Cfg.Tuning.War
	if c.Appetite > 0 {
		c.Appetite *= math.Pow(0.5, w.dt*1000/t.AppetiteHalf)
		if c.Appetite < 0.01 {
			c.Appetite = 0
		}
	}
	if !c.Active() || !c.sits() || c.Aloft || c.Asleep || len(c.Wars) == 0 {
		return
	}
	for _, eid := range sortedInts(c.Wars) {
		wr := w.warBetween(c.ID, eid)
		if wr == nil || wr.Gap != nil {
			continue // a hunt has nobody to treat with or appraise; see gap.go
		}
		i := wr.side(c.ID)
		if !wr.Summon[i] && !w.chance(t.Cadence) {
			continue
		}
		wr.Summon[i] = false
		w.warCouncil(wr, c, w.Civs[eid], i)
		if !c.Active() {
			return
		}
	}
}

// warInput is what one side's council reads of a war: the balance as
// believed against the next target the aim names, and the rest.
func (w *World) warInput(wr *War, c, e *Civ, i int) (mind.WarInput, int) {
	t := w.Cfg.Tuning
	aim := wr.aim(i)
	target := w.aimTarget(c, e, aim, wr)
	ap := w.appraise(c, e, target)
	if target == e.Home {
		// a total aim presses where it can: the home if the odds are there, else the nearest world
		if _, near := w.nearestEnemy(c, e); near != target && w.holds(e, near) {
			if ap2 := w.appraise(c, e, near); ap2.Acted > ap.Acted {
				target, ap = near, ap2
			}
		}
	}
	_, _, far := w.bar(c, e)
	since := max(wr.Began, wr.Fought)
	if pw := w.principalWar(wr); pw != nil {
		since = max(since, pw.Fought) // an ally is not idle while its principal fights
	}
	in := mind.WarInput{
		Aim: aim, Declarer: i == 0, Posture: c.posture(), Hates: c.hates(e),
		Acted: ap.Acted, Reach: c.launches() && w.holds(e, target) && (len(ap.Front) > 0 || far || ap.Lag < t.Campaign.MaxLag), Out: w.campaignOut(c, e),
		Met: wr.aimMet(i), Fought: wr.Battles > 0, Idle: float64(w.Now-since) / (w.dt * 1000), Will: wr.Will[i], Fear: c.Dials.Fear,
		Lost: wr.Lost[i], Taken: wr.Taken[i] + wr.Glassed[i], Appetite: mind.Appetite(c.Appetite, t), Wary: c.Wary[e.ID],
		Treats: w.canTreat(wr) && !w.noTerms(wr), Bound: w.alliance(wr, c),
	}
	if in.Reach && !in.Out {
		k := w.sizeAt(c, e, target)
		in.Ready = k.Send && (c.Muster == nil || c.Muster.Target == e.ID || w.guardWith(c, target, k.Share) != nil)
	}
	return in, target
}

// warCouncil is one side's council on one war.
func (w *World) warCouncil(wr *War, c, e *Civ, i int) {
	if !e.Active() {
		return
	}
	t := w.Cfg.Tuning
	in, target := w.warInput(wr, c, e, i)
	v := mind.WarCouncil(in, t)
	if v.Choice == mind.Press {
		v.Reason += "; " + w.press(wr, c, e, in.Aim, target)
	}
	if in.Out && w.tracing() {
		v.Reason += "; " + w.outWhy(c, e)
	}
	if v.Choice == mind.Hold && in.Aim == mind.AimDefence && wr.Principal >= 0 && (v.Key == "ships" || v.Key == "odds" || v.Key == "reach") {
		if pr := w.Civs[wr.Principal]; pr.Active() {
			v.Reason += "; " + w.standWith(c, pr, e)
		}
	}
	w.explain(c, sprintf("at war with the %s (war %d, for %s)", e.Tok(), wr.ID, in.Aim), v)
	wr.Verdict[i], wr.Why[i] = v.Choice, v.Key
	switch v.Choice {
	case mind.Sue:
		if wr.Offered[i] == 0 || float64(w.Now-wr.Offered[i]) >= t.War.OfferEvery {
			w.sue(wr, i)
		}
	}
}

// alliance is what the alliance is worth to c, when c is an ally in a war
// joined by pact and its principal fights on: the pact's age and size,
// the principal's renown and faith, the enemy's menace to c itself. A
// separate peace would throw it away. Nothing for anyone else.
func (w *World) alliance(wr *War, c *Civ) float64 {
	if wr.Principal < 0 || wr.Sides[0] != c.ID || w.principalWar(wr) == nil {
		return 0
	}
	pr, e := w.Civs[wr.Principal], w.Civs[wr.Sides[1]]
	p := w.pactWith(c, pr)
	if p == nil {
		return 0
	}
	mil, _ := w.believe(c, e)
	menace := mind.Threatens(mind.ThreatInput{Believed: mil, Mil: c.Mil, Hostile: e.hostile(), Hates: e.hates(c), AtWar: true, Rules: e.Ruled > 0, Grudge: c.Grudge[e.ID] > 0}, w.Cfg.Tuning)
	return mind.Bound(mind.BoundInput{Age: float64(w.Now - p.Formed), Members: len(p.Members), Menace: menace, Renown: w.renown(pr), Betrayed: w.betrayed(c, pr)}, w.Cfg.Tuning)
}

// standWith is an ally at war that cannot carry the war to the enemy
// sending ships to stand at its principal's home instead, as the relief
// a call brings: it fights there when the enemy comes, and that battle is
// its war's too (battle.go). Once, while its relief stands; never
// leaving its own home unsafe.
func (w *World) standWith(m, v, a *Civ) string {
	for _, x := range w.fleetsOf(m) {
		if x.Kind == Relief && x.Target == v.ID && !x.Returning {
			return "its ships stand with the " + v.Tok() + " already"
		}
	}
	milA, _ := w.believe(m, a)
	k := mind.AnswerCall(mind.CallInput{Ships: w.standing(m), Total: w.ships(m), Q: w.quality(m), Victim: battle.Strength(w.standing(v), w.quality(v)), Believed: milA + a.warBonus() + mind.ShipLevels(w.believeShips(m, a)), Confederate: m.posture() == mind.Confederate, Dials: m.Dials}, w.Cfg.Tuning)
	if !k.Safe || k.Share > w.standing(m) {
		return "no ships to spare to stand with the " + v.Tok()
	}
	w.launch(m, Relief, v, v.Home, k.Share)
	return sprintf("%d ships go to stand with the %s", k.Share, v.Tok())
}

// campaignOut says whether c has a campaign against e that is doing
// something: on its way, at a base and kept, or gathering. A fleet laid
// up in the enemy's sky, unfed and rotting, is nothing the council can
// wait on.
func (w *World) campaignOut(c, e *Civ) bool {
	for _, x := range w.fleetsOf(c) {
		if x.Target == e.ID && x.Kind == Campaign && !x.Returning && !x.LaidUp {
			return true
		}
	}
	return c.Muster != nil && c.Muster.Target == e.ID
}

// outWhy says where a people's campaign against another is, for -ai.
func (w *World) outWhy(c, e *Civ) string {
	for _, x := range w.fleetsOf(c) {
		if x.Target != e.ID || x.Kind != Campaign || x.Returning {
			continue
		}
		if x.Base < 0 {
			if x.Arrive <= w.Now {
				return "arriving"
			}
			return sprintf("in flight, %d years to go", x.Arrive-w.Now)
		}
		return sprintf("at a base, laid up %v, enemy's %v, own %v", x.LaidUp, w.holds(e, x.Base), w.Owner[x.Base] == c.ID)
	}
	if c.Muster != nil && c.Muster.Target == e.ID {
		return sprintf("mustering for %d years", w.Now-c.Muster.Since)
	}
	return "?"
}

// press sends a campaign at the next target the aim names: the home for
// a total aim if a fleet can take it, else the nearest.
func (w *World) press(wr *War, c, e *Civ, aim string, target int) string {
	targets := []int{target}
	if target == e.Home {
		if _, near := w.nearestEnemy(c, e); near != target {
			targets = append(targets, near)
		}
	}
	for _, s := range targets {
		if w.holds(e, s) && !w.protected(c, e, s) && w.sizeCampaign(c, e, because(wr.Cause), s) {
			if c.Muster != nil && c.Muster.Target == e.ID && c.Muster.Since == w.Now {
				return "a muster is called"
			}
			return "a fleet sails"
		}
	}
	k := w.sizeAt(c, e, targets[len(targets)-1])
	switch {
	case !k.LagOK:
		return "too far to send one"
	case k.Cap < k.Need:
		return sprintf("too few ships: %d wanted, %d could sail", k.Need, k.Cap)
	case c.Muster != nil:
		return "the guards are gathering for another war"
	}
	return "no fleet goes"
}

// Terms.

// offer is what a side that sues puts on the table: the kind, and the
// worlds or the object it gives.
type offer struct {
	kind   string // lines, worlds, tribute, artifact, vassal
	worlds []int
	source *Source
}

// offerFor is what l offers v: peace on the lines once its own aim is
// met; otherwise the first thing it can give of what v's aim wants.
func (w *World) offerFor(wr *War, li int) offer {
	l, v := w.Civs[wr.Sides[li]], w.Civs[wr.Sides[1-li]]
	if wr.aimMet(li) {
		return offer{kind: "lines"}
	}
	if l.Yields[v.ID] >= w.Cfg.Tuning.War.Yields && w.takesClient(v, l) && w.bends(l) {
		return offer{kind: "vassal"} // bought off often enough: it offers itself before its worlds (seed 15 of step 11's batches: a realm bought the same neighbour off twenty times, a world or two at a time)
	}
	var wants []string
	switch wr.aim(1 - li) {
	case mind.AimWorld:
		wants = []string{"worlds", "tribute", "artifact"}
	case mind.AimTribute:
		wants = []string{"tribute", "artifact", "worlds"}
	case mind.AimRedress:
		wants = []string{"worlds", "artifact", "tribute"}
	case mind.AimSubmission:
		wants = []string{"vassal", "worlds", "tribute"}
	}
	for _, k := range wants {
		switch k {
		case "worlds":
			if v.Aloft {
				continue // a horde holds no worlds: it takes them by stripping, not by treaty
			}
			n := 1
			switch wr.aim(1 - li) {
			case mind.AimRedress:
				n = max(1, wr.Asks)
			case mind.AimWorld:
			default:
				n = 2
			}
			if ws := w.cedable(l, v, n); len(ws) > 0 {
				return offer{kind: k, worlds: ws}
			}
		case "tribute":
			if _, amt := w.tributeOf(l, v); amt > 0 {
				return offer{kind: k}
			}
		case "artifact":
			if s := w.artifactFor(l, v); s != nil {
				return offer{kind: k, source: s}
			}
		case "vassal":
			if w.bends(l) {
				return offer{kind: k}
			}
		}
	}
	return offer{kind: "lines"}
}

// bends says whether a people would offer itself as a vassal rather than
// fight on: not the unyielding, not a conqueror, not what cannot be held.
func (w *World) bends(l *Civ) bool {
	switch l.posture() {
	case mind.Unyielding, mind.Conqueror:
		return false
	}
	return l.Free() && !l.Aloft && l.Own < 0 && l.Species.Profile().Can(species.Reseats)
}

// cedable is up to n of l's worlds it could hand to v: the front first,
// then the nearest to v; never the home.
func (w *World) cedable(l, v *Civ, n int) []int {
	var out []int
	for _, s := range append(w.front(v, l), w.fleetFront(v, l)...) {
		if len(out) < n && s != l.Home && contains(l.Systems, s) && !contains(out, s) {
			out = append(out, s)
		}
	}
	for len(out) < n {
		best, bd := -1, 0.0
		for _, s := range l.Systems {
			if s == l.Home || contains(out, s) {
				continue
			}
			if _, d := w.nearest(v, s); best < 0 || d < bd {
				best, bd = s, d
			}
		}
		if best < 0 {
			break
		}
		out = append(out, best)
	}
	return out
}

// artifactFor is a mobile rarity l holds at home or at a world and v
// lacks, the first by id.
func (w *World) artifactFor(l, v *Civ) *Source {
	for _, id := range w.mobile {
		if s := w.Sources[id]; s.Holder == l.ID && s.Carried < 0 && !v.has(s.Key) {
			return s
		}
	}
	return nil
}

// worth is how much of a side's aim an offer gives it, from 0 to 1,
// counting what it has taken already.
func worth(wr *War, j int, o offer) float64 {
	took := float64(wr.Taken[j] + wr.Glassed[j])
	n := float64(len(o.worlds))
	var x float64
	switch wr.aim(j) {
	case mind.AimWorld:
		x = map[string]float64{"worlds": 1, "tribute": 0.6, "artifact": 0.6, "vassal": 1, "lines": min(1, took)}[o.kind]
	case mind.AimTribute:
		x = map[string]float64{"worlds": 0.9, "tribute": 1, "artifact": 0.9, "vassal": 1, "lines": min(1, 0.8*took)}[o.kind]
	case mind.AimRedress:
		asks := float64(max(1, wr.Asks))
		x = map[string]float64{"worlds": (n + took) / asks, "tribute": 0.6, "artifact": 0.6, "vassal": 1, "lines": took / asks}[o.kind]
	case mind.AimSubmission:
		x = map[string]float64{"worlds": 0.25 * (n + took), "tribute": 0.3, "artifact": 0.3, "vassal": 1, "lines": 0.2 * took}[o.kind]
	case mind.AimDefence:
		x = 0.5
	case mind.AimHold:
		x = 1
		if o.kind == "lines" {
			x = max(0, 1-0.34*float64(wr.Lost[j])) + 0.2*took
		}
	}
	return min(1, x)
}

// sue is side i offering terms, and the other side's council answering.
func (w *World) sue(wr *War, i int) {
	l, v := w.Civs[wr.Sides[i]], w.Civs[wr.Sides[1-i]]
	wr.Offered[i] = w.Now
	o := w.offerFor(wr, i)
	j := 1 - i
	in, _ := w.warInput(wr, v, l, j)
	bar, ok := mind.PressBar(in, w.Cfg.Tuning)
	a := mind.AnswerTerms(mind.TermsInput{Worth: worth(wr, j, o), Acted: in.Acted, Will: wr.Will[j], Posture: v.posture(), Hates: v.hates(l), Greed: v.Dials.Greed,
		Presses: in.Out || (ok && in.Reach && in.Ready && in.Acted >= bar), Bound: in.Bound}, w.Cfg.Tuning)
	w.explain(v, "offered "+o.kind+" by the "+l.Tok(), a)
	if !a.Accept {
		w.event(KTermsRefused, l, v, -1, P{"terms": o.kind, "war": wr.ID})
		return
	}
	w.settleTerms(wr, l, v, o, worth(wr, j, o))
}

// settleTerms is terms accepted: what was offered changes hands, the war
// ends on them, and the truce is longer for each world given up. A side
// that took terms bought off by the other won the war, and its quarrel is
// ended; one that took peace on the lines from a side whose aim was met
// lost it, and what the terms gave it is paid against its grudge.
func (w *World) settleTerms(wr *War, l, v *Civ, o offer, paid float64) {
	g := v.Grudge[l.ID]
	li, vi := wr.side(l.ID), wr.side(v.ID)
	p := P{"terms": o.kind, "n": len(o.worlds)}
	switch o.kind {
	case "worlds":
		for _, s := range o.worlds {
			w.handOver(l, v, s)
		}
		wr.Taken[vi] += len(o.worlds)
		wr.Lost[li] += len(o.worlds)
	case "tribute":
		k := w.tributeTo(l, v)
		p["res"], p["for"] = k.Pay.Res, k.Until-w.Now
	case "artifact":
		w.transfer(o.source, l, v)
		o.source.Star = v.Home
		p["source"] = o.source.ID
	}
	net := (wr.Taken[li] + wr.Glassed[li]) - (wr.Taken[vi] + wr.Glassed[vi])
	p["net"] = net
	w.fact(FSettled, l, v, -1).with(p).with(w.warSpanP(wr))
	if o.kind == "vassal" {
		w.vassal(v, l, "surrender")
	}
	// who won: peace on the lines offered with the offerer's aim met is
	// its; anything else offered is the offerer buying the war off, and
	// the side that took it won it
	if o.kind == "lines" && wr.aimMet(li) {
		wr.Winner = l.ID
	} else {
		wr.Winner = v.ID
	}
	if wr.Winner == v.ID && o.kind != "lines" && o.kind != "vassal" {
		w.yielded(l, v) // a war bought off is a yield, and counts toward the knee bent
	}
	w.renounce(v, l) // a claim settled by treaty is a claim given up
	w.endWar(wr, "terms")
	if wr.Winner != v.ID {
		if g *= 1 - paid; g < w.Cfg.Tuning.Ossify.GrudgeFloor {
			delete(v.Grudge, l.ID)
		} else {
			v.Grudge[l.ID] = g
		}
	}
	if n := len(o.worlds); n > 0 {
		extra := Year(float64(n) * w.Cfg.Tuning.War.TruceWorld)
		l.Truce[v.ID] += extra
		v.Truce[l.ID] += extra
	}
}

// renounce is c giving up its claims on what e holds: settled by treaty,
// or won.
func (w *World) renounce(c, e *Civ) {
	for _, s := range e.Systems {
		delete(c.Claim, s)
	}
}

// tributeOf is the commodity and amount l would pay v in tribute: its
// best spare, carried as far as its transport goes; nothing when either
// cannot trade.
func (w *World) tributeOf(l, v *Civ) (flow.Kind, float64) {
	if !l.Species.Profile().Can(species.Trades) || !v.Species.Profile().Can(species.Trades) {
		return flow.O, 0
	}
	spare := w.spare(l)
	best, bestSpare := flow.O, 0.0
	for _, k := range flow.Kinds {
		if spare[k] > bestSpare {
			best, bestSpare = k, spare[k]
		}
	}
	if bestSpare <= 0.1 {
		return flow.O, 0
	}
	return best, bestSpare * w.transportCap(l)
}

// tributeTo makes the tribute l pays v as a running contract.
func (w *World) tributeTo(l, v *Civ) *Contract {
	t := w.Cfg.Tuning.Contract
	res, amt := w.tributeOf(l, v)
	k := w.newContract(l, v, Term{Kind: mind.TermPeace, Target: v.ID}, Term{Kind: mind.TermFlow, Res: res, Amount: amt}, v)
	k.Length = t.TributeLength
	k.State, k.Formed, k.Until, k.Tribute = Running, w.Now, w.Now+Year(t.TributeLength*1000), true
	l.Tally.Tributes++
	peaceStart(w, k, v, l, k.Ask)
	return k
}

// Vassalage without a war.

// yoke is a large realm, about to strike a small neighbour it believes
// would clearly lose, offering it vassalage first. Returns whether the
// question was put: accepted, the small people is a vassal; refused, the
// refusal is the war's cause.
func (w *World) yoke(c, e *Civ, ap Appraisal) (asked bool) {
	t := &w.Cfg.Tuning.War
	if !mind.OffersYoke(c.posture(), c.hates(e)) || ap.Acted < t.YokeOdds || !c.Free() || !e.Free() || e.Aloft || e.Own >= 0 || c.Own >= 0 {
		return false
	}
	if float64(len(c.Systems)) < max(t.YokeWorlds, t.YokeSize*float64(len(e.Systems))) {
		return false
	}
	if at, ok := c.Yoked[e.ID]; ok && float64(w.Now-at) < t.YokeAgain {
		return false
	}
	if c.Yoked == nil {
		c.Yoked = map[int]Year{}
	}
	c.Yoked[e.ID] = w.Now
	if !w.mutual(c, e) || !e.treats() {
		return false // an offer that cannot be understood is not an offer
	}
	back := w.appraise(e, c, -1) // the war it would fight, as it believes it
	a := mind.AnswerYoke(mind.YokeInput{Acted: back.Acted, Fear: e.Dials.Fear, Posture: e.posture(), Hates: e.hates(c)}, w.Cfg.Tuning)
	w.explain(e, "offered the yoke by the "+c.Tok(), a)
	if a.Accept {
		w.event(KYoke, c, e, e.Home, P{"answer": "accepted"})
		w.vassal(c, e, "offer")
		return true
	}
	w.event(KYoke, c, e, e.Home, P{"answer": "refused"})
	return true
}
