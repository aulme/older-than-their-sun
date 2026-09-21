package history

import (
	"sort"

	"worldgen/internal/mind"
	"worldgen/internal/species"
)

// The waking: how a people that sends no fleets makes war. A living
// world (species.Planetary, or anything whose profile fixes a
// neighbourhood) is godlike within its reach and nothing outside it: its
// strikes on worlds inside the neighbourhood are faced as the waking, a
// survival filter, not fought as battles, and a conscious one warns
// before it wakes: a demand through the council first, and the waking
// only if the demand is refused. An eldritch thing with the unmaking
// ends a world inside its reach without touching it. Everything else
// about such a people's wars is the war engine's: others come for it by
// fleet, its body stands as guns (guns.go), and the war drains as any
// war does.

// presenceCouncil is the council of a people that cannot launch: each
// people holding worlds inside its reach is judged as any enemy is, its
// body counted as what it fights with; the best that clears its bar is
// struck at by the waking or the unmaking.
func (w *World) presenceCouncil(c *Civ) {
	t := w.Cfg.Tuning
	if !w.canWake(c) && !c.miracle("unmaking") {
		return
	}
	w.unmakings(c)
	type cand struct {
		e      *Civ
		worlds []int
	}
	var cands []cand
	var verdicts []mind.Verdict
	compelled := w.R.Float64() < mind.Compulsion(c.posture() == mind.Conqueror, c.Wis, t)
	for _, eid := range sortedInts(c.Met) {
		e := w.Civs[eid]
		if !e.Active() || !e.Free() || c.Wars[eid] || w.allied(c, e) || c.Truce[eid] > w.Now || !e.Met[c.ID] {
			continue
		}
		worlds := w.inside(c, e)
		if len(worlds) == 0 {
			continue
		}
		bar, wants, _ := w.bar(c, e)
		if !wants {
			continue
		}
		ap := w.appraise(c, e, worlds[0])
		v := w.weigh(c, e, ap, bar, false, compelled)
		if v.Action == mind.Nothing {
			continue
		}
		cands = append(cands, cand{e, worlds})
		verdicts = append(verdicts, v)
	}
	if i := mind.Council(verdicts); i >= 0 {
		w.presenceStrike(c, cands[i].e, cands[i].worlds)
	}
}

// canWake says whether a people's strikes are the waking: it has a
// neighbourhood and is not asleep.
func (w *World) canWake(c *Civ) bool {
	return c.Species.Profile().Neighbourhood > 0 && !c.Asleep
}

// inside is e's worlds within c's reach of any holding of c, nearest
// first.
func (w *World) inside(c, e *Civ) []int {
	var out []int
	for _, s := range e.Systems {
		if w.inReach(c, s) {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		_, di := w.nearest(c, out[i])
		_, dj := w.nearest(c, out[j])
		return di < dj
	})
	return out
}

// presenceStrike is the strike a people that sends no fleets makes: a
// conscious world demands first and wakes when the demand has stood
// refused; an unconscious one wakes; an eldritch thing with the
// unmaking unmakes the nearest world. Returns whether a war followed.
func (w *World) presenceStrike(c, e *Civ, worlds []int) bool {
	t := &w.Cfg.Tuning.Kinds
	if len(worlds) == 0 {
		return false
	}
	switch {
	case w.canWake(c) && !c.Species.Profile().NoOne:
		if since, ok := c.demanded[e.ID]; !ok || w.Now-since < Year(t.Demand)*w.Cfg.Step {
			if !ok {
				return w.demand(c, e, worlds)
			}
			return false // the demand stands; the waking waits on it
		}
		w.waking(c, e, worlds)
		return true
	case w.canWake(c):
		w.waking(c, e, worlds)
		return true
	case c.miracle("unmaking"):
		// no war is opened that nothing follows: the unmaking is its opening stroke, and it rests between one world and the next
		if w.Now-c.LastUnmade < Year(t.UnmakeRest)*w.Cfg.Step {
			return false
		}
		w.unmake(c, e, worlds[0])
		return true
	}
	return false
}

// unmakings is the unmaking in a running war: at rest for a while after
// each world, the nearest world of an enemy inside reach is unmade.
func (w *World) unmakings(c *Civ) {
	t := &w.Cfg.Tuning.Kinds
	if !c.miracle("unmaking") || w.Now-c.LastUnmade < Year(t.UnmakeRest)*w.Cfg.Step {
		return
	}
	for _, eid := range sortedInts(c.Wars) {
		e := w.Civs[eid]
		if !e.Active() {
			continue
		}
		if worlds := w.inside(c, e); len(worlds) > 0 {
			w.unmake(c, e, worlds[0])
			return
		}
	}
}

