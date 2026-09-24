package history

import (
	"math"

	"worldgen/internal/mind"
)

// The council is where a people decides: whom to strike, whether to send a
// fleet or a scout first, whom to ask for a pact. It sits on a cadence and
// whenever an event summons it. Every option is scored through the
// appraisal; posture sets the bar; a little folly is wanted. The judgment
// is mind.Judge and mind.Council; this file gathers and executes.

// council sits, judges each enemy in turn, scouts or watches as each
// verdict says, then strikes the best of those that cleared their bar.
// Each judgment follows the scouts sent before it, since a scout takes a
// level from home.
func (w *World) council(c *Civ) {
	t := w.Cfg.Tuning
	if !c.Active() || !c.sits() {
		return
	}
	if !c.Summoned && !w.chance(t.Council.Cadence) {
		return
	}
	c.Summoned = false
	if !c.launches() {
		w.presenceCouncil(c) // a world, or a thing with no ships: the waking and the unmaking; see waking.go
		w.proposePact(c)
		return
	}
	w.watchRival(c)
	type cand struct {
		e   *Civ
		ap  Appraisal
		far bool
	}
	var cands []cand
	var verdicts []mind.Verdict
	compelled := w.R.Float64() < mind.Compulsion(c.posture() == mind.Conqueror, c.Wis, t)
	for _, eid := range metOf(c) {
		e := w.Civs[eid]
		if !e.Active() || !w.mayWar(c, e, "") || c.Wars[eid] || w.allied(c, e) || c.Truce[eid] > w.Now || (c.Muster != nil && c.Muster.Target == eid) {
			continue // a muster against them is the council's answer already
		}
		if !e.Met[c.ID] && w.perceives(e, c) {
			continue // still only watched from orbit; that is the primitives' matter
		}
		bar, wants, far := w.bar(c, e)
		if !wants {
			continue
		}
		ap := w.appraise(c, e, -1)
		v := w.weigh(c, e, ap, bar, far, compelled)
		if v.Action == mind.Nothing {
			continue
		}
		cands = append(cands, cand{e, ap, far})
		verdicts = append(verdicts, v)
	}
	if i := mind.Council(verdicts); i >= 0 {
		e, ap := cands[i].e, cands[i].ap
		switch {
		case !w.yoke(c, e, ap):
			w.strikeFirst(c, e, ap, cands[i].far, w.cause(c, e))
		case !e.Free():
			// it bent the knee
		default:
			w.strikeFirst(c, e, ap, cands[i].far, because("defiance"))
		}
	}
	if c.Ruled > 0 {
		w.punish(c) // a client that pays short long enough is made war on; see tribute.go
	}
	w.armPlagues(c)
	w.proposePact(c)
	w.deduce(c) // the ledger read for a hole; see gap.go
}

// weigh is the council's view of one enemy, with the scout or the watch
// it calls for done at once.
func (w *World) weigh(c, e *Civ, ap Appraisal, bar float64, far, compelled bool) mind.Verdict {
	v := mind.Judge(mind.JudgeInput{Appraisal: ap.Appraisal, Bar: bar, Far: far, Front: len(ap.Front), Vengeful: c.posture() == mind.Vengeful, Compelled: compelled, Wis: c.Wis, Wary: c.Wary[e.ID]}, w.Cfg.Tuning)
	w.explain(c, "on the "+e.Tok(), v)
	if v.Action != mind.Nothing {
		c.Tally.Judged++
		c.Tally.ActedGap += math.Abs(ap.Acted - ap.Odds)
		if v.Deterred {
			c.Tally.Deterred++
		}
	}
	switch v.Action {
	case mind.ScoutFirst:
		w.maybeScout(c, e)
	case mind.Watch:
		if !c.Watched[e.ID] {
			if w.Now-c.Scouted[e.ID] < 5000 && c.posture() == mind.Conqueror {
				w.event(KStayedHome, c, e, -1, P{})
			}
			c.Watched[e.ID] = true
		}
	}
	return v
}

// consider is the council on one people, at first meeting. Returns whether
// a war followed.
func (w *World) consider(c, e *Civ) bool {
	if !c.Active() || !e.Active() || !c.sits() || !w.mayWar(c, e, "") {
		return false
	}
	bar, wants, far := w.bar(c, e)
	if !wants {
		return false
	}
	ap := w.appraise(c, e, -1)
	v := mind.Judge(mind.JudgeInput{Appraisal: ap.Appraisal, Bar: bar, Far: far, Front: len(ap.Front), Vengeful: c.posture() == mind.Vengeful, Wis: c.Wis}, w.Cfg.Tuning)
	w.explain(c, "at the meeting of the "+e.Tok(), v)
	switch v.Action {
	case mind.Strike:
		return w.strikeFirst(c, e, ap, far, w.cause(c, e))
	case mind.ScoutFirst:
		w.maybeScout(c, e)
	}
	return false
}

