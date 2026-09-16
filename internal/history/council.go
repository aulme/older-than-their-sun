package history

// The council is where a people decides: whom to strike, whether to send a
// fleet or a scout first, whom to ask for a pact. It sits on a cadence and
// whenever an event summons it. Every option is scored through the
// appraisal; posture sets the bar; a little folly is wanted.

func (w *World) council(c *Civ) {
	if !c.Active() || !c.Free() {
		return
	}
	if !c.Summoned && !w.chance(0.3) {
		return
	}
	c.Summoned = false
	type cand struct {
		e   *Civ
		ap  Appraisal
		bar float64
		far bool
	}
	var best *cand
	compelled := c.posture() == "conqueror" && w.R.Float64() < 0.1
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
		if compelled {
			bar = min(bar, 0.25)
		}
		ap := w.appraise(c, e, -1)
		if len(ap.Front) == 0 && !far {
			continue
		}
		if c.posture() == "vengeful" {
			ap.Acted = ap.High // against the grudge target the cost is not counted
		}
		if w.Cfg.TraceAI {
			w.log("[council of the %s on the %s: odds %.2f (%.2f to %.2f), acting on %.2f, bar %.2f, front %d, lag %s]", c.Name, e.Name, ap.Odds, ap.Low, ap.High, ap.Acted, bar, len(ap.Front), span(Year(ap.Lag)))
		}
		switch {
		case ap.Acted >= bar:
			if best == nil || ap.Acted-bar > best.ap.Acted-best.bar {
				best = &cand{e, ap, bar, far}
			}
		case ap.Low < bar && ap.High > bar:
			w.maybeScout(c, e)
		case !c.Watched[eid]:
			if w.Now-c.Scouted[eid] < 5000 && c.posture() == "conqueror" {
				w.log("The %s look hard at the %s, and stay home.", c.Name, e.Name)
			}
			c.Watched[eid] = true
		}
	}
	if best != nil {
		w.strikeFirst(c, best.e, best.ap, best.far)
	}
	w.proposePact(c)
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
	if len(ap.Front) == 0 && !far {
		return false
	}
	if c.posture() == "vengeful" {
		ap.Acted = ap.High
	}
	if ap.Acted >= bar {
		return w.strikeFirst(c, e, ap, far)
	}
	if ap.Low < bar && ap.High > bar {
		w.maybeScout(c, e)
	}
	return false
}

// maybeScout sends a scout when a report would change the decision and
// the level can be spared. The Sight reads for free.
func (w *World) maybeScout(c, e *Civ) {
	if c.miracle("foresight") && !c.Searching {
		w.observe(c, e, e.Home, 0.1)
		return
	}
	if c.Mil-1 < 1 || (c.Dials.Fear > 0.8 && c.Mil < 4) {
		return
	}
	for _, x := range w.Expeditions {
		if !x.Over && x.Kind == Scout && x.Owner == c.ID && x.Target == e.ID {
			return
		}
	}
	w.launch(c, Scout, e, e.Home, 1)
}

// cause is what a posture calls its war.
func (w *World) cause(c, e *Civ) string {
	switch {
	case c.hates(e):
		return "extermination"
	case c.Fought[e.ID] > 0:
		return "the old quarrel"
	}
	switch c.posture() {
	case "opportunist":
		return "opportunity"
	case "conqueror":
		return "conquest"
	case "vengeful":
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
// none is running. Nobody sends a fleet that cannot take its first world.
func (w *World) maybeCampaign(c, e *Civ, cause string) bool {
	_, near := w.nearestEnemy(c, e)
	targets := []int{near}
	if c.posture() == "conqueror" || c.hates(e) {
		targets = []int{e.Home, near} // the home if it can be had, else what can
	}
	for _, target := range targets {
		if w.sizeCampaign(c, e, cause, target) {
			return true
		}
	}
	return false
}

func (w *World) sizeCampaign(c, e *Civ, cause string, target int) bool {
	ap := w.appraise(c, e, target)
	def := w.strength(c, e) - ap.Margin
	need := def - (2*c.Dials.Risk-1)*ap.Spread // even odds on what they believe, tilted by risk
	total := c.Mil + c.Away
	floor := max(0.1*total, 1)
	cap := c.Mil * (1 - 0.4*c.Dials.Fear)
	share := min(max(need, floor), cap)
	lagOK := ap.Lag < 20_000 || (c.posture() == "conqueror" && ap.Lag < 40_000)
	if w.Cfg.TraceAI {
		w.log("[the %s size a fleet against the %s at %s: need %.1f, floor %.1f, cap %.1f, crossing %s]", c.Name, e.Name, w.star(target), need, floor, cap, span(Year(ap.Lag)))
	}
	if share < need || share < floor || !lagOK {
		return false
	}
	if w.warBetween(c.ID, e.ID) == nil {
		w.declare(c, e, cause)
	}
	w.launch(c, Campaign, e, target, share)
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
