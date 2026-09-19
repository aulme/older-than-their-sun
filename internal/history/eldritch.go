package history

import (
	"strings"

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
			if w.Owner[s] < 0 && w.Held[s] < 0 && s != w.G.Sol && w.canLive(c, s) && !contains(cands, s) {
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
	w.fact(FAppeared, c, nil, s)
	w.log("Another of the %s is at %s. Nothing was seen to cross.", c.Name, w.star(s))
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
	f := w.fact(FDeepened, c, nil, c.Home)
	f.What = p.Name
	w.log("%s", capitalise(strings.ReplaceAll(p.Line, "{S}", "the "+c.Name)))
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
			w.fact(FTithed, e, c, c.Home)
			w.log("Something is taken from every harvest of the %s within reach of the %s. Nobody agreed to it, and nothing can be found to refuse.", c.Name, e.Name)
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
		w.log("The %s speak to the %s, and what answers is their own voice, older than they are. A cult of the signal grows among them and is never quite rooted out.", s.Name, m.Name)
		w.recompute(s)
	}
}

// sleep is the long sleep: it sleeps when its will is spent, and sits
// every tick out but its guns until disturbed.
func (w *World) sleep(c *Civ) {
	if c.Asleep || !c.Active() {
		return
	}
	c.Asleep = true
	c.Slept = w.Now
	c.Tally.Sleeps++
	c.Voyages = nil
	w.log("The %s go still. There is nothing left they want, and nothing near them moves. They sleep.", c.Name)
}

// rouse is the sleeper disturbed: by whoever settled inside its reach, on
// whom a world wakes (waking.go) and anything else declares war.
func (w *World) rouse(c, by *Civ) {
	if !c.Asleep {
		return
	}
	c.Asleep = false
	c.Tally.Wakings++
	w.recompute(c)
	w.log("Something settled too close, and the %s wake.", c.Name)
	if by == nil || !by.Active() {
		return
	}
	c.Met[by.ID], by.Met[c.ID] = true, true
	if c.Species.Profile().Neighbourhood > 0 {
		if worlds := w.inside(c, by); len(worlds) > 0 {
			w.waking(c, by, worlds)
			return
		}
	}
	w.declare(c, by, "the disturbing of its sleep")
}

// spent is a war ended with a side's will gone: a people with the long
// sleep sleeps.
func (w *World) spent(wr *War) {
	for i, id := range wr.Sides {
		c := w.Civs[id]
		if c.Species.HasPower("sleep") && wr.Will[i] <= 0 && len(c.Wars) == 0 {
			w.sleep(c)
		}
	}
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