// maybeScout sends a scout when a report would change the decision and
// the level can be spared. The Sight reads for free.
func (w *World) maybeScout(c, e *Civ) {
	out := false
	for _, x := range w.fleetsOf(c) {
		if x.Kind == Scout && x.Target == e.ID {
			out = true
			break
		}
	}
	s := mind.Scout(mind.ScoutInput{Sight: c.miracle("foresight") && !c.Searching, Ships: w.standing(c), Fear: c.Dials.Fear, Out: out}, w.Cfg.Tuning)
	w.explain(c, "scouting the "+e.Tok(), s)
	switch {
	case s.Look:
		w.observe(c, e, e.Home, w.Cfg.Tuning.Scout.SightNoise)
	case s.Send:
		w.launch(c, Scout, e, e.Home, 1)
	}
}

// cause is what a posture calls its war.
func (w *World) cause(c, e *Civ) reason {
	switch {
	case c.hates(e):
		return because("extermination")
	case c.Fought[e.ID] > 0:
		return because("old_quarrel")
	case e.Embargo[c.ID]:
		return because("embargo")
	}
	switch c.posture() {
	case mind.Opportunist:
		return because("opportunity")
	case mind.Conqueror:
		return because("conquest")
	case mind.Vengeful:
		return because("revenge")
	}
	return because("border")
}

// strikeFirst opens the war the council chose, with the fleet that opens
// it: at the front if there is one, beyond it if the posture sends fleets
// that far. No war is declared that no fleet follows.
func (w *World) strikeFirst(c, e *Civ, ap Appraisal, far bool, cause reason) bool {
	if !c.launches() {
		return w.presenceStrike(c, e, w.inside(c, e))
	}
	if len(ap.Front) == 0 && !far {
		return false
	}
	return w.maybeCampaign(c, e, cause)
}

// maybeCampaign sizes and sends a fleet against e, declaring war first if
// none is running. Conquerors and the hating try for the home first.
// Nobody sends a fleet that cannot take its first world.
func (w *World) maybeCampaign(c, e *Civ, cause reason) bool {
	_, near := w.nearestEnemy(c, e)
	targets := []int{near}
	if c.posture() == mind.Conqueror || c.hates(e) {
		targets = []int{e.Home, near} // the home if it can be had, else what can
	}
	for _, target := range targets {
		if w.holds(e, target) && !w.protected(c, e, target) && w.sizeCampaign(c, e, cause, target) {
			return true // nobody sails at a star the enemy does not hold
		}
	}
	return false
}

// sizeCampaign sizes a fleet in ships and sends it: from the guard that
// has the ships, or by a muster at the holding nearest the target. A need
// the ships cannot meet is written down as a want, and the docks build
// toward it for a while.
func (w *World) sizeCampaign(c, e *Civ, cause reason, target int) bool {
	k := w.sizeAt(c, e, target)
	w.explain(c, "sizing a fleet against the "+e.Tok()+" at "+w.star(target), k)
	if k.Short && k.LagOK {
		c.WantShips, c.WantSince = k.Need, w.Now
	}
	if !k.Send {
		return false
	}
	if w.guardWith(c, target, k.Share) == nil {
		if c.Muster != nil && c.Muster.Target != e.ID {
			return false // the docks and the guards are gathering for another war
		}
		w.muster(c, e, target, cause, k.Share)
		return true
	}
	if w.warBetween(c.ID, e.ID) == nil {
		w.openWar(c, e, cause, target)
	}
	w.launch(c, Campaign, e, target, k.Share)
	return true
}

// sizeAt is the mind's sizing of a fleet against a world, from what the
// people has manned.
func (w *World) sizeAt(c, e *Civ, target int) mind.Campaign {
	ap := w.appraise(c, e, target)
	return mind.SizeCampaign(mind.CampaignInput{
		Appraisal: ap.Appraisal, Strength: w.strength(c, e), Mil: c.Mil, Bonus: c.warBonus(), Ships: w.standing(c), Total: w.ships(c),
		Risk: c.Dials.Risk, Fear: c.Dials.Fear, Conqueror: c.posture() == mind.Conqueror,
	}, w.Cfg.Tuning)
}

// nearestEnemy is e's world nearest to any of c's, and the distance: for
// a horde, the nearest of its fleets' bases.
func (w *World) nearestEnemy(c, e *Civ) (float64, int) {
	best, bd := e.Home, 1e9
	for _, s := range w.holdings(e) {
		if w.protected(c, e, s) {
			continue
		}
		if _, d := w.nearest(c, s); d < bd {
			best, bd = s, d
		}
	}
	return bd, best
}

// watchRival is the council naming its rival: the neighbour it most
// fears in reach (threat), which its docks build against (want) and its
// pickets watch hardest (picket). A change of rival is noted, so the
// record holds the build-up of a cold war.
func (w *World) watchRival(c *Civ) {
	id := -1
	if e := w.threat(c); e != nil {
		id = e.ID
	}
	if id != c.Rival {
		c.Rival = id
		w.note(KRival, c, nil, -1, P{"rival": id})
	}
}
