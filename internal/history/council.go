package history

import "worldgen/internal/mind"

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
	if !c.Active() || !c.Free() {
		return
	}
	if !c.Summoned && !w.chance(t.Council.Cadence) {
		return
	}
	c.Summoned = false
	type cand struct {
		e   *Civ
		ap  Appraisal
		far bool
	}
	var cands []cand
	var verdicts []mind.Verdict
	compelled := c.posture() == mind.Conqueror && w.R.Float64() < t.Council.Compulsion
	for _, eid := range sortedInts(c.Met) {
		e := w.Civs[eid]
		if !e.Active() || !e.Free() || c.Wars[eid] || w.allied(c, e) || c.Truce[eid] > w.Now {
			continue
		}
		if !e.Met[c.ID] {
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
		w.strikeFirst(c, cands[i].e, cands[i].ap, cands[i].far)
	}
	w.proposePact(c)
}

// weigh is the council's view of one enemy, with the scout or the watch
// it calls for done at once.
func (w *World) weigh(c, e *Civ, ap Appraisal, bar float64, far, compelled bool) mind.Verdict {
	v := mind.Judge(mind.JudgeInput{Appraisal: ap.Appraisal, Bar: bar, Far: far, Front: len(ap.Front), Vengeful: c.posture() == mind.Vengeful, Compelled: compelled}, w.Cfg.Tuning)
	w.explain(c, "on the "+e.Name, v)
	switch v.Action {
	case mind.ScoutFirst:
		w.maybeScout(c, e)
	case mind.Watch:
		if !c.Watched[e.ID] {
			if w.Now-c.Scouted[e.ID] < 5000 && c.posture() == mind.Conqueror {
				w.log("The %s look hard at the %s, and stay home.", c.Name, e.Name)
			}
			c.Watched[e.ID] = true
		}
	}
	return v
}

// consider is the council on one people, at first meeting. Returns whether
// a war followed.
func (w *World) consider(c, e *Civ) bool {
	if !c.Active() || !e.Active() || !c.Free() || !e.Free() {
		return false
	}
	bar, wants, far := w.bar(c, e)
	if !wants {
		return false
	}
	ap := w.appraise(c, e, -1)
	v := mind.Judge(mind.JudgeInput{Appraisal: ap.Appraisal, Bar: bar, Far: far, Front: len(ap.Front), Vengeful: c.posture() == mind.Vengeful}, w.Cfg.Tuning)
	w.explain(c, "at the meeting of the "+e.Name, v)
	switch v.Action {
	case mind.Strike:
		return w.strikeFirst(c, e, ap, far)
	case mind.ScoutFirst:
		w.maybeScout(c, e)
	}
	return false
}

// maybeScout sends a scout when a report would change the decision and
// the level can be spared. The Sight reads for free.
func (w *World) maybeScout(c, e *Civ) {
	out := false
	for _, x := range w.Expeditions {
		if !x.Over && x.Kind == Scout && x.Owner == c.ID && x.Target == e.ID {
			out = true
			break
		}
	}
	s := mind.Scout(mind.ScoutInput{Sight: c.miracle("foresight") && !c.Searching, Ships: w.standing(c), Fear: c.Dials.Fear, Out: out}, w.Cfg.Tuning)
	w.explain(c, "scouting the "+e.Name, s)
	switch {
	case s.Look:
		w.observe(c, e, e.Home, w.Cfg.Tuning.Scout.SightNoise)
	case s.Send:
		w.launch(c, Scout, e, e.Home, 1)
	}
}

// cause is what a posture calls its war.
func (w *World) cause(c, e *Civ) string {
	switch {
	case c.hates(e):
		return "extermination"
	case c.Fought[e.ID] > 0:
		return "the old quarrel"
	case e.Embargo[c.ID]:
		return "the embargo"
	}
	switch c.posture() {
	case mind.Opportunist:
		return "opportunity"
	case mind.Conqueror:
		return "conquest"
	case mind.Vengeful:
		return "revenge"
	}
	return "a border"
}

// strikeFirst opens the war the council chose: at the front if there is
// one, by fleet if the posture sends fleets and the sizing allows.
func (w *World) strikeFirst(c, e *Civ, ap Appraisal, far bool) bool {
	cause := w.cause(c, e)
	if len(ap.Front) > 0 {
		w.declare(c, e, cause)
		if far && !w.inReach(c, e.Home) {
			w.maybeCampaign(c, e, cause)
		}
		return true
	}
	if !far {
		return false
	}
	return w.maybeCampaign(c, e, cause)
}

// maybeCampaign sizes and sends a fleet against e, declaring war first if
// none is running. Conquerors and the hating try for the home first.
// Nobody sends a fleet that cannot take its first world.
func (w *World) maybeCampaign(c, e *Civ, cause string) bool {
	_, near := w.nearestEnemy(c, e)
	targets := []int{near}
	if c.posture() == mind.Conqueror || c.hates(e) {
		targets = []int{e.Home, near} // the home if it can be had, else what can
	}
	for _, target := range targets {
		if w.sizeCampaign(c, e, cause, target) {
			return true
		}
	}
	return false
}

// sizeCampaign sizes a fleet in ships and sends it. A need the ships
// cannot meet is written down as a want, and the docks build toward it
// for a while; the muster lands with garrisons.
func (w *World) sizeCampaign(c, e *Civ, cause string, target int) bool {
	ap := w.appraise(c, e, target)
	k := mind.SizeCampaign(mind.CampaignInput{
		Appraisal: ap.Appraisal, Strength: w.strength(c, e), Mil: c.Mil, Bonus: c.warBonus(), Ships: w.standing(c), Total: w.ships(c),
		Risk: c.Dials.Risk, Fear: c.Dials.Fear, Conqueror: c.posture() == mind.Conqueror,
	}, w.Cfg.Tuning)
	w.explain(c, "sizing a fleet against the "+e.Name+" at "+w.star(target), k)
	if k.Short && k.LagOK {
		c.WantShips, c.WantSince = k.Need, w.Now
	}
	if !k.Send {
		return false
	}
	if w.warBetween(c.ID, e.ID) == nil {
		w.declare(c, e, cause)
	}
	w.launch(c, Campaign, e, target, k.Share)
	return true
}

// nearestEnemy is e's world nearest to any of c's, and the distance.
func (w *World) nearestEnemy(c, e *Civ) (float64, int) {
	best, bd := e.Home, 1e9
	for _, s := range e.Systems {
		if _, d := w.nearest(c, s); d < bd {
			best, bd = s, d
		}
	}
	return bd, best
}
