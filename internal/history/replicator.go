package history

import (
	"worldgen/internal/flow"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// The replicator: a people whose profile says it grows by eating
// (species.Profile.Eats; the replicator modifier, or anything made that
// way). It has no uses and feeds nothing: every world it holds turns the
// yield of what its body is made of into ships, standing where they were
// grown, and a world it takes in war is stripped of whoever was there
// and held empty, to be eaten in turn. It makes peace only by exhaustion
// (war.go's judge reads NoTerms) and when its will is spent it goes
// quiet: the long sleep of eldritch.go, since its profile says Dormant,
// woken by whoever settles inside its reach, by a war brought to it, or
// by the Find. A monster to everyone by rule (lore.go's monster). Nothing
// here names the substrate: a biological one eats flesh and a machine one
// metal, by the body the profile gives it, and one made of neither eats
// both.

// bodyKinds is what a body is made of, as the flow kinds it eats.
func bodyKinds(body string) []flow.Kind {
	switch body {
	case "organic":
		return []flow.Kind{flow.O}
	case "metal":
		return []flow.Kind{flow.M}
	}
	return []flow.Kind{flow.O, flow.M}
}

// eat is the civ step after the flows for a people that grows by eating:
// at every world held, what is there of its body's matter becomes ships
// at the decision's rate, whatever the people knows how to harness,
// since it does not harness anything: it eats it. A world is finite:
// from the first bite its sources wear to nothing over EatOut years, for
// the eater and for whoever holds the world after, so what eats must
// move on or go quiet.
func (w *World) eat(c *Civ) {
	p := c.Species.Profile()
	if !p.Eats || c.Aloft {
		return
	}
	t := &w.Cfg.Tuning.Kinds
	kinds := bodyKinds(p.Body)
	for _, s := range c.Systems {
		food := 0.0
		for _, id := range w.sourcesAt[s] {
			src := w.Sources[id]
			if src.Legacy >= 0 || src.Mobile {
				continue
			}
			if src.Wear == 0 {
				src.Since, src.Wear = w.Now, Year(t.EatOut) // the first bite: the world is being eaten out
			}
			y := src.Yield
			if src.Wear > 0 {
				y = y.Scale(max(0, 1-float64(w.Now-src.Since)/float64(src.Wear)))
			}
			for _, k := range kinds {
				food += y[k]
			}
		}
		if n := w.count(t.Eat * food); n > 0 {
			w.addGuard(c, s, n)
			c.Tally.Eaten += n
			c.Tally.Built += n
			if !c.firstShip {
				c.firstShip = true
				w.log("At %s the %s have begun to make more of themselves out of what is there.", w.star(s), c.Name)
			}
		}
	}
	if n := w.ships(c); n > c.PeakShips {
		c.PeakShips = n
	}
}

// consume is a world taken in war by a people that eats: whoever was
// there is stripped from it, as a horde strips, and the world is held
// empty, since what is there now is more of the taker. The ordinary war
// facts: a taking, and the loss.
func (w *World) consume(wr *War, c, e *Civ, t int) {
	i := wr.side(c.ID)
	if g := w.guardAt(e, t); g != nil {
		g.Over = true
	}
	w.loseSystem(e, t, "stripped world", sprintf("were consumed by the %s", c.Name))
	if w.Owner[t] >= 0 {
		return // somebody else's now: a fleeing people's, or a rider's
	}
	w.Owner[t] = c.ID
	c.Systems = append(c.Systems, t)
	c.Peak = max(c.Peak, len(c.Systems))
	wr.Taken[i]++
	wr.Lost[1-i]++
	c.Tally.Taken++
	c.Tally.Consumed++
	e.Tally.Lost++
	wr.Will[i] += 0.3
	wr.Will[1-i] -= 0.3
	w.fact(FTaken, c, e, t)
	w.log("The %s take %s from the %s and strip it. Nothing that was there is left; what is there now is more of the %s.", c.Name, w.star(t), e.Name, c.Name)
	if wr.Name == "" {
		wr.Name = "the war of " + w.star(t)
	}
	if e.Active() && (wr.Lost[1-i] == 1 || wr.Lost[1-i]%3 == 0) {
		w.face(e, "hold", 0)
	}
	if !e.Active() {
		w.endWar(wr, "destroyed")
	}
}

// treats says whether a people can be brought to terms at all: bends
// the knee, asks to be ruled, yields, sues. A thing that eats cannot.
func (c *Civ) treats() bool { return !c.Species.Profile().NoTerms }

// noTerms says whether a war can end only by exhaustion: a side cannot
// treat, and no offer of terms from the other is heard.
func (w *World) noTerms(wr *War) bool {
	return !w.Civs[wr.Sides[0]].treats() || !w.Civs[wr.Sides[1]].treats()
}

// innate is what a people has from birth by its nature: the nodes its
// profile names, known, their filters counted as faced, since there is
// nothing to fall from in what one is.
func (w *World) innate(c *Civ) {
	for _, k := range c.Species.Profile().Innate {
		if n := tech.Get(k); n != nil && !c.Known[k] {
			c.Known[k] = true
			if n.Filter != "" {
				c.Faced[n.Filter] = true
			}
		}
	}
}

// ariseAt is a made people arising at a star that may be someone's:
// whoever held it is stripped from it first (and ends, if it was their
// last), and the new people holds it. Never at Sol: the nearest star
// instead. The species is made, with the story's fixed parts and the
// rest rolled; sp.Made says how.
func (w *World) ariseAt(star int, sp *species.Species, made string) *Civ {
	if star == w.G.Sol {
		for _, s := range w.G.Near(star, 40) {
			if s != w.G.Sol {
				star = s
				break
			}
		}
	}
	if o := w.Owner[star]; o >= 0 {
		h := w.Civs[o]
		w.loseSystem(h, star, "stripped world", sprintf("were consumed by what woke at %s", w.star(star)))
		if w.Owner[star] >= 0 {
			return nil // a rider or a fleeing people holds it still
		}
	}
	sp.Made = made
	return w.spawnCiv(star, sp, -1, "")
}

// replicatorAt is a replicator people arising at a star: of the body
// given, dormant or awake. Dormant, it sleeps until disturbed.
func (w *World) replicatorAt(star int, sub species.Substrate, made string, dormant bool) *Civ {
	sp := species.GenerateWith(w.R, w.G.Stars[star].Mult, w.G.Sys[star].Arch, sub, species.Replicator)
	c := w.ariseAt(star, sp, made)
	if c == nil {
		return nil
	}
	c.Origin = made
	if dormant {
		w.sleep(c)
	}
	return c
}
