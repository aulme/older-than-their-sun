package history

import (
	"worldgen/internal/flow"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// The eldritch: a people with no tree (species.Eldritch, or anything whose
// profile denies Researches). In place of research it holds powers from
// the pool (species/pool.go): a few from birth, and each tick at a small
// chance it deepens, one more, told as a change in what it is. Each power
// is a level in a domain through the profile; the ones that stand for a
// miracle bring that miracle and its filter, the hunger brings Overshoot,
// the wound tears the wall. It does not settle by ship: with a second
// presence, another of it is simply there. The tithe takes a share of
// every neighbour's harvest; the mirror answers in the sender's own
// voice; the long sleep is a dormancy that ends when somebody settles
// too close. Nothing here names the substrate: every rule reads the
// profile or the powers, and a made people of any substrate that cannot
// research would live the same way.

// mirrorWis is what the mirror takes off the difference toward its holder.
const mirrorWis = 2

// eldritchStep is the civ step in research's place for a people with no
// tree: appearing, deepening, and the wound's wear on the wall.
func (w *World) eldritchStep(c *Civ) {
	if c.Species.Profile().Can(species.Researches) || !c.Active() || c.Asleep {
		return
	}
	w.appear(c)
	w.deepen(c)
	if c.Species.HasPower("wound") {
		w.Thin += w.Cfg.Tuning.Kinds.WoundWear * w.dt
	}
}

// appear is the second presence: with a small chance its powers raise, a
// star within reach becomes its, and nobody saw anything cross.
func (w *World) appear(c *Civ) {
	t := &w.Cfg.Tuning.Kinds
	p := c.Species.Profile()
	if !c.Species.HasPower("presence") || c.Aloft || !c.Free() || c.Reach < 1 {
		return
	}
	if p.Worlds > 0 && len(c.Systems) >= p.Worlds {
		return
	}
	if !w.chance(t.Appear * c.expandMul(w)) {
		return
	}
	var cands []int
	for _, from := range c.Systems {
		for _, s := range w.G.Near(from, c.Reach) {
			if w.Owner[s] < 0 && s != w.G.Sol && w.canLive(c, s) && !contains(cands, s) {
				cands = append(cands, s)
			}
		}
	}
	if len(cands) == 0 {
		return
	}
	s := w.pick(cands)
	w.holdWorld(c, s)
	c.Tally.Appeared++
	w.told(FAppeared, c, nil, s)
	w.afterHold(c, s)
}

// deepen is one more power from the pool, at the decision's rate: the
// base times one plus a quarter per power held, so what has grown grows.
func (w *World) deepen(c *Civ) {
	if !w.chance(w.Cfg.Tuning.Kinds.Deepen * (1 + float64(len(c.Species.Powers))/4)) {
		return
	}
	if p := c.Species.DrawPower(w.R); p != nil {
		w.gainPower(c, p)
	}
}

// gainPower is a deepening: the power, the line, the fact, and what it
// brings.
func (w *World) gainPower(c *Civ, p *species.Power) {
	if !c.Species.AddPower(p.Key) {
		return
	}
	c.Tally.Deepened++
	w.stir(c)
	w.told(FDeepened, c, nil, c.Home).with(P{"power": p.Key})
	w.bring(c, p, true)
}

// bring is what a power brings with it: the miracle it stands for, as
// had, with that miracle's filter faced if fire says so; a filter of its
// own; the wound's tear. The powers a people is born with are had as a
// born miracle is had, with nothing to fall from.
func (w *World) bring(c *Civ, p *species.Power, fire bool) {
	switch {
	case p.Node != "":
		n := tech.Get(p.Node)
		c.Known[n.Key] = true
		how := "deepening"
		if !fire {
			how = "born"
		}
		w.gain(c, n.Key, how)
		if n.Filter != "" {
			if fire {
				w.face(c, n.Filter, 0)
			} else {
				c.Faced[n.Filter] = true
			}
		}
	case p.Filter != "":
		w.recompute(c)
		if fire {
			w.face(c, p.Filter, 0)
		}
	case p.Key == "wound":
		w.tear(0.5)
		w.name(c, "wound")
	}
	w.recompute(c)
}

// bornPowers is what a people has from birth of the pool: had, with
// nothing to fall from.
func (w *World) bornPowers(c *Civ) {
	for _, k := range c.Species.Powers {
		if p := species.PowerByKey(k); p != nil {
			w.bring(c, p, false)
		}
	}
}

// tithed is the tithe: every people holding one takes a share of the
// harvest of every neighbour with a holding within its reach, as a
// tribute nobody agreed to. The neighbour resents it a little more each
// tick; the first taking is a fact.
func (w *World) tithed(c *Civ, in flow.Income) flow.Income {
	t := &w.Cfg.Tuning.Kinds
	for _, e := range w.Civs {
		if e == c || !e.Active() || e.Asleep || !e.Species.HasPower("tithe") {
			continue
		}
		near := false
		for _, s := range c.Systems {
			if w.inReach(e, s) {
				near = true
				break
			}
		}
		if !near {
			continue
		}
		take := in.Scale(t.Tithe)
		in = in.Less(take)
		e.Loot.Add(take)
		c.resent(e.ID, t.TitheGrudge*w.dt)
		if c.tithedBy == nil {
			c.tithedBy = map[int]bool{}
		}
		if !c.tithedBy[e.ID] {
			c.tithedBy[e.ID] = true
			e.Tally.Tithed++
			w.told(FTithed, e, c, c.Home)
		}
	}
	return in
}

// tithes counts the live tithes, for the hazard.
func (w *World) tithes() int {
	n := 0
	for _, c := range w.Civs {
		if c.Active() && !c.Asleep && c.Species.HasPower("tithe") {
			n++
		}
	}
	return n
}

// mirrored is a people's first word with a mirror: what answers is its
// own voice, older, and some of them listen. The signal's scar, at a
// chance, in the sender.
func (w *World) mirrored(a, b *Civ) {
	t := &w.Cfg.Tuning.Kinds
	for _, pair := range [][2]*Civ{{a, b}, {b, a}} {
		s, m := pair[0], pair[1]
		if !m.Species.HasPower("mirror") || !s.Species.Profile().Can(species.Believes) || s.Scars[ScarSignal] || !s.Active() {
			continue
		}
		if w.R.Float64() >= t.Mirror {
			continue
		}
		s.Scars[ScarSignal] = true
		s.Morale -= 1
		w.event(KMirrored, s, m, -1, P{})
		w.recompute(s)
	}
}

// sleep is the long sleep, for a people whose profile says Dormant (the
// long sleep of the pool, or a thing that eats): it sleeps when its will
// is spent, and sits every tick out but its guns until disturbed.
func (w *World) sleep(c *Civ) {
	if c.Asleep || !c.Active() {
		return
	}
	c.Asleep = true
	c.Slept = w.Now
	c.Tally.Sleeps++
	c.Voyages = nil
	w.event(KSlept, c, nil, -1, P{})
}

// rouse is the sleeper disturbed: by whoever settled inside its reach,
// brought a war to it, or unleashed it; by nothing, when the wall is
// thin. A world wakes on everything inside its neighbourhood, the
// disturber first (waking.go); anything else declares war on the
// disturber.
func (w *World) rouse(c, by *Civ) {
	if !c.Asleep {
		return
	}
	if by == c {
		by = nil // its own doing: the Find beneath its own cities
	}
	c.Asleep = false
	c.Tally.Wakings++
	w.recompute(c)
	w.event(KRoused, c, by, -1, P{})
	if by != nil && by.Active() && !(c.Met[by.ID] && by.Met[c.ID]) {
		c.Met[by.ID], by.Met[c.ID] = true, true
		w.meeting(c, by, c.Home, "touch")
	}
	if w.canWake(c) {
		w.wakeOnAll(c, by)
		return
	}
	if by != nil && by.Active() && c.Truce[by.ID] <= w.Now {
		w.declare(c, by, because("sleep_disturbed")) // not again within the truce of the last time: it stirs, and that is all
	}
}

// wakeOnAll is the waking on everything inside the neighbourhood: the
// disturber first, then every other people with a world inside it.
func (w *World) wakeOnAll(c, first *Civ) {
	var order []*Civ
	if first != nil && first.Active() {
		order = append(order, first)
	}
	for _, e := range w.Civs {
		if e != c && e != first && e.Active() {
			order = append(order, e)
		}
	}
	for _, e := range order {
		if !c.Active() || c.Asleep {
			return
		}
		if worlds := w.inside(c, e); len(worlds) > 0 && e.Active() {
			if !(c.Met[e.ID] && e.Met[c.ID]) {
				w.meeting(c, e, worlds[0], "touch")
			}
			c.Met[e.ID], e.Met[c.ID] = true, true
			w.waking(c, e, worlds)
		}
	}
}

// spent is a war ended with a side's will gone: a people that sleeps
// sleeps.
func (w *World) spent(wr *War) {
	for i, id := range wr.Sides {
		c := w.Civs[id]
		if c.Species.Profile().Dormant && wr.Will[i] <= 0 && len(c.Wars) == 0 {
			w.sleep(c)
		}
	}
}

// sleeperAt is a sleeper: an eldritch thing, one world and no one home,
// with the long sleep, asleep at the star until disturbed. From the
// deep pass as a legacy, or from the Door's scar; the legacy of kind
// Sleeper points at it.
func (w *World) sleeperAt(star int, made Origin) *Civ {
	sp := species.GenerateWith(w.R, w.G.Stars[star].Mult, w.G.Sys[star].Arch, species.Eldritch, species.Planetary|species.Unconscious)
	sp.AddPower("sleep")
	c := w.ariseAt(star, sp, made)
	if c == nil {
		return nil
	}
	c.Origin = made
	w.sleep(c)
	return c
}

// disturbed is a world newly held: every sleeper with the star inside
// its reach wakes on whoever holds it.
func (w *World) disturbed(c *Civ, t int) {
	for _, e := range w.Civs {
		if e != c && e.Asleep && e.Active() && w.inReach(e, t) {
			w.rouse(e, c)
		}
	}
}

// sleeperStar is where something that came through a people's door goes
// to sleep: the nearest star to one of its worlds that nobody holds, or
// -1 when there is none within a hop.
func (w *World) sleeperStar(c *Civ) int {
	from := w.aWorld(c)
	for _, s := range w.G.Near(from, 20) {
		if w.Owner[s] < 0 && s != w.G.Sol {
			return s
		}
	}
	return -1
}