// demand is a living world telling a people to leave the worlds it holds
// inside the neighbourhood. The told people heeds by its fear, the gap
// and its posture, and leaves; or refuses, and the demand stands until
// the waking. A home cannot be left.
func (w *World) demand(c, e *Civ, worlds []int) bool {
	if c.demanded == nil {
		c.demanded = map[int]Year{}
	}
	c.demanded[e.ID] = w.Now
	c.Tally.Demands++
	f := w.unplaced(FDemand, c, e, worlds[0])
	home := contains(worlds, e.Home)
	gap := c.Mil + mind.ShipLevels(float64(w.bodyGuns(c))) - w.levelOf(e)
	h := mind.Heed(mind.HeedInput{Posture: e.posture(), Fear: e.Dials.Fear, Risk: e.Dials.Risk, Gap: gap, Grudge: e.Grudge[c.ID] > 0}, w.Cfg.Tuning)
	w.explain(e, "told to leave by the "+c.Tok(), h)
	if home || w.R.Float64() >= h.Chance {
		w.place(f.with(P{"outcome": "refused", "home": home, "seat": e.Home}))
		e.resent(c.ID, 0.5)
		e.Summoned = true
		return false
	}
	f.with(P{"outcome": "left", "home": false, "seat": e.Home})
	for _, s := range worlds {
		w.loseSystem(e, s, "abandoned "+e.Species.Flavour().Colony, "")
	}
	w.place(f)
	delete(c.demanded, e.ID)
	return false
}

// waking is a living world striking: a war if none is running, and the
// waking faced by the people whose worlds lie inside the neighbourhood,
// at a difficulty that carries the world's Military and whether they can
// flee.
func (w *World) waking(c, e *Civ, worlds []int) {
	t := &w.Cfg.Tuning.Kinds
	if w.warBetween(c.ID, e.ID) == nil {
		if w.declare(c, e, "the waking") == nil {
			return
		}
	}
	delete(c.demanded, e.ID)
	c.Tally.Wakings++
	w.told(FWaking, c, e, worlds[0]).with(P{"seat": c.Home}).N = len(worlds)
	w.blastWorlds = append([]int(nil), worlds...)
	w.blastWhat = "the " + c.Tok()
	adj := t.WakingBase + t.WakingMil*c.Mil
	if e.Reach < 12 {
		adj += t.WakingYoung
	}
	w.face(e, "waking", adj)
	w.event(KStillAgain, c, nil, -1, P{})
}

// unmake is the unmaking turned on a world of another people inside its
// holder's reach: the world is not there any more. A war if none is
// running; the wall wears.
func (w *World) unmake(c, e *Civ, s int) {
	if w.warBetween(c.ID, e.ID) == nil {
		if w.declare(c, e, "the unmaking") == nil {
			return
		}
	}
	wr := w.warBetween(c.ID, e.ID)
	c.Tally.Unmade++
	c.Tally.Taken++
	c.LastUnmade = w.Now
	w.tear(0.3)
	w.Bio[s] = BioNone
	w.told(FUnmade, c, e, s)
	wasHome := s == e.Home
	w.loseSystem(e, s, "unmade world", sprintf("were unmade by the %s", c.Tok()))
	if wr != nil && !wr.Over {
		i := wr.side(c.ID)
		wr.Taken[i]++
		wr.Lost[1-i]++
		wr.Will[1-i] -= 0.5
		if !e.Active() {
			w.endWar(wr, "extinction")
		} else if wasHome {
			wr.Will[1-i] -= 0.5
		}
	}
}

// launches says whether a people sends fleets at all.
func (c *Civ) launches() bool { return c.Species.Profile().Can(species.Launches) }

func init() {
	def(&Filter{
		Key: "waking", Name: "the waking", Levels: []string{"sur"}, Diff: 5, Repeat: true, Domain: "propulsion",
		Overcome: func(w *World, c *Civ) {
			w.faced(c, "waking", "overcome", "", -1).with(P{"what": w.blastWhat})
		},
		Scar: func(w *World, c *Civ) {
			for _, s := range w.blastWorlds {
				if s != c.Home {
					w.loseSystem(c, s, "silent world", "")
				}
			}
			w.faced(c, "waking", "scarred", "", -1).with(P{"what": w.blastWhat})
		},
		Decline: func(w *World, c *Civ) {
			homeHit := contains(w.blastWorlds, c.Home)
			for _, s := range w.blastWorlds {
				if s != c.Home {
					w.loseSystem(c, s, "silent world", sprintf("were unmade by %s", w.blastWhat))
				}
			}
			if !c.Active() {
				return
			}
			if homeHit {
				if len(c.Systems) > 1 && c.Reach >= 12 {
					w.leaveHome(c, "flee "+w.blastWhat)
				} else {
					w.loseSystem(c, c.Home, "silent world", sprintf("were unmade by %s", w.blastWhat))
				}
			} else {
				w.contract(c, sprintf("withdrew to %s after %s took their worlds", w.star(c.Home), w.blastWhat))
			}
		},
	})
}
